// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/common/build"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface InstallProcess --output fakes/install_process.go
//go:generate faux --interface SitePackageProcess --output fakes/site_package_process.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

// InstallProcess defines the interface for installing the poetry dependency into a layer.
type InstallProcess interface {
	Execute(version, targetLayerPath, pipLayerPath string) error
}

// SitePackageProcess defines the interface for looking site packages within a layer.
type SitePackageProcess interface {
	Execute(targetLayerPath string) (string, error)
}

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

// PoetryBuildParameters encapsulates the pip specific parameters for the
// Build function
type PoetryBuildParameters struct {
	DependencyManager  DependencyManager
	InstallProcess     InstallProcess
	SitePackageProcess SitePackageProcess
}

func Build(
	buildParameters PoetryBuildParameters,
	parameters build.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		installProcess := buildParameters.InstallProcess
		siteProcess := buildParameters.SitePackageProcess
		dependencyManager := buildParameters.DependencyManager
		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		logger.Process("Resolving Poetry version")
		planner := draft.NewPlanner()
		entry, entries := planner.Resolve(PoetryDependency, context.Plan.Entries, Priorities)
		logger.Candidates(entries)

		version, ok := entry.Metadata["version"].(string)
		if !ok {
			version = "default"
		}

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())
		legacySBOM := dependencyManager.GenerateBillOfMaterials(dependency)

		poetryLayer, err := context.Layers.Get(PoetryLayerName)
		if err != nil {
			return packit.BuildResult{}, err
		}

		launch, build := planner.MergeLayerTypes(PoetryDependency, context.Plan.Entries)

		var buildMetadata = packit.BuildMetadata{}
		var launchMetadata = packit.LaunchMetadata{}
		if build {
			buildMetadata = packit.BuildMetadata{BOM: legacySBOM}
		}

		if launch {
			launchMetadata = packit.LaunchMetadata{BOM: legacySBOM}
		}

		cachedChecksum, ok := poetryLayer.Metadata[DependencyChecksumKey].(string)
		if ok && cachedChecksum == dependency.Checksum {
			logger.Process("Reusing cached layer %s", poetryLayer.Path)
			logger.Break()

			poetryLayer.Launch, poetryLayer.Build, poetryLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{poetryLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		poetryLayer, err = poetryLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		poetryLayer.Launch, poetryLayer.Build, poetryLayer.Cache = launch, build, build

		logger.Process("Executing build process")
		logger.Subprocess("Installing Poetry %s", dependency.Version)
		pipLayer, err := context.Layers.Get(Pip)
		if err != nil {
			return packit.BuildResult{}, err
		}

		duration, err := clock.Measure(func() error {
			err = installProcess.Execute(dependency.Version, poetryLayer.Path, pipLayer.Path)
			if err != nil {
				return err
			}

			// Look up the site packages path and prepend it onto $PYTHONPATH
			sitePackagesPath, err := siteProcess.Execute(poetryLayer.Path)
			if err != nil {
				return fmt.Errorf("failed to locate site packages in poetry layer: %w", err)
			}

			if sitePackagesPath == "" {
				return fmt.Errorf("poetry installation failed: site packages are missing from the poetry layer")
			}

			poetryLayer.SharedEnv.Prepend("PYTHONPATH", strings.TrimRight(sitePackagesPath, "\n"), ":")

			return nil
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.GeneratingSBOM(poetryLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, poetryLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		poetryLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.EnvironmentVariables(poetryLayer)

		poetryLayer.Metadata = map[string]interface{}{
			DependencyChecksumKey: dependency.Checksum,
		}

		return packit.BuildResult{
			Layers: []packit.Layer{poetryLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
