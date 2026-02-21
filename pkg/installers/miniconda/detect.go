// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package miniconda

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"

	"github.com/paketo-buildpacks/python-installers/pkg/build"
)

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection always passes, and will contribute a  Build Plan that provides conda.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		var requirements []packit.BuildPlanRequirement

		if version, ok := os.LookupEnv(EnvVersion); ok {
			requirements = []packit.BuildPlanRequirement{
				{
					Name: Conda,
					Metadata: build.BuildPlanMetadata{
						VersionSource: EnvVersion,
						Version:       version,
					},
				},
			}
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Conda},
				},
				Requires: requirements,
			},
		}, nil
	}
}
