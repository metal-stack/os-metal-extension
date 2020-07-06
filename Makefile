# Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

REGISTRY                    := docker.io
IMAGE_PREFIX                := $(REGISTRY)/metalstack
IMAGE_TAG                   := $(or ${GITHUB_TAG_NAME}, latest)
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
HOSTNAME                    := $(shell hostname)
VERSION                     := $(shell bash -c 'source $(HACK_DIR)/common.sh && echo $$VERSION')
LD_FLAGS                    := "-X 'github.com/metal-pod/v.Version=$(VERSION)' \
								-X 'github.com/metal-pod/v.Revision=$(GITVERSION)' \
								-X 'github.com/metal-pod/v.GitSHA1=$(SHA)' \
								-X 'github.com/metal-pod/v.BuildDate=$(BUILDDATE)'"
VERIFY                      := true
LEADER_ELECTION             := false
IGNORE_OPERATION_ANNOTATION := false

export CGO_ENABLED := 0
export GO111MODULE := on

### Build commands

.PHONY: format
format:
	@./hack/format.sh

.PHONY: clean
clean:
	@./hack/clean.sh
	rm os-metal

.PHONY: generate
generate:
	@./hack/generate.sh

.PHONE: generate-in-docker
generate-in-docker:
	docker run --rm -it -v $(PWD):/go/src/github.com/metal-stack/os-metal-extension golang:1.14 \
		sh -c "cd /go/src/github.com/metal-stack/os-metal-extension \
				&& ./hack/install-requirements.sh \
				&& make generate \
				&& chown -R $(shell id -u):$(shell id -g) ."

.PHONY: check
check:
	@./hack/check.sh

.PHONY: test
test:
	@./hack/test.sh

.PHONY: verify
verify: check generate test format

.PHONY: install
install:
	@./hack/install.sh

.PHONY: all
all: generate
	go build -trimpath -tags netgo -o os-metal cmd/main.go
	strip os-metal

.PHONY: docker-image
docker-image:
	@docker build --build-arg VERIFY=$(VERIFY) -t $(IMAGE_PREFIX)/os-metal-extension:$(IMAGE_TAG) -f Dockerfile .

.PHONY: docker-push
docker-push:
	@docker push $(IMAGE_PREFIX)/os-metal-extension:$(IMAGE_TAG)

### Debug / Development commands
.PHONY: start-os-metal
start-os-metal:
	@LEADER_ELECTION_NAMESPACE=garden go run \
		-ldflags $(LD_FLAGS) \
		./cmd \
		--leader-election=$(LEADER_ELECTION)
