// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/fs"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	pythoninstallers "github.com/paketo-buildpacks/python-installers"
	pkgcommon "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
	// conda "github.com/paketo-buildpacks/python-installers/pkg/installers/conda"
	// pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	// pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	// poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	buildParameters := pkgcommon.CommonBuildParameters{
		SbomGenerator: pkgcommon.Generator{},
		Clock:         chronos.DefaultClock,
		Logger:        logger,
	}

	// packagerParameters := map[string]pythoninstallers.PackagerParameters{
	// 	conda.CondaEnvPlanEntry: conda.CondaBuildParameters{
	// 		Runner: conda.NewCondaRunner(pexec.NewExecutable("conda"), fs.NewChecksumCalculator(), logger),
	// 	},
	// 	pip.Manager: pip.PipBuildParameters{
	// 		InstallProcess:      pip.NewpipProcess(pexec.NewExecutable("pip"), logger),
	// 		SitePackagesProcess: pip.NewSiteProcess(pexec.NewExecutable("python")),
	// 	},
	// 	pipenv.Manager: pipenv.PipEnvBuildParameters{
	// 		InstallProcess: pipenv.NewpipenvProcess(pexec.NewExecutable("pipenv"), logger),
	// 		SiteProcess:    pipenv.NewSiteProcess(pexec.NewExecutable("python")),
	// 		VenvDirLocator: pipenv.NewVenvLocator(),
	// 	},
	// 	poetry.PoetryVenv: poetry.PoetryEnvBuildParameters{
	// 		EntryResolver:           draft.NewPlanner(),
	// 		InstallProcess:          poetry.NewpoetryProcess(pexec.NewExecutable("poetry"), logger),
	// 		PythonPathLookupProcess: poetry.NewPythonPathProcess(),
	// 	},
	// }

	packit.Run(
		pythoninstallers.Detect(logger),
		// pythoninstallers.Build(logger, buildParameters, packagerParameters),
	)
}
