// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package pythoninstallers_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestUnitPythonInstallers(t *testing.T) {
	suite := spec.New("python-installers", spec.Report(report.Terminal{}), spec.Sequential())
	suite("Detect", testDetect)
	suite("Build", testBuild)
	suite.Run(t)
}
