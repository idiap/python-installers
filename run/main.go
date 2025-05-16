// SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
// SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/cargo"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/pexec"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/scribe"
	pythoninstallers "github.com/paketo-buildpacks/python-installers"
	pkgcommon "github.com/paketo-buildpacks/python-installers/pkg/installers/common"
	miniconda "github.com/paketo-buildpacks/python-installers/pkg/installers/miniconda"
	pip "github.com/paketo-buildpacks/python-installers/pkg/installers/pip"
	pipenv "github.com/paketo-buildpacks/python-installers/pkg/installers/pipenv"
	poetry "github.com/paketo-buildpacks/python-installers/pkg/installers/poetry"
)

func main() {
	logger := scribe.NewEmitter(os.Stdout).WithLevel(os.Getenv("BP_LOG_LEVEL"))

	buildParameters := pkgcommon.CommonBuildParameters{
		SbomGenerator: pkgcommon.Generator{},
		Clock:         chronos.DefaultClock,
		Logger:        logger,
	}

	packagerParameters := map[string]pythoninstallers.PackagerParameters{
		miniconda.Conda: miniconda.CondaBuildParameters{
			DependencyManager: postal.NewService(cargo.NewTransport()),
			Runner:            miniconda.NewScriptRunner(pexec.NewExecutable("bash")),
		},
		pip.Pip: pip.PipBuildParameters{
			DependencyManager:  postal.NewService(cargo.NewTransport()),
			InstallProcess:     pip.NewPipInstallProcess(pexec.NewExecutable("python")),
			SitePackageProcess: pip.NewSiteProcess(pexec.NewExecutable("python")),
		},
		pipenv.Pipenv: pipenv.PipEnvBuildParameters{
			DependencyManager:  postal.NewService(cargo.NewTransport()),
			InstallProcess:     pipenv.NewPipenvInstallProcess(pexec.NewExecutable("pip")),
			SitePackageProcess: pipenv.NewSiteProcess(pexec.NewExecutable("python")),
		},
		poetry.PoetryDependency: poetry.PoetryBuildParameters{
			DependencyManager:  postal.NewService(cargo.NewTransport()),
			InstallProcess:     poetry.NewPoetryInstallProcess(pexec.NewExecutable("python")),
			SitePackageProcess: poetry.NewSiteProcess(pexec.NewExecutable("python")),
		},
	}

	packit.Run(
		pythoninstallers.Detect(logger, poetry.NewPyProjectParser()),
		pythoninstallers.Build(logger, buildParameters, packagerParameters),
	)
}
