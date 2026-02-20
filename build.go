// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers

import (
	"slices"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	miniconda "github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda"
	pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	pixi "github.com/paketo-buildpacks/python-installers/pkg/installers/pixi"
	poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
	uv "github.com/paketo-buildpacks/python-installers/pkg/installers/uv"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
)

type PackagerParameters interface {
}

func Build(
	logger scribe.Emitter,
	commonBuildParameters pythoninstallers.CommonBuildParameters,
	buildParameters map[string]PackagerParameters,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {

		if len(context.Plan.Entries) == 0 {
			return packit.BuildResult{}, packit.Fail.WithMessage("empty plan: %s", context.Plan)
		}

		var results []packit.BuildResult

		orderedInstallers := []string{
			pip.Pip,
			pipenv.Pipenv,
			poetry.PoetryDependency,
			miniconda.Conda,
			uv.Uv,
			pixi.Pixi,
		}

		doneInstallers := []string{}

		for _, installer := range orderedInstallers {
			for _, entry := range context.Plan.Entries {
				if slices.Contains(doneInstallers, entry.Name) {
					continue
				}
				if entry.Name == installer {
					doneInstallers = append(doneInstallers, entry.Name)
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

						if err != nil {
							return packit.BuildResult{}, err
						}
						results = append(results, result)

					case pipenv.Pipenv:
						result, err := pipenv.Build(
							parameters.(pipenv.PipEnvBuildParameters),
							commonBuildParameters,
						)(context)

						if err != nil {
							return packit.BuildResult{}, err
						}
						results = append(results, result)

					case poetry.PoetryDependency:
						result, err := poetry.Build(
							parameters.(poetry.PoetryBuildParameters),
							commonBuildParameters,
						)(context)

						if err != nil {
							return packit.BuildResult{}, err
						}
						results = append(results, result)

					case miniconda.Conda:
						result, err := miniconda.Build(
							parameters.(miniconda.CondaBuildParameters),
							commonBuildParameters,
						)(context)

						if err != nil {
							return packit.BuildResult{}, err
						}
						results = append(results, result)

					case uv.Uv:
						result, err := uv.Build(
							parameters.(uv.UvBuildParameters),
							commonBuildParameters,
						)(context)

						if err != nil {
							return packit.BuildResult{}, err
						}
						results = append(results, result)

					case pixi.Pixi:
						result, err := pixi.Build(
							parameters.(pixi.PixiBuildParameters),
							commonBuildParameters,
						)(context)

						if err != nil {
							return packit.BuildResult{}, err
						}
						results = append(results, result)

					default:
						return packit.BuildResult{}, packit.Fail.WithMessage("unknown plan: %s", entry.Name)
					}
				}
			}
		}

		return combineResults(results...), nil
	}
}

func combineResults(results ...packit.BuildResult) packit.BuildResult {
	if len(results) < 1 {
		return packit.BuildResult{}
	}
	combinedResults := results[0]

	for i := range results {
		if i == 0 {
			continue
		}
		combinedResults.Layers = append(combinedResults.Layers, results[i].Layers...)
		combinedResults.Launch.BOM = append(combinedResults.Launch.BOM, results[i].Launch.BOM...)
		combinedResults.Build.BOM = append(combinedResults.Build.BOM, results[i].Build.BOM...)
	}
	return combinedResults
}
