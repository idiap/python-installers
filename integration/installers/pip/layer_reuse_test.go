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
)

func testLayerReuse(t *testing.T, context spec.G, it spec.S) {
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

		source, err = occam.Source(filepath.Join("testdata", "default_app"))
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

	context("when the app is rebuilt and the same pip version is required", func() {
		it("reuses the cached pip layer", func() {
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
					settings.Buildpacks.CPython.Online,
					settings.Buildpacks.Pip.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}
			firstContainer, err = docker.Container.Run.
				WithCommand("pip --version").
				Execute(firstImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			secondImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.CPython.Online,
					settings.Buildpacks.Pip.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[secondImage.ID] = struct{}{}

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Pip version",
				"    Candidate version sources (in priority order):",
				"      <unknown> -> \"\"",
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected Pip version \(using <unknown>\): \d+\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				fmt.Sprintf("  Reusing cached layer /layers/%s/pip", strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_")),
			))

			secondContainer, err = docker.Container.Run.
				WithCommand("pip --version").
				Execute(secondImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(secondContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(MatchRegexp(fmt.Sprintf(`pip \d+\.\d+(\.\d+)? from /layers/%s/pip/lib/python\d+.\d+/site-packages/pip`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))))

			Expect(secondImage.Buildpacks[1].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[1].Layers["pip"].SHA).To(Equal(firstImage.Buildpacks[1].Layers["pip"].SHA))
		})
	})

	context("when the app is rebuilt and a different pip version is required", func() {
		it("rebuilds", func() {
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
					settings.Buildpacks.CPython.Online,
					settings.Buildpacks.Pip.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_PIP_VERSION": buildpackInfo.Metadata.Dependencies[0].Version}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[firstImage.ID] = struct{}{}
			firstContainer, err = docker.Container.Run.
				WithCommand("pip --version").
				Execute(firstImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[firstContainer.ID] = struct{}{}

			secondImage, logs, err = pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.CPython.Online,
					settings.Buildpacks.Pip.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_PIP_VERSION": buildpackInfo.Metadata.Dependencies[1].Version}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), logs.String)

			imageIDs[secondImage.ID] = struct{}{}

			Expect(logs).To(ContainLines(
				MatchRegexp(fmt.Sprintf(`%s \d+\.\d+\.\d+`, buildpackInfo.Buildpack.Name)),
				"  Resolving Pip version",
				"    Candidate version sources (in priority order):",
				MatchRegexp(`      BP_PIP_VERSION -> "\d+\.\d+\.\d+"`),
				"      <unknown>      -> \"\"",
			))
			Expect(logs).To(ContainLines(
				MatchRegexp(`    Selected Pip version \(using BP_PIP_VERSION\): \d+\.\d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				"  Executing build process",
				MatchRegexp(`    Installing Pip \d+\.\d+\.\d+`),
				MatchRegexp(`      Completed in \d+\.\d+`),
			))
			Expect(logs).To(ContainLines(
				"  Configuring build environment",
				MatchRegexp(fmt.Sprintf(`    PYTHONPATH -> "\/layers\/%s\/pip\/lib\/python\d+\.\d+\/site-packages:\$PYTHONPATH"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
				"",
				"  Configuring launch environment",
				MatchRegexp(fmt.Sprintf(`    PYTHONPATH -> "\/layers\/%s\/pip\/lib\/python\d+\.\d+\/site-packages:\$PYTHONPATH"`, strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))),
			))

			secondContainer, err = docker.Container.Run.
				WithCommand("pip --version").
				Execute(secondImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containerIDs[secondContainer.ID] = struct{}{}

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(secondContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(MatchRegexp(fmt.Sprintf(`pip %s from /layers/%s/pip/lib/python\d+.\d+/site-packages/pip`, strings.Replace(buildpackInfo.Metadata.Dependencies[1].Version, ".0", `(\.\d+)?`, 1), strings.ReplaceAll(buildpackInfo.Buildpack.ID, "/", "_"))))

			Expect(secondImage.Buildpacks[1].Key).To(Equal(buildpackInfo.Buildpack.ID))
			Expect(secondImage.Buildpacks[1].Layers["pip"].SHA).ToNot(Equal(firstImage.Buildpacks[1].Layers["pip"].SHA))
		})
	})
}
