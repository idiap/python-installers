// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixi

import (
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/common/build"
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

// InstallProcess defines the interface for installing the poetry dependency into a layer.
type InstallProcess interface {
	Execute(sourcePath, targetLayerPath, dependencyName string) error
}

// PixiBuildParameters encapsulates the pixi specific parameters for the
// Build function
type PixiBuildParameters struct {
	DependencyManager DependencyManager
	InstallProcess    InstallProcess
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right pixi dependency to download, download it
// into a layer, run the pixi-install script to install pixi into a separate
// layer and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(
	buildParameters PixiBuildParameters,
	parameters build.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		dependencyManager := buildParameters.DependencyManager
		installProcess := buildParameters.InstallProcess
		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		planner := draft.NewPlanner()

		logger.Process("Resolving pixi version")
		entry, sortedEntries := planner.Resolve(Pixi, context.Plan.Entries, Priorities)
		logger.Candidates(sortedEntries)

		version, _ := entry.Metadata["version"].(string)

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		legacySBOM := dependencyManager.GenerateBillOfMaterials(dependency)

		pixiLayer, err := context.Layers.Get("pixi")
		if err != nil {
			return packit.BuildResult{}, err
		}

		launch, build := planner.MergeLayerTypes("pixi", context.Plan.Entries)

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}
		if build {
			buildMetadata = packit.BuildMetadata{BOM: legacySBOM}
		}

		if launch {
			launchMetadata = packit.LaunchMetadata{BOM: legacySBOM}
		}

		cachedChecksum, ok := pixiLayer.Metadata[DepKey].(string)
		dependencyChecksum := dependency.Checksum
		if dependencyChecksum == "" {
			//nolint:staticcheck // SHA256 is only a fallback in case Checksum is not present
			dependencyChecksum = dependency.SHA256
		}

		if ok && cachedChecksum != "" && cargo.Checksum(cachedChecksum).MatchString(dependencyChecksum) {
			logger.Process("Reusing cached layer %s", pixiLayer.Path)
			logger.Break()

			pixiLayer.Launch, pixiLayer.Build, pixiLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{pixiLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		pixiLayer, err = pixiLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		pixiLayer.Launch, pixiLayer.Build, pixiLayer.Cache = launch, build, build

		// This temporary layer is created because the path to a deterministic and
		// easier to make assertions about during testing. Because this layer has
		// no type set to true the lifecycle will ensure that this layer is
		// removed.
		pixiScriptTempLayer, err := context.Layers.Get("pixi-temp-layer")
		if err != nil {
			return packit.BuildResult{}, err
		}

		pixiScriptTempLayer, err = pixiScriptTempLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Executing build process")
		logger.Subprocess("Installing pixi %s", dependency.Version)

		duration, err := clock.Measure(func() error {
			err := dependencyManager.Deliver(dependency, context.CNBPath, pixiScriptTempLayer.Path, context.Platform.Path)
			if err != nil {
				return err
			}

			return installProcess.Execute(pixiLayer.Path, pixiScriptTempLayer.Path, dependency.Arch)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		pixiLayer.Metadata = map[string]interface{}{
			DepKey: dependencyChecksum,
		}

		logger.GeneratingSBOM(pixiLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, pixiLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		pixiLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{pixiLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
