// Copyright 2023 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operatingsystemconfig_test

import (
	"context"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	. "github.com/metal-stack/os-metal-extension/pkg/controller/operatingsystemconfig"
)

var _ = Describe("Actuator", func() {
	var (
		ctx        = context.TODO()
		log        = logr.Discard()
		fakeClient client.Client
		mgr        manager.Manager

		osc      *extensionsv1alpha1.OperatingSystemConfig
		actuator operatingsystemconfig.Actuator
	)

	BeforeEach(func() {
		fakeClient = fakeclient.NewClientBuilder().Build()
		mgr = test.FakeManager{Client: fakeClient}

		osc = &extensionsv1alpha1.OperatingSystemConfig{
			Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
				Purpose: extensionsv1alpha1.OperatingSystemConfigPurposeProvision,
				Units:   []extensionsv1alpha1.Unit{{Name: "some-unit.service", Content: ptr.To("foo")}},
				Files:   []extensionsv1alpha1.File{{Path: "/some/file", Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: "bar"}}}},
			},
		}
	})

	When("UseGardenerNodeAgent is true", func() {
		BeforeEach(func() {
			actuator = NewActuator(mgr)
		})

		When("purpose is 'provision'", func() {
			expectedUserData := `#!/bin/bash
# Fix mis-configuration of dockerd
mkdir -p /etc/docker
echo '{ "storage-driver": "devicemapper" }' > /etc/docker/daemon.json
sed -i '/Environment=DOCKER_SELINUX=--selinux-enabled=true/s/^/#/g' /run/systemd/system/docker.service

# Change existing worker to use docker registry-mirror
file="/etc/docker/daemon.json"
if [ $(jq -r 'has("registry-mirrors")' "${file}") == "false" ]; then
    contents=$(jq -M '. += {"registry-mirrors": ["https://mirror.gcr.io"]}' ${file})
    echo "${contents}" > ${file}
fi

systemctl daemon-reload
systemctl reload docker

mkdir -p "/some"

cat << EOF | base64 -d > "/some/file"
YmFy
EOF


cat << EOF | base64 -d > "/etc/systemd/system/some-unit.service"
Zm9v
EOF

systemctl enable 'some-unit.service' && systemctl restart --no-block 'some-unit.service'
`

			Describe("#Reconcile", func() {
				It("should not return an error", func() {
					userData, command, unitNames, fileNames, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(userData)).To(Equal(expectedUserData))
					Expect(command).To(BeNil())
					Expect(unitNames).To(BeEmpty())
					Expect(fileNames).To(BeEmpty())
					Expect(extensionUnits).To(BeEmpty())
					Expect(extensionFiles).To(BeEmpty())
				})
			})
		})

		When("purpose is 'reconcile'", func() {
			BeforeEach(func() {
				osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
			})

			Describe("#Reconcile", func() {
				It("should not return an error", func() {
					userData, command, unitNames, fileNames, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
					Expect(err).NotTo(HaveOccurred())

					Expect(userData).NotTo(BeEmpty()) // legacy logic is tested in ./generator/generator_test.go
					Expect(command).To(BeNil())
					Expect(unitNames).To(ConsistOf("some-unit.service"))
					Expect(fileNames).To(ConsistOf("/some/file"))
					Expect(extensionUnits).To(ConsistOf(
						extensionsv1alpha1.Unit{
							Name: "containerd.service",
							DropIns: []extensionsv1alpha1.DropIn{{
								Name: "11-exec_config.conf",
								Content: `[Service]
ExecStart=
ExecStart=/etc/containerd/config.toml
`,
							}},
							FilePaths: []string{"/etc/containerd/config.toml"},
						},
					))
					Expect(extensionFiles).To(ConsistOf(
						extensionsv1alpha1.File{
							Path:        "/etc/containerd/config.toml",
							Permissions: ptr.To(int32(0644)),
							Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `
# created by os-extension-metal
[plugins.cri.registry.mirrors]
  [plugins.cri.registry.mirrors."docker.io"]
    endpoint = ["https://mirror.gcr.io"]
`}},
						},
					))
				})
			})
		})
	})
})
