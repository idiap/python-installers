// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package miniconda_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	pythoninstallers "github.com/paketo-buildpacks/python-installers/pkg/installers/common"

	"github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect        packit.DetectFunc
		detectContext packit.DetectContext
	)

	it.Before(func() {
		detect = miniconda.Detect()
		detectContext = packit.DetectContext{
			WorkingDir: "/working-dir",
		}
	})

	context("detection", func() {
		it("returns a plan that provides conda", func() {
			result, err := detect(detectContext)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.DetectResult{
				Plan: packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: "conda"},
					},
				},
			}))
		})
	})

	context("when BP_MINICONDA_VERSION is set", func() {
		it.Before(func() {
			t.Setenv("BP_MINICONDA_VERSION", "some-version")
		})

		it("returns a build plan that provides the version of conda from BP_MINICONDA_VERSION", func() {
			result, err := detect(detectContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: "conda"},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: "conda",
						Metadata: pythoninstallers.BuildPlanMetadata{
							Version:       "some-version",
							VersionSource: "BP_MINICONDA_VERSION",
						},
					},
				},
			}))
		})
	})
}
