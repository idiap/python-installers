// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/python-installers/pkg/executable/fakes"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"

	. "github.com/onsi/gomega"
)

func testPoetryInstallProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		version       string
		destLayerPath string
		pipLayerPath  string
		executable    *fakes.Executable

		poetryInstallProcess poetry.PoetryInstallProcess
	)

	it.Before(func() {
		var err error
		destLayerPath, err = os.MkdirTemp("", "poetry")
		Expect(err).NotTo(HaveOccurred())
		pipLayerPath, err = os.MkdirTemp("", "pip")
		Expect(err).NotTo(HaveOccurred())

		version = "1.2.3-some.version"

		executable = &fakes.Executable{}

		poetryInstallProcess = poetry.NewPoetryInstallProcess(executable)
	})

	context("Execute", func() {
		context("there is a poetry dependency to install", func() {
			it("installs it to the poetry layer", func() {
				err := poetryInstallProcess.Execute(version, destLayerPath, pipLayerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Env).To(Equal(append(os.Environ(),
					fmt.Sprintf("PYTHONPATH=%s", pipLayerPath),
					fmt.Sprintf("PYTHONUSERBASE=%s", destLayerPath)),
				))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"-m", "pip", "install", "poetry==1.2.3-some.version", "--user"}))
			})
		})

		context("failure cases", func() {
			context("the install process fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						_, err := fmt.Fprintln(execution.Stdout, "stdout output")
						Expect(err).NotTo(HaveOccurred())
						_, err = fmt.Fprintln(execution.Stderr, "stderr output")
						Expect(err).NotTo(HaveOccurred())
						return errors.New("installing poetry failed")
					}
				})

				it("returns an error", func() {
					err := poetryInstallProcess.Execute(version, destLayerPath, pipLayerPath)
					Expect(err).To(MatchError(ContainSubstring("installing poetry failed")))
					Expect(err).To(MatchError(ContainSubstring("stdout output")))
					Expect(err).To(MatchError(ContainSubstring("stderr output")))
				})
			})
		})
	})
}
