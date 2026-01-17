// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uv

const (
	// uv is the name of the layer into which uv dependency is installed.
	Uv = "uv"
	// LockfileName is the name of the uv lock file
	LockfileName = "uv.lock"

	// CPython is the name of the python runtime dependency provided by the CPython buildpack: https://github.com/paketo-buildpacks/cpython
	CPython = "cpython"

	// This is the key name that we use to store the sha of the script we
	// download in the layer metadata, which is used to determine if the uvs
	// layer can be reused on during a rebuild.
	DepKey = "dependency-sha"

	UvArchiveTemplateName = "uv-%s-unknown-linux-gnu"
)

var Priorities = []interface{}{
	"BP_UV_VERSION",
}
