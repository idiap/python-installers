// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/common/executable/fakes"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"

	. "github.com/onsi/gomega"
)

func testSiteProcess(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		targetLayerPath string
		executable      *fakes.Executable

		siteProcess poetry.SiteProcess
	)

	it.Before(func() {
		var err error
		targetLayerPath, err = os.MkdirTemp("", "poetry")
		Expect(err).NotTo(HaveOccurred())

		executable = &fakes.Executable{}
		executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
			if execution.Stdout != nil {
				_, err := fmt.Fprint(execution.Stdout, targetLayerPath, "/poetry/lib/python/site-packages")
				Expect(err).NotTo(HaveOccurred())
			}
			return nil
		}

		siteProcess = poetry.NewSiteProcess(executable)
	})

	it.After(func() {
		Expect(os.RemoveAll(targetLayerPath)).To(Succeed())
	})

	context("Execute", func() {
		context("there are site packages in the poetry layer", func() {
			it("returns the full path to the packages", func() {
				sitePackagesPath, err := siteProcess.Execute(targetLayerPath)
				Expect(err).NotTo(HaveOccurred())

				Expect(executable.ExecuteCall.Receives.Execution.Env).To(Equal(append(os.Environ(), fmt.Sprintf("PYTHONUSERBASE=%s", targetLayerPath))))
				Expect(executable.ExecuteCall.Receives.Execution.Args).To(Equal([]string{"-m", "site", "--user-site"}))

				Expect(sitePackagesPath).To(Equal(filepath.Join(targetLayerPath, "poetry", "lib", "python", "site-packages")))
			})
		})

		context("failure cases", func() {
			context("site package lookup fails", func() {
				it.Before(func() {
					executable.ExecuteCall.Stub = func(execution pexec.Execution) error {
						_, err := fmt.Fprintln(execution.Stdout, "stdout output")
						Expect(err).NotTo(HaveOccurred())
						_, err = fmt.Fprintln(execution.Stderr, "stderr output")
						Expect(err).NotTo(HaveOccurred())
						return errors.New("locating site packages failed")
					}
				})

				it("returns an error", func() {
					_, err := siteProcess.Execute(targetLayerPath)
					Expect(err).To(MatchError(ContainSubstring("failed to locate site packages:")))
					Expect(err).To(MatchError(ContainSubstring("stderr output")))
					Expect(err).To(MatchError(ContainSubstring("error: locating site packages failed")))
				})
			})
		})
	})
}
