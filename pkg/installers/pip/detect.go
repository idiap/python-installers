// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pip

import (
	"os"
	"regexp"

	"github.com/paketo-buildpacks/packit/v2"

	"github.com/paketo-buildpacks/python-installers/pkg/build"
)

// Return a pip requirement
func GetVersionedRequirement() *packit.BuildPlanRequirement {
	pipVersion := os.Getenv(EnvVersion)
	if pipVersion == "" {
		return nil
	}

	// Pip releases are of the form X.Y rather than X.Y.0, so in order
	// to support selecting the exact version X.Y we have to up-convert
	// X.Y to X.Y.0.
	// Otherwise X.Y would match the latest patch release
	// X.Y.Z if it is available.
	var xDotYPattern = regexp.MustCompile(`^\d+\.\d+$`)
	if xDotYPattern.MatchString(pipVersion) {
		pipVersion = pipVersion + ".0"
	}

	return &packit.BuildPlanRequirement{
		Name: Pip,
		Metadata: build.BuildPlanMetadata{
			VersionSource: EnvVersion,
			Version:       pipVersion,
		},
	}
}

func GetRequirement() packit.BuildPlanRequirement {
	requirement := GetVersionedRequirement()
	if requirement != nil {
		return *requirement
	}

	return packit.BuildPlanRequirement{
		Name: Pip,
		Metadata: build.BuildPlanMetadata{
			Build: true,
		},
	}
}

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection always passes, and will contribute a  Build Plan that provides pip,
// and requires cpython OR python, python_packages, and requirements.
//
// If a version is provided via the $BP_PIP_VERSION environment variable, that
// version of pip will be a requirement.
func Detect() packit.DetectFunc {
	return func(_ packit.DetectContext) (packit.DetectResult, error) {

		requirements := []packit.BuildPlanRequirement{
			{
				Name: CPython,
				Metadata: build.BuildPlanMetadata{
					Build: true,
				},
			},
		}

		pipRequirement := GetVersionedRequirement()

		if pipRequirement != nil {
			requirements = append(requirements, *pipRequirement)
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Pip},
				},
				Requires: requirements,
			},
		}, nil
	}
}
