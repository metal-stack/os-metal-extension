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

ENSURE_GARDENER_MOD         := $(shell go get github.com/gardener/gardener@$$(go list -m -f "{{.Version}}" github.com/gardener/gardener))
GARDENER_HACK_DIR    		:= $(shell go list -m -f "{{.Dir}}" github.com/gardener/gardener)/hack
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

GOLANGCI_LINT_VERSION := v1.61.0
GO_VERSION := 1.23

ifeq ($(CI),true)
  DOCKER_TTY_ARG=""
else
  DOCKER_TTY_ARG=t
endif

CGO_ENABLED := 0
export GO111MODULE := on

### Build commands

TOOLS_DIR := $(HACK_DIR)/tools
include $(GARDENER_HACK_DIR)/tools.mk

.PHONY: all
all:
	go build -trimpath -tags netgo -o os-metal cmd/main.go
	strip os-metal

.PHONY: clean
clean:
	rm os-metal
	@$(shell find ./example -type f -name "controller-registration.yaml" -exec rm '{}' \;)
	@bash $(GARDENER_HACK_DIR)/clean.sh ./cmd/... ./pkg/...

.PHONY: install
install: tidy $(CONTROLLER_GEN) $(GEN_CRD_API_REFERENCE_DOCS) $(HELM) $(MOCKGEN)
	@LD_FLAGS="-w -X github.com/gardener/$(EXTENSION_PREFIX)-$(NAME)/pkg/version.Version=$(VERSION)" \
	bash $(GARDENER_HACK_DIR)/install.sh ./...

.PHONY: tidy
tidy:
	@GO111MODULE=on go mod tidy
	@mkdir -p $(REPO_ROOT)/.ci/hack && cp $(GARDENER_HACK_DIR)/.ci/* $(REPO_ROOT)/.ci/hack/ && chmod +xw $(REPO_ROOT)/.ci/hack/*

.PHONY: generate
generate: $(VGOPATH) $(HELM) $(YQ)
	@REPO_ROOT=$(REPO_ROOT) VGOPATH=$(VGOPATH) GARDENER_HACK_DIR=$(GARDENER_HACK_DIR) bash $(GARDENER_HACK_DIR)/generate-sequential.sh ./charts/... ./cmd/... ./pkg/...

.PHONE: generate-in-docker
generate-in-docker: tidy $(HELM) $(YQ)
	docker run --rm -it -v $(PWD):/go/src/github.com/metal-stack/os-metal-extension golang:$(GO_VERSION) \
		sh -c "cd /go/src/github.com/metal-stack/os-metal-extension \
				&& make generate \
				&& chown -R $(shell id -u):$(shell id -g) ."

.PHONY: check-generate
check-generate:
	@bash $(GARDENER_HACK_DIR)/check-generate.sh $(REPO_ROOT)

.PHONY: check
check: $(GOIMPORTS) $(GOLANGCI_LINT) $(HELM)
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check.sh --golangci-lint-config=./.golangci.yaml ./cmd/... ./pkg/...
	@REPO_ROOT=$(REPO_ROOT) bash $(GARDENER_HACK_DIR)/check-charts.sh ./charts

.PHONY: test
test:
	@bash $(GARDENER_HACK_DIR)/test.sh ./cmd/... ./pkg/...

.PHONY: test-in-docker
test-in-docker: tidy
	docker run --rm -i$(DOCKER_TTY_ARG) \
		--user $$(id -u):$$(id -g) \
		--mount type=tmpfs,destination=/.cache \
		--volume $(PWD):/go/src/github.com/metal-stack/os-metal-extension golang:$(GO_VERSION) \
			sh -c "cd /go/src/github.com/metal-stack/os-metal-extension \
					&& make install check test"

.PHONY: docker-image
docker-image:
	@docker build --build-arg VERIFY=$(VERIFY) -t $(IMAGE_PREFIX)/os-metal-extension:$(IMAGE_TAG) -f Dockerfile .

.PHONY: docker-push
docker-push:
	@docker push $(IMAGE_PREFIX)/os-metal-extension:$(IMAGE_TAG)
