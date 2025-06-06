// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers

import (
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	miniconda "github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda"
	pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
)

type PackagerParameters interface {
}

func validateResult(result packit.BuildResult, err error) (packit.BuildResult, error) {
	if err != nil {
		return packit.BuildResult{}, err
	}

	return result, err
}

func Build(
	logger scribe.Emitter,
	commonBuildParameters pythoninstallers.CommonBuildParameters,
	buildParameters map[string]PackagerParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		for _, entry := range context.Plan.Entries {
			logger.Title("Handling %s", entry.Name)
			parameters, ok := buildParameters[entry.Name]

			if !ok {
				return packit.BuildResult{}, packit.Fail.WithMessage("missing parameters for: %s", entry.Name)
			}

			switch entry.Name {
			case pip.Pip:
				result, err := pip.Build(
					parameters.(pip.PipBuildParameters),
					commonBuildParameters,
				)(context)

				return validateResult(result, err)

			case pipenv.Pipenv:
				result, err := pipenv.Build(
					parameters.(pipenv.PipEnvBuildParameters),
					commonBuildParameters,
				)(context)

				return validateResult(result, err)

			case miniconda.Conda:
				result, err := miniconda.Build(
					parameters.(miniconda.CondaBuildParameters),
					commonBuildParameters,
				)(context)

				return validateResult(result, err)

			case poetry.PoetryDependency:
				result, err := poetry.Build(
					parameters.(poetry.PoetryBuildParameters),
					commonBuildParameters,
				)(context)

				return validateResult(result, err)

			default:
				return packit.BuildResult{}, packit.Fail.WithMessage("unknown plan: %s", entry.Name)
			}
		}

		return packit.BuildResult{}, packit.Fail.WithMessage("empty plan: %s", context.Plan)
	}
}
