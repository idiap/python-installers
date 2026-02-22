// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixi

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"

	"github.com/paketo-buildpacks/python-installers/pkg/build"
)

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection always passes, and will contribute a  Build Plan that provides uv.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		lockfile := filepath.Join(context.WorkingDir, LockfileName)
		lockfileExists, err := fs.Exists(lockfile)
		if err != nil {
			return packit.DetectResult{}, err
		}

		projectFile := filepath.Join(context.WorkingDir, ProjectFilename)
		projeectfileExists, err := fs.Exists(projectFile)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if !lockfileExists && !projeectfileExists {
			return packit.DetectResult{}, packit.Fail.WithMessage("neither %s nor %s are present", LockfileName, ProjectFilename)
		}

		plan := packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: Pixi},
			},
		}

		if version, ok := os.LookupEnv(EnvVersion); ok {
			plan.Requires = []packit.BuildPlanRequirement{
				{
					Name: Pixi,
					Metadata: build.BuildPlanMetadata{
						VersionSource: EnvVersion,
						Version:       version,
					},
				},
			}
		}

		return packit.DetectResult{
			Plan: plan,
		}, nil
	}
}
