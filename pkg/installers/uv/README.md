<!--
// SPDX-FileCopyrightText: Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
SPDX-FileContributor: Samuel Gaist <samuel.gaist@idiap.ch>

SPDX-License-Identifier: Apache-2.0
-->

# Sub-package for uv installation

This sub-package installs uv into a layer and makes it available on the
PATH.

## Integration

The uv CNB provides uv as a dependency. Downstream buildpacks can
require the uv dependency by generating a [Build Plan
TOML](https://github.com/buildpacks/spec/blob/master/buildpack.md#build-plan-toml)
file that looks like the following:

```toml
[[requires]]

  # The name of the Uv dependency is "uv". This value is considered
  # part of the public API for the buildpack and will not change without a plan
  # for deprecation.
  name = "uv"

  # The version of the uv dependency is not required. In the case it
  # is not specified, the buildpack will provide the default version, which can
  # be seen in the buildpack.toml file.
  # If you wish to request a specific version, the buildpack supports
  # specifying a semver constraint in the form of "0.*", "0.9.*", or even
  # "0.9.22".
  version = "0.9.22"

  # The Miniconda buildpack supports some non-required metadata options.
  [requires.metadata]

    # Setting the build flag to true will ensure that the uv
    # dependency is available on the $PATH for subsequent buildpacks during
    # their build phase. If you are writing a buildpack that needs to run
    # uv during its build process, this flag should be set to true.
    build = true

    # Setting the launch flag to true will ensure that the uv
    # dependency is available on the $PATH for the running application. If you are
    # writing an application that needs to run uv at runtime, this flag
    # should be set to true.
    launch = true
```
