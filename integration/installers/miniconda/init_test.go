// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	integration_helpers "github.com/paketo-buildpacks/python-installers/integration"
)

var settings struct {
	Buildpacks struct {
		BuildPlan struct {
			Online string
		}
		Miniconda struct {
			Online  string
			Offline string
		}
	}
	Buildpack struct {
		ID   string
		Name string
	}
	Config struct {
		BuildPlan string `json:"build-plan"`
	}
}

func TestMinicondaIntegration(t *testing.T) {
	// Do not truncate Gomega matcher output
	// The buildpack output text can be large and we often want to see all of it.
	format.MaxLength = 0

	Expect := NewWithT(t).Expect

	file, err := os.Open("integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&settings.Config)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	file, err = os.Open("./../../../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())

	_, err = toml.NewDecoder(file).Decode(&settings)
	Expect(err).NotTo(HaveOccurred())
	Expect(file.Close()).To(Succeed())

	root, err := filepath.Abs("./../../..")
	Expect(err).ToNot(HaveOccurred())

	buildpackStore := integration_helpers.NewBuildpackStore("Miniconda")

	settings.Buildpacks.Miniconda.Online, err = buildpackStore.Get.
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.Miniconda.Offline, err = buildpackStore.Get.
		WithVersion("1.2.3").
		WithOfflineDependencies().
		Execute(root)
	Expect(err).ToNot(HaveOccurred())

	settings.Buildpacks.BuildPlan.Online, err = buildpackStore.Get.
		Execute(settings.Config.BuildPlan)
	Expect(err).ToNot(HaveOccurred())

	SetDefaultEventuallyTimeout(30 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Default", testDefault)
	suite("Logging", testLogging)
	suite("LayerReuse", testReusingLayerRebuild)
	suite("TestOffline", testOffline)
	suite.Run(t)
}
