// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package uv_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/uv"
)

func testUvInstallProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
		srcDir     string
		dstDir     string

		installProcess uv.UvInstallProcess
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		srcDir = filepath.Join(workingDir, "src")
		Expect(os.MkdirAll(srcDir, os.ModePerm)).To(Succeed())

		dstDir = filepath.Join(workingDir, "dst")
		Expect(os.MkdirAll(dstDir, os.ModePerm)).To(Succeed())

		installProcess = uv.NewUvInstallProcess()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Calling Execute", func() {
		archs := []string{"amd64", "arm64"}
		for _, arch := range archs {
			context(fmt.Sprintf("Test for %s", arch), func() {
				it.Before(func() {
					archFolder := filepath.Join(srcDir, fmt.Sprintf(uv.UvArchiveTemplateName, installProcess.TranslateArchitecture(arch)))
					Expect(os.MkdirAll(archFolder, os.ModePerm)).To(Succeed())
				})
				it(fmt.Sprintf("copies file for %s", arch), func() {
					err := installProcess.Execute(dstDir, srcDir, arch)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		}

		context("error handling", func() {
			it("fails if arch is unknown", func() {
				err := installProcess.Execute("dummy", "dummy", "invalid")
				Expect(err).To(HaveOccurred())
			})
		})
	})
}
