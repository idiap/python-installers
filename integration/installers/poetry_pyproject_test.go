// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func poetryTestPyProject(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		pack occam.Pack
	)

	it.Before(func() {
		pack = occam.NewPack()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "poetry", "non_poetry_app"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		it("does not build with non poetry pyproject.toml", func() {
			_, logs, err := pack.WithNoColor().Build.
				WithPullPolicy("never").
				WithBuildpacks(
					settings.Buildpacks.PythonInstallers.Online,
					// settings.Buildpacks.BuildPlan.Online,
				).
				Execute(name, source)
			Expect(err).To(HaveOccurred())
			Expect(logs).To(ContainLines("        this is not a poetry project"))
		})
	})
}
