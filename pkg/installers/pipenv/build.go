// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenv

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
//go:generate faux --interface InstallProcess --output fakes/install_process.go
//go:generate faux --interface SitePackageProcess --output fakes/site_package_process.go
//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go

// DependencyManager defines the interface for picking the best matching
// dependency and installing it.
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

// InstallProcess defines the interface for installing the pipenv dependency into a layer.
type InstallProcess interface {
	Execute(version, destLayerPath string) error
}

// SitePackageProcess defines the interface for looking up site packages within a layer.
type SitePackageProcess interface {
	Execute(targetLayerPath string) (string, error)
}

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

// PipEnvBuildParameters encapsulates the pip specific parameters for the
// Build function
type PipEnvBuildParameters struct {
	DependencyManager  DependencyManager
	InstallProcess     InstallProcess
	SitePackageProcess SitePackageProcess
}

// Build will return a packit.BuildFunc that will be invoked during the build
// phase of the buildpack lifecycle.
//
// Build will find the right pipenv dependency to install, install it in a
// layer, and generate Bill-of-Materials. It also makes use of the checksum of
// the dependency to reuse the layer when possible.
func Build(
	buildParameters PipEnvBuildParameters,
	parameters pythoninstallers.CommonBuildParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		installProcess := buildParameters.InstallProcess
		siteProcess := buildParameters.SitePackageProcess
		dependencyManager := buildParameters.DependencyManager
		sbomGenerator := parameters.SbomGenerator
		clock := parameters.Clock
		logger := parameters.Logger

		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		planner := draft.NewPlanner()

		logger.Process("Resolving Pipenv version")
		entry, sortedEntries := planner.Resolve(Pipenv, context.Plan.Entries, Priorities)

		logger.Candidates(sortedEntries)

		version, _ := entry.Metadata["version"].(string)

		dependency, err := dependencyManager.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, version, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		legacySBOM := dependencyManager.GenerateBillOfMaterials(dependency)
		launch, build := planner.MergeLayerTypes(Pipenv, context.Plan.Entries)

		var launchMetadata packit.LaunchMetadata
		if launch {
			launchMetadata.BOM = legacySBOM
		}

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = legacySBOM
		}

		pipenvLayer, err := context.Layers.Get(Pipenv)
		if err != nil {
			return packit.BuildResult{}, err
		}

		cachedChecksum, ok := pipenvLayer.Metadata[DependencyChecksumKey].(string)
		if ok && cachedChecksum == dependency.Checksum {
			logger.Process("Reusing cached layer %s", pipenvLayer.Path)
			pipenvLayer.Launch, pipenvLayer.Build, pipenvLayer.Cache = launch, build, build

			return packit.BuildResult{
				Layers: []packit.Layer{pipenvLayer},
				Build:  buildMetadata,
				Launch: launchMetadata,
			}, nil
		}

		pipenvLayer, err = pipenvLayer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		pipenvLayer.Launch, pipenvLayer.Build, pipenvLayer.Cache = launch, build, build

		logger.Process("Executing build process")
		logger.Subprocess(fmt.Sprintf("Installing Pipenv %s", dependency.Version))

		duration, err := clock.Measure(func() error {
			return installProcess.Execute(dependency.Version, pipenvLayer.Path)
		})

		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.GeneratingSBOM(pipenvLayer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, pipenvLayer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		pipenvLayer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		pipenvLayer.Metadata = map[string]interface{}{
			DependencyChecksumKey: dependency.Checksum,
		}

		// Look up the site packages path and prepend it onto $PYTHONPATH
		sitePackagesPath, err := siteProcess.Execute(pipenvLayer.Path)
		if err != nil {
			return packit.BuildResult{}, err
		}

		if sitePackagesPath == "" {
			return packit.BuildResult{}, fmt.Errorf("pipenv installation failed: site packages are missing from the pipenv layer")
		}

		pipenvLayer.SharedEnv.Prepend("PYTHONPATH", strings.TrimRight(sitePackagesPath, "\n"), ":")

		logger.EnvironmentVariables(pipenvLayer)

		return packit.BuildResult{
			Layers: []packit.Layer{pipenvLayer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
