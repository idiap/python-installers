// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixi_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/python-installers/pkg/build"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/pixi"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string

		detect packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, pixi.LockfileName), []byte(``), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, pixi.ProjectFilename), []byte(``), 0755)).To(Succeed())

		detect = pixi.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when the BP_PIXI_VERSION is NOT set", func() {
		it("returns a plan that provides pixi", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: pixi.Pixi},
					},
				},
			}))
		})
	})

	context("when the BP_PIXI_VERSION is set", func() {
		it.Before(func() {
			t.Setenv(pixi.EnvVersion, "some-version")
		})

		it("returns a plan that requires that version of pixi", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: pixi.Pixi},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: pixi.Pixi,
							Metadata: build.BuildPlanMetadata{
								VersionSource: pixi.EnvVersion,
								Version:       "some-version",
							},
						},
					},
				},
			}))
		})
	})

	context("when files are missing", func() {
		context("when pixi.lock is not present", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, pixi.ProjectFilename), []byte(``), 0755)).To(Succeed())
				Expect(os.RemoveAll(filepath.Join(workingDir, pixi.LockfileName))).To(Succeed())
			})

			it("returns a plan that provides pixi", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(packit.DetectResult{
					Plan: packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: pixi.Pixi},
						},
					},
				}))
			})
		})

		context("when pixi.toml is not present", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, pixi.LockfileName), []byte(``), 0755)).To(Succeed())
				Expect(os.RemoveAll(filepath.Join(workingDir, pixi.ProjectFilename))).To(Succeed())
			})

			it("returns a plan that provides pixi", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(packit.DetectResult{
					Plan: packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: pixi.Pixi},
						},
					},
				}))
			})
		})

		context("error case case", func() {
			context("when both pixi.lock and pixi.toml are missing", func() {
				it.Before(func() {
					Expect(os.RemoveAll(filepath.Join(workingDir, pixi.ProjectFilename))).To(Succeed())
					Expect(os.RemoveAll(filepath.Join(workingDir, pixi.LockfileName))).To(Succeed())
				})

				it("fails to build", func() {
					_, err := detect(packit.DetectContext{
						WorkingDir: workingDir,
					})
					Expect(err).To(MatchError(packit.Fail.WithMessage("neither pixi.lock nor pixi.toml are present")))
				})
			})
		})

	})

	context("error handling", func() {
		context("when there is an error determining if the pixi.lock and pixi.toml files exist", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns the error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
