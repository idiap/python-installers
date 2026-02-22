// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/poetry/fakes"

	"github.com/paketo-buildpacks/python-installers/pkg/build"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		parsePythonVersion *fakes.PoetryPyProjectParser

		workingDir string

		detect packit.DetectFunc
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		parsePythonVersion = &fakes.PoetryPyProjectParser{}
		parsePythonVersion.ParsePythonVersionCall.Returns.String = "1.2.3"
		parsePythonVersion.IsPoetryProjectCall.Returns.Bool = true

		Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte(""), 0755)).To(Succeed())

		detect = poetry.Detect(parsePythonVersion)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	it("returns a plan that provides poetry", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(result).To(Equal(packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: poetry.Pip},
					{Name: poetry.PoetryDependency},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: poetry.CPython,
						Metadata: build.BuildPlanMetadata{
							Build:         true,
							Version:       "1.2.3",
							VersionSource: "pyproject.toml",
						},
					},
					{
						Name: poetry.Pip,
						Metadata: build.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			},
		}))
	})

	context("when the BP_POETRY_VERSION is set", func() {
		it.Before(func() {
			t.Setenv(poetry.EnvVersion, "some-version")
		})

		it("returns a plan that requires that version of poetry", func() {
			result, err := detect(packit.DetectContext{
				WorkingDir: workingDir,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(parsePythonVersion.ParsePythonVersionCall.Receives.String).To(Equal(filepath.Join(workingDir, "pyproject.toml")))
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: poetry.Pip},
						{Name: poetry.PoetryDependency},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: poetry.CPython,
							Metadata: build.BuildPlanMetadata{
								Build:         true,
								Version:       "1.2.3",
								VersionSource: "pyproject.toml",
							},
						},
						{
							Name: poetry.Pip,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: poetry.PoetryDependency,
							Metadata: build.BuildPlanMetadata{
								VersionSource: poetry.EnvVersion,
								Version:       "some-version",
							},
						},
					},
				},
			}))
		})

		context("when pyproject.toml is not present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "pyproject.toml"))).To(Succeed())
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(packit.Fail.WithMessage("pyproject.toml is not present")))
			})
		})

		context("when pyproject.toml is not for poetry", func() {
			it.Before(func() {
				parsePythonVersion.IsPoetryProjectCall.Returns.Bool = false
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(packit.Fail.WithMessage("this is not a poetry project")))
			})
		})

		context("when no python version is returned from the parser", func() {
			it.Before(func() {
				parsePythonVersion.ParsePythonVersionCall.Returns.String = ""
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(packit.Fail.WithMessage("pyproject.toml must include [tool.poetry.dependencies.python], see https://python-poetry.org/docs/pyproject/#dependencies-and-dev-dependencies")))
			})
		})

		context("error handling", func() {
			context("when there is an error determining if the pyproject.toml file exists", func() {
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

			context("when the pyproject parser returns an error", func() {
				it.Before(func() {
					parsePythonVersion.ParsePythonVersionCall.Returns.Error = errors.New("some-error")
				})

				it("returns the error", func() {
					_, err := detect(packit.DetectContext{
						WorkingDir: workingDir,
					})
					Expect(err).To(Equal(errors.New("some-error")))
				})
			})
		})
	})
}
