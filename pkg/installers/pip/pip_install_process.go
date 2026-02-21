// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pip

import (
	"bytes"
	"fmt"
	"os"

	"github.com/paketo-buildpacks/packit/v2/pexec"

	"github.com/paketo-buildpacks/python-installers/pkg/executable"
)

// PipInstallProcess implements the InstallProcess interface.
type PipInstallProcess struct {
	executable executable.Executable
}

// NewPipInstallProcess creates an instance of the PipInstallProcess given an Executable that runs `python`.
func NewPipInstallProcess(executable executable.Executable) PipInstallProcess {
	return PipInstallProcess{
		executable: executable,
	}
}

// Execute installs the pip binary from source code located in the given srcPath into the a layer path designated by targetLayerPath.
func (p PipInstallProcess) Execute(srcPath, targetLayerPath string) error {
	buffer := bytes.NewBuffer(nil)

	err := p.executable.Execute(pexec.Execution{
		// Install pip from source with the pip that comes pre-installed with cpython
		Args: []string{"-m", "pip", "install", srcPath, "--user", "--no-index", fmt.Sprintf("--find-links=%s", srcPath)},
		// Set the PYTHONUSERBASE to ensure that pip is installed to the newly created target layer.
		Env:    append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath)),
		Stdout: buffer,
		Stderr: buffer,
	})
	if err != nil {
		return fmt.Errorf("failed to configure pip:\n%s\nerror: %w", buffer.String(), err)
	}

	return nil
}
