// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers

import (
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error)
}

type Generator struct{}

func (f Generator) GenerateFromDependency(dependency postal.Dependency, path string) (sbom.SBOM, error) {
	return sbom.GenerateFromDependency(dependency, path)
}

// CommonBuildParameters are the parameters shared
// by all packager build function implementation
type CommonBuildParameters struct {
	SbomGenerator SBOMGenerator
	Clock         chronos.Clock
	Logger        scribe.Emitter
}

// BuildPlanMetadata is the buildpack specific data included in build plan
// requirements.
type BuildPlanMetadata struct {
	// Build denotes the dependency is needed at build-time.
	Build bool `toml:"build"`

	// Launch denotes the dependency is needed at runtime.
	Launch bool `toml:"launch"`

	// Version denotes the version of a dependency, if there is one.
	Version string `toml:"version"`

	// VersionSource denotes where dependency version came from (e.g. an environment variable).
	VersionSource string `toml:"version-source"`
}
