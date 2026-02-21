// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry

const (
	DependencyChecksumKey = "dependency-checksum"
	PoetryDependency      = "poetry"
	PoetryLayerName       = "poetry"
	CPython               = "cpython"
	Pip                   = "pip"
	EnvVersion            = "BP_POETRY_VERSION"
)

var Priorities = []interface{}{
	EnvVersion,
}
