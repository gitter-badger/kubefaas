# Copyright 2017 The Fission Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.DEFAULT_GOAL := check
SHELL := /bin/bash

check: static-check build clean

static-check:
	hack/verify-gofmt.sh
	hack/verify-govet.sh
	hack/verify-staticcheck.sh
	hack/runtests.sh
	@rm -f coverage.txt

# run basic check scripts
run-test:
	source test/init_tools.sh && \
	source test/test_utils.sh && \
	dump_system_info && \
	set_local_build_and_deploy_env && \
	build_images_and_cli && \
	helm_remove_old_releases && \
	install && \
	set_local_test_env && \
	run_test_suites && \
	check_result && \
	cleanup

run-full-test:
	source test/init_tools.sh && \
	source test/test_utils.sh && \
	dump_system_info && \
	set_local_build_and_deploy_env && \
	build_images_and_cli && \
	build_env_images && \
	helm_remove_old_releases && \
	install && \
	set_local_test_env && \
	run_test_suites && \
	check_result && \
	cleanup

# ensure the changes are buildable
build:
	go build -o cmd/bundle/bundle ./cmd/bundle/
	go build -o cmd/cli/kubefaas ./cmd/cli/
	go build -o cmd/fetcher/fetcher ./cmd/fetcher/
	go build -o cmd/fetcher/builder ./cmd/builder/

# install CLI binary to $PATH
install: build
	mv cmd/cli/kubefaas $(GOPATH)/bin

# build images (environment images are not included)
image:
	docker build -t bundle -f cmd/bundle/Dockerfile .
	docker build -t fetcher -f cmd/fetcher/Dockerfile .
	docker build -t builder -f cmd/builder/Dockerfile .

clean:
	@rm -f cmd/bundle/bundle
	@rm -f cmd/cli/kubefaas
	@rm -f cmd/fetcher/fetcher
	@rm -f cmd/fetcher/builder
