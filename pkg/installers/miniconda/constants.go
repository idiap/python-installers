// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package miniconda

const (
	// Conda is the name of the layer into which conda dependency is installed.
	Conda = "conda"

	// This is the key name that we use to store the sha of the script we
	// download in the layer metadata, which is used to determine if the conda
	// layer can be resued on during a rebuild.
	DepKey = "dependency-sha"

	// DepName is the name of the metadata.dependencies id
	DepId = "miniconda3"
)

// Priorities is a list of possible places where the buildpack could look for a
// specific version of miniconda3 to install, ordered from highest to lowest priority.
var Priorities = []interface{}{"BP_MINICONDA_VERSION"}
