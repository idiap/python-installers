// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package pip_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPip(t *testing.T) {
	suite := spec.New("pip", spec.Report(report.Terminal{}))
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite("InstallProcess", testPipInstallProcess)
	suite("SiteProcess", testSiteProcess)
	suite.Run(t)
}
