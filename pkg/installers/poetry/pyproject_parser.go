// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry

import (
	"strings"

	"github.com/BurntSushi/toml"
)

type BuildSystem struct {
	Requires     []string
	BuildBackend string `toml:"build-backend"`
}

type PyProjectToml struct {
	Tool struct {
		Poetry struct {
			Dependencies struct {
				Python string
			}
		}
	}
	Project struct {
		RequiresPython string `toml:"requires-python"`
	}
	BuildSystem BuildSystem `toml:"build-system"`
}

type PoetryPyProjectParser struct {
}

func NewPyProjectParser() PoetryPyProjectParser {
	return PoetryPyProjectParser{}
}

func (p PoetryPyProjectParser) ParsePythonVersion(pyProjectToml string) (string, error) {
	var pyProject PyProjectToml

	_, err := toml.DecodeFile(pyProjectToml, &pyProject)
	if err != nil {
		return "", err
	}

	if pyProject.Project.RequiresPython != "" {
		return strings.Trim(pyProject.Project.RequiresPython, "="), nil
	}
	return strings.Trim(pyProject.Tool.Poetry.Dependencies.Python, "="), nil
}

func (p PoetryPyProjectParser) IsPoetryProject(pyProjectToml string) (bool, error) {
	var pyProject PyProjectToml

	_, err := toml.DecodeFile(pyProjectToml, &pyProject)
	if err != nil {
		return false, err
	}

	return pyProject.BuildSystem.BuildBackend == "" || pyProject.BuildSystem.BuildBackend == "poetry.core.masonry.api", nil
}
