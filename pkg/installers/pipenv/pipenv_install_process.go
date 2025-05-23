// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pipenv

import (
	"bytes"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/v2/pexec"
)

//go:generate faux --interface Executable --output fakes/executable.go

// Executable defines the interface for invoking an executable.
type Executable interface {
	Execute(pexec.Execution) error
}

type PipenvInstallProcess struct {
	executable Executable
}

// NewPipenvInstallProcess creates a PipenvInstallProcess instance.
func NewPipenvInstallProcess(executable Executable) PipenvInstallProcess {
	return PipenvInstallProcess{
		executable: executable,
	}
}

// Execute installs the provided version of pipenv from the internet into the
// layer path designated by targetLayerPath
func (p PipenvInstallProcess) Execute(version, targetLayerPath string) error {
	buffer := bytes.NewBuffer(nil)

	err := p.executable.Execute(pexec.Execution{
		Args: []string{"install", fmt.Sprintf("pipenv==%s", version), "--user"},
		// Set the PYTHONUSERBASE to ensure that pip is installed to the newly created target layer.
		Env:    append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath)),
		Stdout: buffer,
		Stderr: buffer,
	})

	if err != nil {
		return fmt.Errorf("failed to configure pipenv:\n%s\nerror: %w", buffer.String(), err)
	}

	return nil
}
