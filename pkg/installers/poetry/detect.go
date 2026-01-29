// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry

import (
	"os"
	"path/filepath"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/pip"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/fs"
)

//go:generate faux --interface PyProjectPythonVersionParser --output fakes/pyproject_parser.go
type PyProjectPythonVersionParser interface {
	// ParsePythonVersion extracts `tool.poetry.dependencies.python`
	// from pyproject.toml
	ParsePythonVersion(string) (string, error)
}

const PyProjectTomlFile = "pyproject.toml"

func Detect(parser PyProjectPythonVersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		pyProjectToml := filepath.Join(context.WorkingDir, PyProjectTomlFile)

		if exists, err := fs.Exists(pyProjectToml); err != nil {
			return packit.DetectResult{}, err
		} else if !exists {
			return packit.DetectResult{}, packit.Fail.WithMessage("%s is not present", PyProjectTomlFile)
		}

		pythonVersion, err := parser.ParsePythonVersion(pyProjectToml)
		if err != nil {
			return packit.DetectResult{}, err
		}

		if pythonVersion == "" {
			return packit.DetectResult{}, packit.Fail.WithMessage("%s must include [tool.poetry.dependencies.python], see https://python-poetry.org/docs/pyproject/#dependencies-and-dev-dependencies", PyProjectTomlFile)
		}

		requirements := []packit.BuildPlanRequirement{
			{
				Name: CPython,
				Metadata: pythoninstallers.BuildPlanMetadata{
					Build:         true,
					Version:       pythonVersion,
					VersionSource: PyProjectTomlFile,
				},
			},
		}

		requirements = append(requirements, pip.GetRequirement())

		if version, ok := os.LookupEnv("BP_POETRY_VERSION"); ok {
			requirements = append(requirements, packit.BuildPlanRequirement{
				Name: PoetryDependency,
				Metadata: pythoninstallers.BuildPlanMetadata{
					VersionSource: "BP_POETRY_VERSION",
					Version:       version,
				},
			})
		}

		return packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: Pip},
					{Name: PoetryDependency},
				},
				Requires: requirements,
			},
		}, nil
	}
}
