// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uv

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/draft"
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

// InstallProcess defines the interface for installing the uv dependency into a layer.
type InstallProcess interface {
	Execute(sourcePath, targetLayerPath, dependencyName string) error
}

// UvBuildParameters encapsulates the uv specific parameters for the
// Build function
type UvBuildParameters struct {
	DependencyManager DependencyManager
	InstallProcess    InstallProcess
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right uv dependency to download, download it
// into a layer, run the uv-install script to install uv into a separate
// layer and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(
	buildParameters UvBuildParameters,
	parameters pythoninstallers.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		dependencyManager := buildParameters.DependencyManager
		installProcess := buildParameters.InstallProcess
		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		planner := draft.NewPlanner()

		logger.Process("Resolving uv version")
		entry, sortedEntries := planner.Resolve(Uv, context.Plan.Entries, Priorities)
		logger.Candidates(sortedEntries)

		version, _ := entry.Metadata["version"].(string)

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		legacySBOM := dependencyManager.GenerateBillOfMaterials(dependency)

		uvLayer, err := context.Layers.Get("uv")
		if err != nil {
			return packit.BuildResult{}, err
		}

		launch, build := planner.MergeLayerTypes("uv", context.Plan.Entries)

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}
		if build {
			buildMetadata = packit.BuildMetadata{BOM: legacySBOM}
		}

		if launch {
			launchMetadata = packit.LaunchMetadata{BOM: legacySBOM}
		}

		cachedChecksum, ok := uvLayer.Metadata[DepKey].(string)
		dependencyChecksum := dependency.Checksum
		if dependencyChecksum == "" {
			//nolint:staticcheck // SHA256 is only a fallback in case Checksum is not present
			dependencyChecksum = dependency.SHA256
		}

		if ok && cachedChecksum != "" && cargo.Checksum(cachedChecksum).MatchString(dependencyChecksum) {
			logger.Process("Reusing cached layer %s", uvLayer.Path)
			logger.Break()

			uvLayer.Launch, uvLayer.Build, uvLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{uvLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		uvLayer, err = uvLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		uvLayer.Launch, uvLayer.Build, uvLayer.Cache = launch, build, build

		// This temporary layer is created because the path to a deterministic and
		// easier to make assertions about during testing. Because this layer has
		// no type set to true the lifecycle will ensure that this layer is
		// removed.
		uvScriptTempLayer, err := context.Layers.Get("uv-temp-layer")
		if err != nil {
			return packit.BuildResult{}, err
		}

		uvScriptTempLayer, err = uvScriptTempLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Executing build process")
		logger.Subprocess("Installing uv %s", dependency.Version)

		duration, err := clock.Measure(func() error {
			err := dependencyManager.Deliver(dependency, context.CNBPath, uvScriptTempLayer.Path, context.Platform.Path)
			if err != nil {
				return err
			}

			return installProcess.Execute(uvLayer.Path, uvScriptTempLayer.Path, dependency.Arch)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		uvLayer.Metadata = map[string]interface{}{
			DepKey: dependencyChecksum,
		}

		logger.GeneratingSBOM(uvLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, uvLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		uvLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{uvLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
