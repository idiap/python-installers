// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	pythoninstallers "github.com/paketo-buildpacks/python-installers"
	common "github.com/paketo-buildpacks/python-installers/pkg/installers/common"

	miniconda "github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda"
	pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
	poetryfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry/fakes"
	uv "github.com/paketo-buildpacks/python-installers/pkg/installers/uv"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		buffer     *bytes.Buffer

		parsePythonVersion *poetryfakes.PyProjectPythonVersionParser

		detect packit.DetectFunc

		plans []packit.BuildPlan
	)

	it.Before(func() {
		workingDir = t.TempDir()

		Expect(os.WriteFile(filepath.Join(workingDir, "x.py"), []byte{}, os.ModePerm)).To(Succeed())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		parsePythonVersion = &poetryfakes.PyProjectPythonVersionParser{}
		parsePythonVersion.ParsePythonVersionCall.Returns.String = "1.2.3"

		detect = pythoninstallers.Detect(logger, parsePythonVersion)

		plans = append(plans, packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: pip.Pip},
			},
			Requires: []packit.BuildPlanRequirement{
				{
					Name: pip.CPython,
					Metadata: common.BuildPlanMetadata{
						Build: true,
					},
				},
			},
		},
		)

		plans = append(plans, packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: miniconda.Conda},
			},
		},
		)

		plans = append(plans, packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: pipenv.Pip},
				{Name: pipenv.Pipenv},
			},
			Requires: []packit.BuildPlanRequirement{
				{
					Name: pipenv.CPython,
					Metadata: common.BuildPlanMetadata{
						Build:  true,
						Launch: false,
					},
				},
				{
					Name: pipenv.Pip,
					Metadata: common.BuildPlanMetadata{
						Build:  true,
						Launch: false,
					},
				},
			},
		},
		)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("detection phase", func() {
		context("without pyproject.toml", func() {
			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(pythoninstallers.Or(plans...)))
			})
		})

		context("with pyproject.toml", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte{}, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte(""), 0755)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				withPoetry := append(plans,
					packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: poetry.Pip},
							{Name: poetry.PoetryDependency},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: poetry.CPython,
								Metadata: common.BuildPlanMetadata{
									Build:         true,
									Version:       "1.2.3",
									VersionSource: "pyproject.toml",
								},
							},
							{
								Name: poetry.Pip,
								Metadata: common.BuildPlanMetadata{
									Build: true,
								},
							},
						},
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(pythoninstallers.Or(withPoetry...)))
			})
		})

		context("with uv.lock", func() {
			it.Before(func() {
				Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), []byte{}, os.ModePerm)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, uv.LockfileName), []byte(`requires-python = "3.13.0"`), 0755)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				withUv := append(plans,
					packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: uv.Uv},
						},
					},
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(pythoninstallers.Or(withUv...)))
			})
		})
	})
}
