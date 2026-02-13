// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"

	integration_helpers "github.com/paketo-buildpacks/python-installers/integration"
)

func minicondaTestVersions(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			name   string
			source string

			containersMap map[string]interface{}
			imagesMap     map[string]interface{}
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			containersMap = map[string]interface{}{}
			imagesMap = map[string]interface{}{}
		})

		it.After(func() {
			for containerID := range containersMap {
				Expect(docker.Container.Remove.Execute(containerID)).To(Succeed())
			}
			for imageID := range imagesMap {
				Expect(docker.Image.Remove.Execute(imageID)).To(Succeed())
			}
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("builds and runs successfully with both provided dependency versions", func() {
			var err error

			source, err = occam.Source(filepath.Join("testdata", "conda", "miniconda_app"))
			Expect(err).NotTo(HaveOccurred())

			dependencies := integration_helpers.DependenciesForId(buildpackInfo.Metadata.Dependencies, "miniconda3")

			firstMinicondaVersion := dependencies[0].Version
			secondMinicondaVersion := dependencies[2].Version

			Expect(firstMinicondaVersion).NotTo(Equal(secondMinicondaVersion))

			firstImage, firstLogs, err := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_MINICONDA_VERSION": firstMinicondaVersion}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), firstLogs.String)

			imagesMap[firstImage.ID] = nil

			Expect(firstLogs).To(ContainLines(
				ContainSubstring(fmt.Sprintf(`Selected Miniconda.sh version (using BP_MINICONDA_VERSION): %s`, firstMinicondaVersion)),
			))

			firstContainer, err := docker.Container.Run.
				WithCommand("conda --version").
				Execute(firstImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containersMap[firstContainer.ID] = nil

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(firstContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring(fmt.Sprintf(`conda %s`, firstMinicondaVersion)))

			secondImage, secondLogs, err := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					settings.Buildpacks.BuildPlan.Online,
				).
				WithEnv(map[string]string{"BP_MINICONDA_VERSION": secondMinicondaVersion}).
				Execute(name, source)
			Expect(err).ToNot(HaveOccurred(), secondLogs.String)

			imagesMap[secondImage.ID] = nil

			Expect(secondLogs).To(ContainLines(
				ContainSubstring(fmt.Sprintf(`Selected Miniconda.sh version (using BP_MINICONDA_VERSION): %s`, secondMinicondaVersion)),
			))

			secondContainer, err := docker.Container.Run.
				WithCommand("conda --version").
				Execute(secondImage.ID)
			Expect(err).ToNot(HaveOccurred())

			containersMap[secondContainer.ID] = nil

			Eventually(func() string {
				cLogs, err := docker.Container.Logs.Execute(secondContainer.ID)
				Expect(err).NotTo(HaveOccurred())
				return cLogs.String()
			}).Should(ContainSubstring(fmt.Sprintf(`conda %s`, secondMinicondaVersion)))
		})
	})
}
