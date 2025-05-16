// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenv

const (
	Pipenv                = "pipenv"
	DependencyChecksumKey = "dependency_checksum"
	CPython               = "cpython"
	Pip                   = "pip"
)

var Priorities = []interface{}{"BP_PIPENV_VERSION"}
