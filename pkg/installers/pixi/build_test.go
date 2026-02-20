// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixi_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
	sbomfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/common/sbom/fakes"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/pixi"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/pixi/fakes"

	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir string
		cnbDir    string

		buffer *bytes.Buffer

		dependencyManager *fakes.DependencyManager
		installProcess    *fakes.InstallProcess
		sbomGenerator     *sbomfakes.SBOMGenerator

		build        packit.BuildFunc
		buildContext packit.BuildContext
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		dependencyManager = &fakes.DependencyManager{}
		dependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "pixi",
			Name:     "pixi-dependency-name",
			Checksum: "pixi-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "pixi-dependency-uri",
			Version:  "pixi-dependency-version",
		}

		// Legacy SBOM
		dependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "pixi",
				Metadata: paketosbom.BOMMetadata{
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "pixi-dependency-sha",
					},
					URI:     "pixi-dependency-uri",
					Version: "pixi-dependency-version",
				},
			},
		}

		// Syft SBOM
		sbomGenerator = &sbomfakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		installProcess = &fakes.InstallProcess{}

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		build = pixi.Build(
			pixi.PixiBuildParameters{
				DependencyManager: dependencyManager,
				InstallProcess:    installProcess,
			},
			pythoninstallers.CommonBuildParameters{
				SbomGenerator: sbomGenerator,
				Clock:         chronos.DefaultClock,
				Logger:        logger,
			},
		)
		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			CNBPath: cnbDir,
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{Name: pixi.Pixi},
				},
			},
			Platform: packit.Platform{Path: "some-platform-path"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
	})

	it("returns a result that installs pixi", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		layer := result.Layers[0]

		Expect(layer.Name).To(Equal("pixi"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "pixi")))

		Expect(layer.SharedEnv).To(BeEmpty())
		Expect(layer.BuildEnv).To(BeEmpty())
		Expect(layer.LaunchEnv).To(BeEmpty())
		Expect(layer.ProcessLaunchEnv).To(BeEmpty())

		Expect(layer.Build).To(BeFalse())
		Expect(layer.Launch).To(BeFalse())
		Expect(layer.Cache).To(BeFalse())

		Expect(layer.Metadata).To(HaveLen(1))
		Expect(layer.Metadata["dependency-sha"]).To(Equal("pixi-dependency-sha"))

		Expect(layer.SBOM.Formats()).To(HaveLen(2))
		var actualExtensions []string
		for _, format := range layer.SBOM.Formats() {
			actualExtensions = append(actualExtensions, format.Extension)
		}
		Expect(actualExtensions).To(ConsistOf("cdx.json", "spdx.json"))

		Expect(dependencyManager.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbDir, "buildpack.toml")))
		Expect(dependencyManager.ResolveCall.Receives.Id).To(Equal("pixi"))
		Expect(dependencyManager.ResolveCall.Receives.Version).To(Equal(""))
		Expect(dependencyManager.ResolveCall.Receives.Stack).To(Equal("some-stack"))

		Expect(dependencyManager.GenerateBillOfMaterialsCall.Receives.Dependencies).To(Equal([]postal.Dependency{
			{
				ID:       "pixi",
				Name:     "pixi-dependency-name",
				Checksum: "pixi-dependency-sha",
				Stacks:   []string{"some-stack"},
				URI:      "pixi-dependency-uri",
				Version:  "pixi-dependency-version",
			},
		}))

		Expect(dependencyManager.DeliverCall.Receives.Dependency).To(Equal(
			postal.Dependency{
				ID:       "pixi",
				Name:     "pixi-dependency-name",
				Checksum: "pixi-dependency-sha",
				Stacks:   []string{"some-stack"},
				URI:      "pixi-dependency-uri",
				Version:  "pixi-dependency-version",
			}))
		Expect(dependencyManager.DeliverCall.Receives.CnbPath).To(Equal(cnbDir))
		Expect(dependencyManager.DeliverCall.Receives.DestinationPath).To(Equal(filepath.Join(layersDir, "pixi-temp-layer")))
		Expect(dependencyManager.DeliverCall.Receives.PlatformPath).To(Equal("some-platform-path"))

		Expect(sbomGenerator.GenerateFromDependencyCall.Receives.Dir).To(Equal(filepath.Join(layersDir, "pixi")))

		Expect(buffer.String()).To(ContainSubstring("Some Buildpack some-version"))
		Expect(buffer.String()).To(ContainSubstring("Executing build process"))
		Expect(buffer.String()).To(ContainSubstring("Installing pixi"))
	})

	context("when the pixi layer is required at build and launch", func() {
		it.Before(func() {
			buildContext.Plan.Entries[0].Metadata = make(map[string]interface{})
			buildContext.Plan.Entries[0].Metadata["launch"] = true
			buildContext.Plan.Entries[0].Metadata["build"] = true
		})

		it("returns a layer with build and launch set true and the BOM is set for build and launch", func() {
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Layers).To(HaveLen(1))
			layer := result.Layers[0]

			Expect(layer.Name).To(Equal("pixi"))

			Expect(layer.Build).To(BeTrue())
			Expect(layer.Launch).To(BeTrue())
			Expect(layer.Cache).To(BeTrue())

			Expect(result.Build.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "pixi",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "pixi-dependency-sha",
							},
							URI:     "pixi-dependency-uri",
							Version: "pixi-dependency-version",
						},
					},
				},
			))

			Expect(result.Launch.BOM).To(Equal(
				[]packit.BOMEntry{
					{
						Name: "pixi",
						Metadata: paketosbom.BOMMetadata{
							Checksum: paketosbom.BOMChecksum{
								Algorithm: paketosbom.SHA256,
								Hash:      "pixi-dependency-sha",
							},
							URI:     "pixi-dependency-uri",
							Version: "pixi-dependency-version",
						},
					},
				},
			))
		})
	})

	context("failure cases", func() {
		context("when the dependency manager resolution fails", func() {
			it.Before(func() {
				dependencyManager.ResolveCall.Returns.Error = errors.New("resolve call failed")
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError("resolve call failed"))
			})
		})

		context("when the layer dir cannot be accessed", func() {
			it.Before(func() {
				Expect(os.Chmod(layersDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layersDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the layer dir cannot be reset", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(layersDir, "pixi", "bin"), os.ModePerm)).To(Succeed())
				Expect(os.Chmod(filepath.Join(layersDir, "pixi"), 0500)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(layersDir, "pixi"), os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when the dependency manager delivery fails", func() {
			it.Before(func() {
				dependencyManager.DeliverCall.Returns.Error = errors.New("deliver call failed")
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError("deliver call failed"))
			})
		})

		context("when generating the SBOM returns an error", func() {
			it.Before(func() {
				buildContext.BuildpackInfo.SBOMFormats = []string{"random-format"}
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(`unsupported SBOM format: 'random-format'`))
			})
		})

		context("when formatting the SBOM returns an error", func() {
			it.Before(func() {
				sbomGenerator.GenerateFromDependencyCall.Returns.Error = errors.New("failed to generate SBOM")
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("failed to generate SBOM")))
			})
		})

		context("when the install process returns an error", func() {
			it.Before(func() {
				installProcess.ExecuteCall.Returns.Error = errors.New("failed to copy files")
			})

			it("returns an error", func() {
				_, err := build(buildContext)

				Expect(err).To(MatchError(ContainSubstring("failed to copy files")))
			})
		})
	})
}
