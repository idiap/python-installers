// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

package poetry_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnit(t *testing.T) {
	suite := spec.New("poetry", spec.Report(report.Terminal{}), spec.Parallel())
	suite("Build", testBuild)
	suite("Detect", testDetect, spec.Sequential())
	suite("InstallProcess", testPoetryInstallProcess)
	suite("SiteProcess", testSiteProcess)
	suite("PyProjectParse", testPyProjectParser)
	suite.Run(t)
}
