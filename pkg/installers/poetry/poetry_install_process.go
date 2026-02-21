// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/packit/v2/pexec"

	"github.com/paketo-buildpacks/python-installers/pkg/executable"
)

type PoetryInstallProcess struct {
	executable executable.Executable
}

// NewPoetryInstallProcess creates a PoetryInstallProcess instance.
func NewPoetryInstallProcess(executable executable.Executable) PoetryInstallProcess {
	return PoetryInstallProcess{
		executable: executable,
	}
}

// Execute installs the provided version of pipenv from the internet into the
// layer path designated by targetLayerPath
func (p PoetryInstallProcess) Execute(version, targetLayerPath, pipLayerPath string) error {
	buffer := bytes.NewBuffer(nil)

	pipPath := fmt.Sprintf("PYTHONPATH=%s", filepath.Join(pipLayerPath))
	err := p.executable.Execute(pexec.Execution{
		Args: []string{"-m", "pip", "install", fmt.Sprintf("poetry==%s", version), "--user"},
		// Set the PYTHONUSERBASE to ensure that poetry is installed to the newly created target layer.
		Env:    append(os.Environ(), pipPath, fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath)),
		Stdout: buffer,
		Stderr: buffer,
	})

	if err != nil {
		return fmt.Errorf("failed to configure poetry:\n%s\nerror: %w", buffer.String(), err)
	}

	return nil
}
