// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers_test

import (
	"bytes"
	// "os"
	// "path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	pythoninstallers "github.com/paketo-buildpacks/python-installers"
	pkgcommon "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
	conda "github.com/paketo-buildpacks/python-installers/pkg/installers/conda"
	condafakes "github.com/paketo-buildpacks/python-installers/pkg/installers/conda/fakes"
	pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	pipfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/pip/fakes"
	pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	pipenvfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv/fakes"
	poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
	poetryfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry/fakes"

	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

		buffer       *bytes.Buffer
		logger       scribe.Emitter
		build        packit.BuildFunc
		buildContext packit.BuildContext

		// common
		sbomGenerator *pipfakes.SBOMGenerator

		// conda
		runner *condafakes.Runner

		// pip
		pipProcess      *pipfakes.InstallProcess
		pipSitePackagesProcess *pipfakes.SitePackagesProcess

		// pipenv
		pipenvProcess      *pipenvfakes.InstallProcess
		pipenvSitePackagesProcess *pipenvfakes.SitePackagesProcess
		pipenvVenvDirLocator      *pipenvfakes.VenvDirLocator

		// poetry
		poetryEntryResolver     *poetryfakes.EntryResolver
		poetryProcess    *poetryfakes.InstallProcess
		poetryPythonPathProcess *poetryfakes.PythonPathLookupProcess

		buildParameters pkgcommon.CommonBuildParameters

		plans []packit.BuildpackPlan
	)

	it.Before(func() {
		layersDir = t.TempDir()
		workingDir = t.TempDir()
		cnbDir = t.TempDir()

		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)

		sbomGenerator = &pipfakes.SBOMGenerator{}
		sbomGenerator.GenerateCall.Returns.SBOM = sbom.SBOM{}

		// conda
		runner = &condafakes.Runner{}
		runner.ShouldRunCall.Returns.Bool = true
		runner.ShouldRunCall.Returns.String = "some-sha"

		// pip
		pipProcess = &pipfakes.InstallProcess{}
		pipSitePackagesProcess = &pipfakes.SitePackagesProcess{}
		pipSitePackagesProcess.ExecuteCall.Returns.SitePackagesPath = "some-site-packages-path"

		// pipenv
		pipenvProcess = &pipenvfakes.InstallProcess{}
		pipenvSitePackagesProcess = &pipenvfakes.SitePackagesProcess{}
		pipenvSitePackagesProcess.ExecuteCall.Returns.SitePackagesPath = "some-site-packages-path"
		pipenvVenvDirLocator = &pipenvfakes.VenvDirLocator{}
		pipenvVenvDirLocator.LocateVenvDirCall.Returns.VenvDir = "some-venv-dir"

		// poetry
		poetryEntryResolver = &poetryfakes.EntryResolver{}
		poetryProcess = &poetryfakes.InstallProcess{}
		poetryProcess.ExecuteCall.Returns.String = "some-venv-dir"
		poetryPythonPathProcess = &poetryfakes.PythonPathLookupProcess{}
		poetryPythonPathProcess.ExecuteCall.Returns.String = "some-python-path"

		buildParameters = pkgcommon.CommonBuildParameters{
			SbomGenerator: pkgcommon.Generator{},
			Clock:         chronos.DefaultClock,
			Logger:        logger,
		}

		packagerParameters := map[string]pythoninstallers.PackagerParameters{
			conda.CondaEnvPlanEntry: conda.CondaBuildParameters{
				Runner: runner,
			},
			pip.Manager: pip.PipBuildParameters{
				InstallProcess:      pipProcess,
				SitePackagesProcess: pipSitePackagesProcess,
			},
			pipenv.Manager: pipenv.PipEnvBuildParameters{
				InstallProcess: pipenvProcess,
				SiteProcess:    pipenvSitePackagesProcess,
				VenvDirLocator: pipenvVenvDirLocator,
			},
			poetry.PoetryVenv: poetry.PoetryEnvBuildParameters{
				EntryResolver:           poetryEntryResolver,
				InstallProcess:          poetryProcess,
				PythonPathLookupProcess: poetryPythonPathProcess,
			},
		}

		build = pythoninstallers.Build(logger, buildParameters, packagerParameters)

		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "Some Buildpack",
				Version:     "some-version",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
			},
			WorkingDir: workingDir,
			CNBPath:    cnbDir,
			// Plan: shall be filled within each test
			Platform: packit.Platform{Path: "some-platform-path"},
			Layers:   packit.Layers{Path: layersDir},
			Stack:    "some-stack",
		}

		plans = []packit.BuildpackPlan{
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
					{
						Name: pip.Manager,
					},
					{
						Name: pipenv.Manager,
					},
					{
						Name: poetry.PoetryVenv,
					},
				},
			},
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
					{
						Name: pip.Manager,
					},
					{
						Name: pipenv.Manager,
					},
				},
			},
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
					{
						Name: pip.Manager,
					},
				},
			},
			packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: conda.CondaEnvPlanEntry,
					},
				},
			},
		}
	})

	it("runs the build process and returns expected layers", func() {
		for _, plan := range plans {
			buildContext.Plan = plan
			result, err := build(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(len(plan.Entries)))
		}
	})

	it("fails if packager parameters is missing", func() {
		packagerParameters := map[string]pythoninstallers.PackagerParameters{}

		build = pythoninstallers.Build(logger, buildParameters, packagerParameters)

		for _, plan := range plans {
			buildContext.Plan = plan
			_, err := build(buildContext)
			Expect(err).To(HaveOccurred())
		}
	})

}
