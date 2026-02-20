// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"

	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	pythoninstallers "github.com/paketo-buildpacks/python-installers"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/common/build"
	dependencyfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/common/dependency/fakes"
	sbomfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/common/sbom/fakes"
	miniconda "github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda"
	minicondafakes "github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda/fakes"
	pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	pipfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/pip/fakes"
	pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	pipenvfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv/fakes"
	pixi "github.com/paketo-buildpacks/python-installers/pkg/installers/pixi"
	pixifakes "github.com/paketo-buildpacks/python-installers/pkg/installers/pixi/fakes"
	poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
	poetryfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry/fakes"
	uv "github.com/paketo-buildpacks/python-installers/pkg/installers/uv"
	uvfakes "github.com/paketo-buildpacks/python-installers/pkg/installers/uv/fakes"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

type TestPlan struct {
	Plan             packit.BuildpackPlan
	OutputLayerCount int
}

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		workingDir string
		cnbDir     string

		buffer       *bytes.Buffer
		logger       scribe.Emitter
		buildFunc    packit.BuildFunc
		buildContext packit.BuildContext

		// common
		sbomGenerator *sbomfakes.SBOMGenerator
		// dependencyManager *dependencyfakes.DependencyManager

		// conda
		minicondaDependencyManager *dependencyfakes.DependencyManager
		runner                     *minicondafakes.Runner

		// pip
		pipDependencyManager  *dependencyfakes.DependencyManager
		pipInstallProcess     *pipfakes.InstallProcess
		pipSitePackageProcess *pipfakes.SitePackageProcess

		// pipenv
		pipenvDependencyManager  *dependencyfakes.DependencyManager
		pipenvProcess            *pipenvfakes.InstallProcess
		pipenvSitePackageProcess *pipenvfakes.SitePackageProcess

		// poetry
		poetryDependencyManager  *dependencyfakes.DependencyManager
		poetryProcess            *poetryfakes.InstallProcess
		poetrySitePackageProcess *poetryfakes.SitePackageProcess

		// uv
		uvDependencyManager *dependencyfakes.DependencyManager
		uvInstallProcess    *uvfakes.InstallProcess

		// pixi
		pixiDependencyManager *dependencyfakes.DependencyManager
		pixiInstallProcess    *pixifakes.InstallProcess

		buildParameters build.CommonBuildParameters

		testPlans []TestPlan
	)

	it.Before(func() {
		layersDir = t.TempDir()
		workingDir = t.TempDir()
		cnbDir = t.TempDir()

		buffer = bytes.NewBuffer(nil)
		logger = scribe.NewEmitter(buffer)

		sbomGenerator = &sbomfakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		// miniconda
		minicondaDependencyManager = &dependencyfakes.DependencyManager{}
		minicondaDependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "miniconda3",
			Name:     "miniconda3-dependency-name",
			Checksum: "miniconda3-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "miniconda3-dependency-uri",
			Version:  "miniconda3-dependency-version",
		}

		// Legacy SBOM
		minicondaDependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "miniconda3",
				Metadata: paketosbom.BOMMetadata{
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "miniconda3-dependency-sha",
					},
					URI:     "miniconda3-dependency-uri",
					Version: "miniconda3-dependency-version",
				},
			},
		}

		runner = &minicondafakes.Runner{}

		// pip
		pipDependencyManager = &dependencyfakes.DependencyManager{}
		pipDependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "pip",
			Name:     "Pip",
			Checksum: "some-sha",
			Stacks:   []string{"some-stack"},
			URI:      "some-uri",
			Version:  "21.0",
		}

		// Legacy SBOM
		pipDependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "pip",
				Metadata: paketosbom.BOMMetadata{
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "pip-dependency-sha",
					},
					URI:     "pip-dependency-uri",
					Version: "pip-dependency-version",
				},
			},
		}

		pipInstallProcess = &pipfakes.InstallProcess{}
		pipInstallProcess.ExecuteCall.Stub = func(srcPath, targetLayerPath string) error {
			err := os.MkdirAll(filepath.Join(layersDir, "pip", "lib", "python1.23", "site-packages"), os.ModePerm)
			if err != nil {
				return fmt.Errorf("issue with stub call: %s", err)
			}
			return nil
		}
		pipSitePackageProcess = &pipfakes.SitePackageProcess{}
		pipSitePackageProcess.ExecuteCall.Returns.String = filepath.Join(layersDir, "pip", "lib", "python1.23", "site-packages")

		// pipenv
		pipenvDependencyManager = &dependencyfakes.DependencyManager{}
		pipenvDependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "pipenv",
			Name:     "pipenv-dependency-name",
			Checksum: "pipenv-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "pipenv-dependency-uri",
			Version:  "pipenv-dependency-version",
		}

		// Legacy SBOM
		pipenvDependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "pipenv",
				Metadata: paketosbom.BOMMetadata{
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "pipenv-dependency-sha",
					},
					URI:     "pipenv-dependency-uri",
					Version: "pipenv-dependency-version",
				},
			},
		}

		pipenvProcess = &pipenvfakes.InstallProcess{}
		pipenvSitePackageProcess = &pipenvfakes.SitePackageProcess{}
		pipenvSitePackageProcess.ExecuteCall.Returns.String = filepath.Join(layersDir, "pipenv", "lib", "python3.8", "site-packages")

		// poetry
		poetryDependencyManager = &dependencyfakes.DependencyManager{}
		poetryDependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "poetry",
			Name:     "poetry-dependency-name",
			Checksum: "poetry-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "poetry-dependency-uri",
			Version:  "poetry-dependency-version",
		}

		poetryDependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "poetry",
				Metadata: paketosbom.BOMMetadata{
					Version: "poetry-dependency-version",
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "poetry-dependency-sha",
					},
					URI: "poetry-dependency-uri",
				},
			},
		}

		poetryProcess = &poetryfakes.InstallProcess{}
		poetrySitePackageProcess = &poetryfakes.SitePackageProcess{}
		poetrySitePackageProcess.ExecuteCall.Returns.String = filepath.Join(layersDir, "poetry", "lib", "python3.8", "site-packages")

		// uv
		uvDependencyManager = &dependencyfakes.DependencyManager{}
		uvDependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "uv",
			Name:     "uv-dependency-name",
			Checksum: "uv-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "uv-dependency-uri",
			Version:  "uv-dependency-version",
		}

		// Legacy SBOM
		uvDependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "uv",
				Metadata: paketosbom.BOMMetadata{
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "uv-dependency-sha",
					},
					URI:     "uv-dependency-uri",
					Version: "uv-dependency-version",
				},
			},
		}

		uvInstallProcess = &uvfakes.InstallProcess{}

		// pixi
		pixiDependencyManager = &dependencyfakes.DependencyManager{}
		pixiDependencyManager.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:       "pixi",
			Name:     "pixi-dependency-name",
			Checksum: "pixi-dependency-sha",
			Stacks:   []string{"some-stack"},
			URI:      "pixi-dependency-uri",
			Version:  "pixi-dependency-version",
		}

		// Legacy SBOM
		pixiDependencyManager.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
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

		pixiInstallProcess = &pixifakes.InstallProcess{}

		buildParameters = build.CommonBuildParameters{
			SbomGenerator: sbomGenerator,
			Clock:         chronos.DefaultClock,
			Logger:        logger,
		}

		packagerParameters := map[string]pythoninstallers.PackagerParameters{
			miniconda.Conda: miniconda.CondaBuildParameters{
				DependencyManager: minicondaDependencyManager,
				Runner:            runner,
			},
			pip.Pip: pip.PipBuildParameters{
				DependencyManager:  pipDependencyManager,
				InstallProcess:     pipInstallProcess,
				SitePackageProcess: pipSitePackageProcess,
			},
			pipenv.Pipenv: pipenv.PipEnvBuildParameters{
				DependencyManager:  pipenvDependencyManager,
				InstallProcess:     pipenvProcess,
				SitePackageProcess: pipenvSitePackageProcess,
			},
			pixi.Pixi: pixi.PixiBuildParameters{
				DependencyManager: pixiDependencyManager,
				InstallProcess:    pixiInstallProcess,
			},
			poetry.PoetryDependency: poetry.PoetryBuildParameters{
				DependencyManager:  poetryDependencyManager,
				InstallProcess:     poetryProcess,
				SitePackageProcess: poetrySitePackageProcess,
			},
			uv.Uv: uv.UvBuildParameters{
				DependencyManager: uvDependencyManager,
				InstallProcess:    uvInstallProcess,
			},
		}

		buildFunc = pythoninstallers.Build(logger, buildParameters, packagerParameters)

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

		testPlans = []TestPlan{
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: miniconda.Conda,
						},
					},
				},
				1,
			},
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: pip.Pip,
						},
					},
				},
				2,
			},
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: pipenv.Pipenv,
						},
					},
				},
				1,
			},
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: poetry.PoetryDependency,
						},
					},
				},
				1,
			},
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: uv.Uv,
						},
					},
				},
				1,
			},
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: pixi.Pixi,
						},
					},
				},
				1,
			},
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: pip.Pip,
						},
						{
							Name: pipenv.Pipenv,
						},
					},
				},
				3,
			},
			{
				packit.BuildpackPlan{
					Entries: []packit.BuildpackPlanEntry{
						{
							Name: pip.Pip,
						},
						{
							Name: poetry.PoetryDependency,
						},
					},
				},
				3,
			},
		}
		Expect(os.WriteFile(filepath.Join(workingDir, "x.py"), []byte{}, os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "pyproject.toml"), []byte(""), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "uv.lock"), []byte(`python-requires = "3.13.0"`), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "pixi.lock"), []byte(``), 0755)).To(Succeed())
	})

	it("runs the build process and returns expected layers", func() {
		for _, testPlan := range testPlans {
			logger.Detail("Doing: %s", testPlan)
			buildContext.Plan = testPlan.Plan
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers).To(HaveLen(testPlan.OutputLayerCount))
		}
	})

	it("runs the build process and returns layers in expected order", func() {
		orderTestPlans := []packit.BuildpackPlan{
			{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: pipenv.Pipenv,
					},
					{
						Name: pip.Pip,
					},
				},
			},
			{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: poetry.PoetryDependency,
					},
					{
						Name: pip.Pip,
					},
				},
			},
		}
		for _, testPlan := range orderTestPlans {
			logger.Detail("Doing: %s", testPlan)
			buildContext.Plan = testPlan
			result, err := buildFunc(buildContext)
			Expect(err).NotTo(HaveOccurred())

			layers := result.Layers
			Expect(layers[0].Name).To(Equal(pip.Pip))
			// Pip adds two layers
			Expect(layers[2].Name).To(Equal(testPlan.Entries[0].Name))
		}
	})

	it("fails if packager parameters is missing", func() {
		packagerParameters := map[string]pythoninstallers.PackagerParameters{}

		buildFunc = pythoninstallers.Build(logger, buildParameters, packagerParameters)

		for _, testPlan := range testPlans {
			buildContext.Plan = testPlan.Plan
			_, err := buildFunc(buildContext)
			Expect(err).To(HaveOccurred())
		}
	})

}
