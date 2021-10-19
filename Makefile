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

REGISTRY                    := ghcr.io
IMAGE_PREFIX                := $(REGISTRY)/metal-stack
IMAGE_TAG                   := $(or ${GITHUB_TAG_NAME}, latest)
REPO_ROOT                   := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
HACK_DIR                    := $(REPO_ROOT)/hack
HOSTNAME                    := $(shell hostname)
VERSION                     := $(shell cat "$(REPO_ROOT)/VERSION")
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


.PHONY: all
all:
	go build -trimpath -tags netgo -o os-metal cmd/main.go
	strip os-metal

.PHONY: clean
clean:
	rm os-metal

.PHONY: revendor
revendor:
	go mod vendor
	go mod tidy
	chmod +x vendor/github.com/gardener/gardener/hack/*.sh

.PHONY: generate
generate:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/generate.sh ./charts/... ./cmd/... ./pkg/...

.PHONE: generate-in-docker
generate-in-docker:
	docker run --rm -it -v $(PWD):/go/src/github.com/metal-stack/os-metal-extension golang:1.16 \
		sh -c "cd /go/src/github.com/metal-stack/os-metal-extension \
				&& make revendor install-requirements generate \
				&& chown -R $(shell id -u):$(shell id -g) ."

.PHONY: install-requirements
install-requirements:
	@$(REPO_ROOT)/vendor/github.com/gardener/gardener/hack/install-requirements.sh

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
