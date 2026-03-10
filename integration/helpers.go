// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package integration_helpers

import (
	"fmt"
	"testing"

	"github.com/paketo-buildpacks/occam"
)

type Buildpack struct {
	ID   string
	Name string
}

type Dependency struct {
	ID      string
	Version string
}

type Metadata struct {
	Dependencies []Dependency
}

type BuildpackInfo struct {
	Buildpack Buildpack
	Metadata  Metadata
}

type TestSettings struct {
	Buildpacks struct {
		// Dependency buildpacks
		CPython struct {
			Online  string
			Offline string
		}
		BuildPlan struct {
			Online string
		}
		// This buildpack
		PythonInstallers struct {
			Online  string
			Offline string
		}
	}

	Config struct {
		CPython   string `json:"cpython"`
		BuildPlan string `json:"build-plan"`
	}
}

func DependenciesForId(dependencies []Dependency, id string) []Dependency {
	output := []Dependency{}

	for _, entry := range dependencies {
		if entry.ID == id {
			output = append(output, entry)
		}
	}

	return output
}

func NewRetryBuild(t *testing.T, retry int) RetryBuild {
	return RetryBuild{t, retry}
}

type RetryBuild struct {
	t     *testing.T
	retry int
}

func (r *RetryBuild) Build(packBuild occam.PackBuild, name string, source string) (occam.Image, fmt.Stringer, error) {
	var image occam.Image
	var logs fmt.Stringer
	var err error

	for i := range r.retry {
		image, logs, err = packBuild.Execute(name, source)
		if err == nil {
			return image, logs, err
		} else {
			r.t.Logf("Build failed: %v\n", err)
			r.t.Logf("Retry %v\n", i)
		}
	}

	return image, logs, err
}
