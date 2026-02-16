// SPDX-FileCopyrightText: © 2026 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pixi

import (
	"fmt"
	"os"
	"path/filepath"
)

type PixiInstallProcess struct {
}

// NewPixiInstallProcess creates a PixiInstallProcess instance.
func NewPixiInstallProcess() PixiInstallProcess {
	return PixiInstallProcess{}
}

func (p PixiInstallProcess) TranslateArchitecture(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return ""
	}
}

// Copy files from pixi archive
func (p PixiInstallProcess) Execute(targetLayerPath, sourcePath, dependencyArch string) error {
	arch := p.TranslateArchitecture(dependencyArch)

	if arch == "" {
		return fmt.Errorf("arch %s is not supported", dependencyArch)
	}

	return os.CopyFS(filepath.Join(targetLayerPath, "bin"), os.DirFS(sourcePath))
}
