// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package miniconda_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda/fakes"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testScriptRunner(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		scriptDir  string
		scriptPath string

		executable *fakes.Executable

		scriptRunner miniconda.ScriptRunner
	)

	it.Before(func() {
		var err error
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		scriptDir, err = os.MkdirTemp("", "miniconda-script-dir")
		Expect(err).NotTo(HaveOccurred())

		scriptPath = filepath.Join(scriptDir, "artifact")
		err = os.WriteFile(scriptPath, nil, 0644)
		Expect(err).NotTo(HaveOccurred())

		executable = &fakes.Executable{}

		scriptRunner = miniconda.NewScriptRunner(executable)
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(scriptDir)).To(Succeed())
	})

	context("Run", func() {
		it("runs the miniconda install script with the correct flags", func() {
			err := scriptRunner.Run(scriptPath, layersDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{
				filepath.Join(scriptDir, "artifact"),
				"-b",
				"-f",
				"-p", layersDir,
			}))
		})

		context("failure cases", func() {
			context("when the script fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Returns.Error = errors.New("script failed to run")
				})

				it("returns an error", func() {
					err := scriptRunner.Run(scriptPath, layersDir)
					Expect(err).To(MatchError("failed while running miniconda install script: script failed to run"))
				})
			})
		})
	})
}
