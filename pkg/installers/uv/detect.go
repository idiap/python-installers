// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uv

import (
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/common/build"
)

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// Detection always passes, and will contribute a  Build Plan that provides uv.
func Detect() packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		lockfile := filepath.Join(context.WorkingDir, LockfileName)

		if exists, err := fs.Exists(lockfile); err != nil {
			return packit.DetectResult{}, err
		} else if !exists {
			return packit.DetectResult{}, packit.Fail.WithMessage("%s is not present", LockfileName)
		}

		parser := NewLockfileParser()
		pythonVersion, err := parser.ParsePythonVersion(lockfile)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if pythonVersion == "" {
			return packit.DetectResult{}, packit.Fail.WithMessage("%s must include requires-python", LockfileName)
		}

		plan := packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: Uv},
			},
		}

		if version, ok := os.LookupEnv("BP_UV_VERSION"); ok {
			plan.Requires = []packit.BuildPlanRequirement{
				{
					Name: Uv,
					Metadata: build.BuildPlanMetadata{
						VersionSource: "BP_UV_VERSION",
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
