// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uv

import (
	"fmt"
	"os"
	"path/filepath"
)

type UvInstallProcess struct {
}

// NewUvInstallProcess creates a UvInstallProcess instance.
func NewUvInstallProcess() UvInstallProcess {
	return UvInstallProcess{}
}

func (p UvInstallProcess) TranslateArchitecture(arch string) string {
	switch arch {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	default:
		return ""
	}
}

// Copy files from uv archive
func (p UvInstallProcess) Execute(targetLayerPath, sourcePath, dependencyArch string) error {
	arch := p.TranslateArchitecture(dependencyArch)

	if arch == "" {
		return fmt.Errorf("arch %s is not supported", dependencyArch)
	}

	folder := fmt.Sprintf(UvArchiveTemplateName, arch)
	return os.CopyFS(filepath.Join(targetLayerPath, "bin"), os.DirFS(filepath.Join(sourcePath, folder)))
}
