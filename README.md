<!--
SPDX-FileCopyrightText: © 2025 Idiap Research Institute <contact@idiap.ch>
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Python Installers Cloud Native Buildpack

The Paketo Buildpack for Python Installers is a Cloud Native Buildpack that
installs python package managers.

The buildpack is published for consumption at
`gcr.io/paketo-buildpacks/python-installers` and
`paketobuildpacks/python-installers`.

## Behavior
This buildpack participates if one of the following detection succeeds:

- (miniconda)[installers/pkg/minconda/README.md] -> Always
- (pip)[installers/ppkg/pip/README.md] -> Always
- (pipenv)[installers/ppkg/pipenv/README.md] -> Always
- (poetry)[installers/ppkg/poetry/README.md] -> `pyproject.toml` is present in the root folder

The buildpack will do the following:
* At build time:
  - Installs the package manager
  - Makes it available on the `PATH`
  - Adjusts `PYTHONPATH` as required
* At run time:
  - Does nothing

## Usage

To package this buildpack for consumption:
```
$ ./scripts/package.sh --version x.x.x
```
This will create a `buildpackage.cnb` file under the build directory which you
can use to build your app as follows: `pack build <app-name> -p <path-to-app>
-b <cpython buildpack> -b <pip buildpack> -b build/buildpackage.cnb -b
<other-buildpacks..>`.

To run the unit and integration tests for this buildpack:
```
$ ./scripts/unit.sh && ./scripts/integration.sh
```
