#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# setup virtual GOPATH
source "$GARDENER_HACK_DIR"/vgopath-setup.sh

CODE_GEN_DIR=$(go list -m -f '{{.Dir}}' k8s.io/code-generator)

# We need to explicitly pass GO111MODULE=off to k8s.io/code-generator as it is significantly slower otherwise,
# see https://github.com/kubernetes/code-generator/issues/100.
export GO111MODULE=off

rm -f $GOPATH/bin/*-gen

PROJECT_ROOT=$(dirname $0)/..

bash "${CODE_GEN_DIR}/generate-internal-groups.sh" \
  deepcopy,defaulter \
  github.com/metal-stack/os-metal-extension/pkg/client \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  "metal:v1alpha1" \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"

bash "${CODE_GEN_DIR}/generate-internal-groups.sh" \
  conversion \
  github.com/metal-stack/os-metal-extension/pkg/client \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  "metal:v1alpha1" \
  --extra-peer-dirs=github.com/metal-stack/os-metal-extension/pkg/apis/metal,github.com/metal-stack/os-metal-extension/pkg/apis/metal/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"

bash "${CODE_GEN_DIR}/generate-internal-groups.sh" \
  deepcopy,defaulter \
  github.com/metal-stack/os-metal-extension/pkg/client/componentconfig \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  "config:v1alpha1" \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"

bash "${CODE_GEN_DIR}/generate-internal-groups.sh" \
  conversion \
  github.com/metal-stack/os-metal-extension/pkg/client/componentconfig \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  github.com/metal-stack/os-metal-extension/pkg/apis \
  "config:v1alpha1" \
  --extra-peer-dirs=github.com/metal-stack/os-metal-extension/pkg/apis/config,github.com/metal-stack/os-metal-extension/pkg/apis/config/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/conversion,k8s.io/apimachinery/pkg/runtime,github.com/gardener/gardener/extensions/pkg/controller/healthcheck/config/v1alpha1 \
  --go-header-file "${PROJECT_ROOT}/vendor/github.com/gardener/gardener/hack/LICENSE_BOILERPLATE.txt"
