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
	uv "github.com/paketo-buildpacks/python-installers/pkg/installers/uv"
)

// Detect will return a packit.DetectFunc that will be invoked during the
// detect phase of the buildpack lifecycle.
//
// If this buildpack detects files that indicate your app is a Python project,
// it will pass detection.
func Detect(logger scribe.Emitter, pyprojectVersionParser poetry.PyProjectPythonVersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {
		plans := []packit.BuildPlan{}

		pipResult, err := pip.Detect()(context)

		if err == nil {
			plans = append(plans, pipResult.Plan)
		} else {
			logger.Detail("%s", err)
		}

		minicondaResult, err := miniconda.Detect()(context)

		if err == nil {
			plans = append(plans, minicondaResult.Plan)
		} else {
			logger.Detail("%s", err)
		}

		pipenvResult, err := pipenv.Detect()(context)

		if err == nil {
			plans = append(plans, pipenvResult.Plan)
		} else {
			logger.Detail("%s", err)
		}

		poetryResult, err := poetry.Detect(pyprojectVersionParser)(context)

		if err == nil {
			plans = append(plans, poetryResult.Plan)
		} else {
			logger.Detail("%s", err)
		}

		uvResult, err := uv.Detect()(context)

		if err == nil {
			plans = append(plans, uvResult.Plan)
		} else {
			logger.Detail("%s", err)
		}

		if len(plans) == 0 {
			return packit.DetectResult{}, packit.Fail.WithMessage("No python packager manager related files found")
		}

		return packit.DetectResult{
			Plan: Or(plans...),
		}, nil
	}
}

func Or(plans ...packit.BuildPlan) packit.BuildPlan {
	if len(plans) < 1 {
		return packit.BuildPlan{}
	}
	combinedPlan := plans[0]

	for i := range plans {
		if i == 0 {
			continue
		}
		combinedPlan.Or = append(combinedPlan.Or, plans[i])
	}
	return combinedPlan
}
