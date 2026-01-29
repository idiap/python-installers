// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenv

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
)

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// This buildpack always passes detection and will contribute a Build Plan that
// provides pipenv.
//
// If a version is provided via the $BP_PIPENV_VERSION environment variable,
// that version of pipenv will be a requirement.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

		requirements := []packit.BuildPlanRequirement{
			{
				Name: CPython,
				Metadata: pythoninstallers.BuildPlanMetadata{
					Build: true,
				},
			},
		}

		requirements = append(requirements, pip.GetRequirement())

		pipEnvVersion, ok := os.LookupEnv("BP_PIPENV_VERSION")
		if ok {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: Pipenv,
				Metadata: pythoninstallers.BuildPlanMetadata{
					Version:       pipEnvVersion,
					VersionSource: "BP_PIPENV_VERSION",
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Pip},
					{Name: Pipenv},
				},
				Requires: requirements,
			},
		}, nil
	}
}
