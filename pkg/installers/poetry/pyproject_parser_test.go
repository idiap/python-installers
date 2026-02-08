// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
)

func testPyProjectParser(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir    string
		pyProjectToml string

		parser poetry.PyProjectParser
	)

	const (
		version = `[tool.poetry.dependencies]
python = "1.2.3"`
		version_pep621 = `[project]
requires-python = ">=1.2.3"`
		exact_version_pep621 = `[project]
requires-python = "==1.2.3"`
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		pyProjectToml = filepath.Join(workingDir, poetry.PyProjectTomlFile)

		parser = poetry.NewPyProjectParser()
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Calling ParsePythonVersion", func() {
		it("parses version", func() {
			Expect(os.WriteFile(pyProjectToml, []byte(version), os.ModePerm)).To(Succeed())

			version, err := parser.ParsePythonVersion(pyProjectToml)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.2.3"))
		})

		it("parses version PEP621", func() {
			Expect(os.WriteFile(pyProjectToml, []byte(version_pep621), os.ModePerm)).To(Succeed())

			version, err := parser.ParsePythonVersion(pyProjectToml)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(">=1.2.3"))
		})

		it("parses exact version PEP621", func() {
			Expect(os.WriteFile(pyProjectToml, []byte(exact_version_pep621), os.ModePerm)).To(Succeed())

			version, err := parser.ParsePythonVersion(pyProjectToml)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal("1.2.3"))
		})

		it("returns empty string if file does not contain 'tool.poetry.dependencies.python' or project.requires-python", func() {
			Expect(os.WriteFile(pyProjectToml, []byte(""), os.ModePerm)).To(Succeed())

			version, err := parser.ParsePythonVersion(pyProjectToml)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(""))
		})

		context("error handling", func() {
			it("fails if file does not exist", func() {
				_, err := parser.ParsePythonVersion("not-a-valid-dir")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	context("Calling IsPoetryProject", func() {
		it("returns true on file without build-system entry", func() {
			Expect(os.WriteFile(pyProjectToml, []byte(""), os.ModePerm)).To(Succeed())

			isPoetryProject, err := parser.IsPoetryProject(pyProjectToml)
			Expect(err).NotTo(HaveOccurred())
			Expect(isPoetryProject).To(BeTrue())
		})

		it("returns true on file with poetry build-system entry", func() {
			content := []byte(`
				[build-system]
				requires = ["poetry-core>=1.0.0"]
				build-backend = "poetry.core.masonry.api"
				`)
			Expect(os.WriteFile(pyProjectToml, content, os.ModePerm)).To(Succeed())

			isPoetryProject, err := parser.IsPoetryProject(pyProjectToml)
			Expect(err).NotTo(HaveOccurred())
			Expect(isPoetryProject).To(BeTrue())
		})

		it("returns false on file with other build-system entry", func() {
			content := []byte(`
				[build-system]
				requires = ["setuptools", "setuptools-scm"]
				build-backend = "setuptools.build_meta"
				`)
			Expect(os.WriteFile(pyProjectToml, content, os.ModePerm)).To(Succeed())

			isPoetryProject, err := parser.IsPoetryProject(pyProjectToml)
			Expect(err).NotTo(HaveOccurred())
			Expect(isPoetryProject).To(BeFalse())
		})
	})
}
