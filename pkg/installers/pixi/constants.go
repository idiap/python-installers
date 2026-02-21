// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixi

const (
	// pixi is the name of the layer into which pixi dependency is installed.
	Pixi = "pixi"
	// LockfileName is the name of the pixi lock file
	LockfileName = "pixi.lock"
	// ProjectFilename is the name of the pixi project file
	ProjectFilename = "pixi.toml"

	// This is the key name that we use to store the sha of the script we
	// download in the layer metadata, which is used to determine if the uvs
	// layer can be reused on during a rebuild.
	DepKey = "dependency-sha"

	PixiArchiveTemplateName = "uv-%s-unknown-linux-musl"

	EnvVersion = "BP_PIXI_VERSION"
)

var Priorities = []interface{}{
	EnvVersion,
}
