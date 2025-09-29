# Copyright (c) 2013-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

.PHONY: retrieve test

retrieve:
	@cd retrieval; \
	go build -o retrieve; \
	./retrieve \
	    --buildpack_toml_path=$(buildpackTomlPath) \
		--output=$(output); \
	rm retrieve

test:
	./test/test.sh \
		--tarballPath $(tarballPath) \
		--expectedVersion $(version)
