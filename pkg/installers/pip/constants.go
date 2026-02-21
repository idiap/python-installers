// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pip

const (
	// Pip is the name of the layer into which pip dependency is installed.
	Pip = "pip"

	PipSrc = "pip-source"

	// CPython is the name of the python runtime dependency provided by the CPython buildpack: https://github.com/paketo-buildpacks/cpython
	CPython = "cpython"

	// DependencyChecksumKey is the name of the key in the pip layer TOML whose value is pip dependency's SHA256.
	DependencyChecksumKey = "dependency_checksum"

	EnvVersion = "BP_PIP_VERSION"
)

// Priorities is a list of possible places where the buildpack could look for a
// specific version of Pip to install, ordered from highest to lowest priority.
var Priorities = []interface{}{EnvVersion}
