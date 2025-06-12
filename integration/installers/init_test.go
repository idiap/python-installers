// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	integration_helpers "github.com/paketo-buildpacks/python-installers/integration"
)

var (
	builder occam.Builder

	buildpackInfo integration_helpers.BuildpackInfo

	settings integration_helpers.TestSettings
)

func TestIntegration(t *testing.T) {
	// Do not truncate Gomega matcher output
	// The buildpack output text can be large and we often want to see all of it.
	format.MaxLength = 0

	Expect := NewWithT(t).Expect

	root, err := filepath.Abs("./../../")
	Expect(err).NotTo(HaveOccurred())

	file, err := os.Open(filepath.Join(root, "/integration.json"))
	Expect(err).NotTo(HaveOccurred())

	err = json.NewDecoder(file).Decode(&settings.Config)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open(filepath.Join(root, "buildpack.toml"))
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&buildpackInfo)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.CPython.Online, err = buildpackStore.Get.
		Execute(settings.Config.CPython)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.CPython.Offline, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(settings.Config.CPython)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Pip.Online, err = buildpackStore.Get.
		Execute(settings.Config.Pip)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.Pip.Offline, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(settings.Config.Pip)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.PythonInstallers.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	settings.Buildpacks.PythonInstallers.Offline, err = buildpackStore.Get.
		WithVersion("1.2.3").
		WithOfflineDependencies().
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	pack := occam.NewPack().WithVerbose()
	builder, err = pack.Builder.Inspect.Execute()
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())

	// miniconda
	suite("Miniconda Default", minicondaTestDefault)
	suite("Miniconda Logging", minicondaTestLogging)
	suite("Miniconda LayerReuse", minicondaTestReusingLayerRebuild)
	suite("Miniconda TestOffline", minicondaTestOffline)

	// pip
	suite("Pip Default", pipTestDefault, spec.Parallel())
	suite("Pip LayerReuse", pipTestLayerReuse, spec.Parallel())
	if strings.Contains(builder.LocalInfo.Stack.ID, "jammy") || strings.Contains(builder.LocalInfo.Stack.ID, "bionic") {
		suite("Pip Offline", pipTestOffline, spec.Parallel())
	}

	// pipenv
	suite("Pipenv Default", pipenvTestDefault, spec.Parallel())
	suite("Pipenv LayerReuse", pipenvTestLayerReuse, spec.Parallel())
	suite("Pipenv Version", pipenvTestVersions, spec.Parallel())

	// poetry
	suite("Poetry Default", poetryTestDefault, spec.Parallel())
	suite("Poetry LayerReuse", poetryTestLayerReuse, spec.Parallel())
	suite("Poetry Versions", poetryTestVersions, spec.Parallel())

	suite.Run(t)
}
