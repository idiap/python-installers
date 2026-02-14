// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package miniconda

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface Runner --output fakes/runner.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

// DependencyManager defines the interface for picking the best matching
// dependency and installing it.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, destinationPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

// Runner defines the interface for invoking the miniconda script downloaded as a dependency.
type Runner interface {
	Run(runPath, layerPath string) error
}

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

func GetEnvOrDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

// CondaBuildParameters encapsulates the conda specific parameters for the
// Build function
type CondaBuildParameters struct {
	DependencyManager DependencyManager
	Runner            Runner
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right miniconda dependency to download, download it
// into a layer, run the miniconda script to install conda into a separate
// layer and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(
	buildParameters CondaBuildParameters,
	parameters pythoninstallers.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		runner := buildParameters.Runner
		dependencyManager := buildParameters.DependencyManager
		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		planner := draft.NewPlanner()

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), "miniconda3", "*", context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		legacySBOM := dependencyManager.GenerateBillOfMaterials(dependency)

		condaLayer, err := context.Layers.Get("conda")
		if err != nil {
			return packit.BuildResult{}, err
		}

		launch, build := planner.MergeLayerTypes("conda", context.Plan.Entries)

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}
		if build {
			buildMetadata = packit.BuildMetadata{BOM: legacySBOM}
		}

		if launch {
			launchMetadata = packit.LaunchMetadata{BOM: legacySBOM}
		}

		cachedChecksum, ok := condaLayer.Metadata[DepKey].(string)
		dependencyChecksum := dependency.Checksum
		if dependencyChecksum == "" {
			//nolint:staticcheck // SHA256 is only a fallback in case Checksum is not present
			dependencyChecksum = dependency.SHA256
		}

		if ok && cachedChecksum != "" && cargo.Checksum(cachedChecksum).MatchString(dependencyChecksum) {
			logger.Process("Reusing cached layer %s", condaLayer.Path)
			logger.Break()

			condaLayer.Launch, condaLayer.Build, condaLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{condaLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		condaLayer, err = condaLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		condaLayer.Launch, condaLayer.Build, condaLayer.Cache = launch, build, build

		// This temporary layer is created because the path to a deterministic and
		// easier to make assertions about during testing. Because this layer has
		// no type set to true the lifecycle will ensure that this layer is
		// removed.
		minicondaScriptTempLayer, err := context.Layers.Get("miniconda-script-temp-layer")
		if err != nil {
			return packit.BuildResult{}, err
		}

		minicondaScriptTempLayer, err = minicondaScriptTempLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Executing build process")
		logger.Subprocess("Installing Miniconda %s", dependency.Version)

		duration, err := clock.Measure(func() error {
			err := dependencyManager.Deliver(dependency, context.CNBPath, minicondaScriptTempLayer.Path, context.Platform.Path)
			if err != nil {
				return err
			}

			scriptPath := filepath.Join(minicondaScriptTempLayer.Path, dependency.Name)
			return runner.Run(scriptPath, condaLayer.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		solver := GetEnvOrDefault("BP_CONDA_SOLVER", "conda")

		if solver == "mamba" {
			logger.Subprocess("Installing mamba solver")

			conda := pexec.NewExecutable(condaLayer.Path + "/bin/conda")
			duration, err = clock.Measure(func() error {
				buffer := bytes.NewBuffer(nil)
				err := conda.Execute(pexec.Execution{
					Args:   []string{"install", "-n", "base", "conda-libmamba-solver", "-y"},
					Env:    append(os.Environ(), "CONDA_PLUGINS_AUTO_ACCEPT_TOS=true"),
					Stdout: buffer,
					Stderr: buffer,
				})
				if err != nil {
					return fmt.Errorf("failed to setup solver:\n%s\nerror: %w", buffer.String(), err)
				}
				return nil
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Action("Solver completed in %s", duration.Round(time.Millisecond))
			logger.Break()

			logger.Subprocess("Configuring mamba solver")
			duration, err = clock.Measure(func() error {
				return conda.Execute(pexec.Execution{
					Args: []string{"config", "--set", "solver", "libmamba"},
				})
			})
			if err != nil {
				return packit.BuildResult{}, err
			}

			logger.Action("Configuration completed in %s", duration.Round(time.Millisecond))
			logger.Break()

		}

		condaLayer.SharedEnv.Append("CONDA_PLUGINS_AUTO_ACCEPT_TOS", "true", ":")
		condaLayer.Metadata = map[string]interface{}{
			DepKey: dependencyChecksum,
		}

		logger.GeneratingSBOM(condaLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, condaLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		condaLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{condaLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
