// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uv_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/uv"

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

		Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), []byte(`requires-python = "==3.13.0"`), 0755)).To(Succeed())

		detect = uv.Detect()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when the BP_UV_VERSION is NOT set", func() {
		it("returns a plan that provides uv", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: uv.Uv},
					},
				},
			}))
		})
	})

	context("when the BP_UV_VERSION is set", func() {
		it.Before(func() {
			t.Setenv("BP_UV_VERSION", "some-version")
		})

		it("returns a plan that requires that version of uv", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: uv.Uv},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: uv.Uv,
							Metadata: pythoninstallers.BuildPlanMetadata{
								VersionSource: "BP_UV_VERSION",
								Version:       "some-version",
							},
						},
					},
				},
			}))
		})
	})

	context("when uv.lock is not present", func() {
		it.Before(func() {
			Expect(os.RemoveAll(filepath.Join(workingDir, uv.LockfileName))).To(Succeed())
		})

		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail.WithMessage("uv.lock is not present")))
		})
	})

	context("when no python version is returned from the parser", func() {
		it.Before(func() {
			var err error
			workingDir, err = os.MkdirTemp("", "working-dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), []byte(""), 0755)).To(Succeed())
		})

		it.After(func() {
			Expect(os.RemoveAll(workingDir)).To(Succeed())
		})

		it("fails detection", func() {
			_, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).To(MatchError(packit.Fail.WithMessage("uv.lock must include requires-python")))
		})
	})

	context("error handling", func() {
		context("when there is an error determining if the uv.lock file exists", func() {
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

		context("when the uv lock file parser returns an error", func() {
			it.Before(func() {
				var err error
				workingDir, err = os.MkdirTemp("", "working-dir")
				Expect(err).NotTo(HaveOccurred())

				Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), []byte("<test>error</test>"), 0755)).To(Succeed())
			})

			it.After(func() {
				Expect(os.RemoveAll(workingDir)).To(Succeed())
			})

			it("returns the error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("toml: line 1: expected '.' or '=', but got '<' instead")))
			})
		})
	})
}
