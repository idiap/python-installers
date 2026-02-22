// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"

	integration_helpers "github.com/paketo-buildpacks/python-installers/integration"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/pixi"
)

func pixiTestLayerReuse(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker

		imageIDs     map[string]struct{}
		containerIDs map[string]struct{}

		name   string
		source string
	)

	it.Before(func() {
		var err error
		name, err = occam.RandomName()
		Expect(err).NotTo(HaveOccurred())

		pack = occam.NewPack()
		docker = occam.NewDocker()

		imageIDs = map[string]struct{}{}
		containerIDs = map[string]struct{}{}

		source, err = occam.Source(filepath.Join("testdata", "pixi", "pixi_app"))
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		for id := range containerIDs {
			Expect(docker.Container.Remove.Execute(id)).To(Succeed())
		}

		for id := range imageIDs {
			Expect(docker.Image.Remove.Execute(id)).To(Succeed())
		}

		Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())

		Expect(os.RemoveAll(source)).To(Succeed())
	})

	context("when the app is rebuilt and the same pixi version is required", func() {
		it("reuses the cached pixi layer", func() {
			var (
				err  error
				logs fmt.Stringer

				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			firstImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}
			firstContainer, err = docker.Container.Run.
				WithCommand("pixi --version").
				Execute(firstImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			secondImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[secondImage.ID] = struct{}{}

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving pixi version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected pixi version \(using <unknown>\): \d+\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				fmt.Sprintf("  Reusing cached layer /layers/%s/pixi", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))

			secondContainer, err = docker.Container.Run.
				WithCommand("pixi --version").
				Execute(secondImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(secondContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(MatchRegexp(`pixi \d+\.\d+(\.\d+)?.*`))

			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers["pixi"].SHA).To(Equal(firstImage.Buildpacks[0].Layers["pixi"].SHA))
		})
	})

	context("when the app is rebuilt and a different pixi version is required", func() {
		it("rebuilds", func() {
			var (
				err  error
				logs fmt.Stringer

				firstImage  occam.Image
				secondImage occam.Image

				firstContainer  occam.Container
				secondContainer occam.Container
			)

			dependencies := integration_helpers.DependenciesForId(buildpackInfo.Metadata.Dependencies, "pixi")

			firstImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{pixi.EnvVersion: dependencies[0].Version}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}
			firstContainer, err = docker.Container.Run.
				WithCommand("pixi --version").
				Execute(firstImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			secondImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{pixi.EnvVersion: dependencies[2].Version}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[secondImage.ID] = struct{}{}

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving pixi version",
				"    Candidate version sources (in priority order):",
				MatchRegexp(`      BP_PIXI_VERSION -> "\d+\.\d+\.\d+"`),
				"      <unknown>       -> \"\"",
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected pixi version \(using BP_PIXI_VERSION\): \d+\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing pixi \d+\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+(\.?\d+)*`),
			))

			secondContainer, err = docker.Container.Run.
				WithCommand("pixi --version").
				Execute(secondImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(secondContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(MatchRegexp(`pixi \d+\.\d+(\.\d+)?.*`))

			Expect(secondImage.Buildpacks[0].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[0].Layers["pixi"].SHA).ToNot(Equal(firstImage.Buildpacks[0].Layers["pixi"].SHA))
		})
	})
}
