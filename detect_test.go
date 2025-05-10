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
	conda "github.com/paketo-buildpacks/python-installers/pkg/installers/conda"
	pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		buffer     *bytes.Buffer

		detect packit.DetectFunc
	)

	it.Before(func() {
		workingDir = t.TempDir()

		Expect(os.WriteFile(filepath.Join(workingDir, "x.py"), []byte{}, os.ModePerm)).To(Succeed())

		buffer = bytes.NewBuffer(nil)
		logger := scribe.NewEmitter(buffer)

		detect = pythoninstallers.Detect(logger)
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("detection phase", func() {
		context("When only an environment.yml file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "environment.yml"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: conda.CondaEnvPlanEntry,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: conda.CondaPlanEntry,
							Metadata: map[string]interface{}{
								"build": true,
							},
						},
					},
				}))

			})
		})

		context("When only a package-list.txt file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "package-list.txt"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: conda.CondaEnvPlanEntry,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: conda.CondaPlanEntry,
							Metadata: map[string]interface{}{
								"build": true,
							},
						},
					},
				}))
			})
		})

		context("When only a requirements.txt file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "requirements.txt"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: pip.SitePackages,
						},
						{
							Name: pip.Manager,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: pip.CPython,
							Metadata: pip.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pip.Pip,
							Metadata: pip.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pip.Manager,
							Metadata: pip.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When only a Pipfile file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "Pipfile"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: pipenv.SitePackages,
						},
						{
							Name: pipenv.Manager,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: pipenv.CPython,
							Metadata: pipenv.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pipenv.Pipenv,
							Metadata: pipenv.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pipenv.Manager,
							Metadata: pipenv.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When only a pyproject.toml file is present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
				Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte{}, os.ModePerm)).To(Succeed())
			})

			it("passes detection", func() {
				result, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{
							Name: poetry.PoetryVenv,
						},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: poetry.CPython,
							Metadata: poetry.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: poetry.Poetry,
							Metadata: poetry.BuildPlanMetadata{
								Build: true,
							},
						},
					},
				}))
			})
		})

		context("When no python related files are present", func() {
			it.Before(func() {
				Expect(os.RemoveAll(filepath.Join(workingDir, "x.py"))).To(Succeed())
			})

			it("fails detection", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})
				Expect(err).To(MatchError(ContainSubstring("No python packager manager related files found")))
			})
		})
	})
}
