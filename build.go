// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers

import (
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	// conda "github.com/paketo-buildpacks/python-installers/pkg/installers/conda"
	// pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	// pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	// poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
)

// filtered returns the slice passed in parameter with the needle removed
func filtered(haystack []packit.BuildpackPlanEntry, needle string) []packit.BuildpackPlanEntry {
	output := []packit.BuildpackPlanEntry{}

	for _, entry := range haystack {
		if entry.Name != needle {
			output = append(output, entry)
		}
	}

	return output
}

type PackagerParameters interface {
}

func Build(
	logger scribe.Emitter,
	commonBuildParameters pythoninstallers.CommonBuildParameters,
	// buildParameters map[string]PackagerParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		// planEntries := filtered(context.Plan.Entries, pip.SitePackages)
		layers := []packit.Layer{}

		// for _, entry := range planEntries {
		// 	logger.Title("Handling %s", entry.Name)

		// 	switch entry.Name {
		// 	case pip.Manager:
		// 		if parameters, ok := buildParameters[pip.Manager]; ok {
		// 			pipResult, err := pip.Build(
		// 				parameters.(pip.PipBuildParameters),
		// 				commonBuildParameters,
		// 			)(context)

		// 			if err != nil {
		// 				return packit.BuildResult{}, err
		// 			}

		// 			layers = append(layers, pipResult.Layers...)
		// 		} else {
		// 			return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
		// 		}

		// 	case pipenv.Manager:
		// 		if parameters, ok := buildParameters[pipenv.Manager]; ok {
		// 			pipEnvResult, err := pipenv.Build(
		// 				parameters.(pipenv.PipEnvBuildParameters),
		// 				commonBuildParameters,
		// 			)(context)

		// 			if err != nil {
		// 				return packit.BuildResult{}, err
		// 			}

		// 			layers = append(layers, pipEnvResult.Layers...)
		// 		} else {
		// 			return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
		// 		}
		// 	case conda.CondaEnvPlanEntry:
		// 		if parameters, ok := buildParameters[conda.CondaEnvPlanEntry]; ok {
		// 			condaResult, err := conda.Build(
		// 				parameters.(conda.CondaBuildParameters),
		// 				commonBuildParameters,
		// 			)(context)

		// 			if err != nil {
		// 				return packit.BuildResult{}, err
		// 			}

		// 			layers = append(layers, condaResult.Layers...)
		// 		} else {
		// 			return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
		// 		}
		// 	case poetry.PoetryVenv:
		// 		if parameters, ok := buildParameters[poetry.PoetryVenv]; ok {
		// 			poetryResult, err := poetry.Build(
		// 				parameters.(poetry.PoetryEnvBuildParameters),
		// 				commonBuildParameters,
		// 			)(context)

		// 			if err != nil {
		// 				return packit.BuildResult{}, err
		// 			}

		// 			layers = append(layers, poetryResult.Layers...)
		// 		} else {
		// 			return packit.BuildResult{}, packit.Fail.WithMessage("missing plan for: %s", entry.Name)
		// 		}
		// 	default:
		// 		return packit.BuildResult{}, packit.Fail.WithMessage("unknown plan: %s", entry.Name)
		// 	}
		// }

		if len(layers) == 0 {
			return packit.BuildResult{}, packit.Fail.WithMessage("empty plan should not happen")
		}

		return packit.BuildResult{
			Layers: layers,
		}, nil
	}
}
