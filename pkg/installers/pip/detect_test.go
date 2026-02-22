// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pip_test

import (
	"testing"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"

	"github.com/paketo-buildpacks/python-installers/pkg/build"
	"github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		detect        packit.DetectFunc
		detectContext packit.DetectContext
	)

	it.Before(func() {
		detect = pip.Detect()
		detectContext = packit.DetectContext{}
	})

	context("detection", func() {
		it("returns a build plan that provides pip", func() {
			result, err := detect(detectContext)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Plan).To(Equal(packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: pip.Pip},
				},
				Requires: []packit.BuildPlanRequirement{
					{
						Name: pip.CPython,
						Metadata: build.BuildPlanMetadata{
							Build: true,
						},
					},
				},
			}))
		})

		context("when BP_PIP_VERSION is set", func() {
			it.Before(func() {
				t.Setenv(pip.EnvVersion, "some-version")
			})

			it("returns a build plan that provides the version of pip from BP_PIP_VERSION", func() {
				result, err := detect(detectContext)
				Expect(err).NotTo(HaveOccurred())

				Expect(result.Plan).To(Equal(packit.BuildPlan{
					Provides: []packit.BuildPlanProvision{
						{Name: pip.Pip},
					},
					Requires: []packit.BuildPlanRequirement{
						{
							Name: pip.CPython,
							Metadata: build.BuildPlanMetadata{
								Build: true,
							},
						},
						{
							Name: pip.Pip,
							Metadata: build.BuildPlanMetadata{
								Version:       "some-version",
								VersionSource: pip.EnvVersion,
							},
						},
					},
				}))
			})

			context("when the provided version is of the form X.Y", func() {
				it.Before(func() {
					t.Setenv(pip.EnvVersion, "2.11")
				})

				it("selects the version X.Y.0", func() {
					result, err := detect(detectContext)

					Expect(err).NotTo(HaveOccurred())

					Expect(result.Plan).To(Equal(packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: pip.Pip},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: pip.CPython,
								Metadata: build.BuildPlanMetadata{
									Build: true,
								},
							},
							{
								Name: pip.Pip,
								Metadata: build.BuildPlanMetadata{
									Version:       "2.11.0",
									VersionSource: pip.EnvVersion,
								},
							},
						},
					}))
				})
			})

			context("when the provided version is of the form X.Y.Z", func() {
				it.Before(func() {
					t.Setenv(pip.EnvVersion, "22.1.3")
				})

				it("selects the exact provided version X.Y.Z", func() {
					result, err := detect(detectContext)

					Expect(err).NotTo(HaveOccurred())

					Expect(result.Plan).To(Equal(packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: pip.Pip},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: pip.CPython,
								Metadata: build.BuildPlanMetadata{
									Build: true,
								},
							},
							{
								Name: pip.Pip,
								Metadata: build.BuildPlanMetadata{
									Version:       "22.1.3",
									VersionSource: pip.EnvVersion,
								},
							},
						},
					}))
				})
			})

			context("when the provided version is of some other form", func() {
				it.Before(func() {
					t.Setenv(pip.EnvVersion, "some.other")
				})

				it("selects the exact provided version", func() {
					result, err := detect(detectContext)

					Expect(err).NotTo(HaveOccurred())

					Expect(result.Plan).To(Equal(packit.BuildPlan{
						Provides: []packit.BuildPlanProvision{
							{Name: pip.Pip},
						},
						Requires: []packit.BuildPlanRequirement{
							{
								Name: pip.CPython,
								Metadata: build.BuildPlanMetadata{
									Build: true,
								},
							},
							{
								Name: pip.Pip,
								Metadata: build.BuildPlanMetadata{
									Version:       "some.other",
									VersionSource: pip.EnvVersion,
								},
							},
						},
					}))
				})
			})
		})

	})
}
