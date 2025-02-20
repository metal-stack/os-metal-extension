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
	_ "embed"
	"encoding/json"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/go-logr/logr"
	metalextensionv1alpha1 "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/v1alpha1"
	. "github.com/metal-stack/os-metal-extension/pkg/controller/operatingsystemconfig"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var _ = Describe("Actuator", func() {
	var (
		ctx        = context.TODO()
		log        = logr.Discard()
		fakeClient client.Client
		mgr        manager.Manager

		osc                           *extensionsv1alpha1.OperatingSystemConfig
		isolatedClusterProviderConfig *runtime.RawExtension
		actuator                      operatingsystemconfig.Actuator
	)

	BeforeEach(func() {
		fakeClient = fakeclient.NewClientBuilder().Build()
		mgr = test.FakeManager{Client: fakeClient}

		osc = &extensionsv1alpha1.OperatingSystemConfig{
			Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
				CRIConfig: &extensionsv1alpha1.CRIConfig{
					Name: "containerd",
				},
				Purpose: extensionsv1alpha1.OperatingSystemConfigPurposeProvision,
				Units:   []extensionsv1alpha1.Unit{{Name: "some-unit.service", Content: ptr.To("foo")}},
				Files:   []extensionsv1alpha1.File{{Path: "/some/file", Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: "bar"}}}},
			},
		}
		isolatedClusterProviderConfig = &runtime.RawExtension{
			Raw: mustMarshal(&metalextensionv1alpha1.ImageProviderConfig{
				NetworkIsolation: &metalextensionv1alpha1.NetworkIsolation{
					AllowedNetworks: metalextensionv1alpha1.AllowedNetworks{
						Ingress: []string{"10.0.0.1/24"},
						Egress:  []string{"100.0.0.1/24"},
					},
					DNSServers: []string{"1.1.1.1", "1.0.0.1"},
					NTPServers: []string{"134.60.1.27", "134.60.111.110"},
					RegistryMirrors: []metalextensionv1alpha1.RegistryMirror{
						{
							Name:     "metal-stack registry",
							Endpoint: "https://r.metal-stack.dev",
							IP:       "1.2.3.4",
							Port:     443,
							MirrorOf: []string{
								"ghcr.io",
								"quay.io",
							},
						},
						{
							Name:     "local registry",
							Endpoint: "http://localhost:8080",
							IP:       "127.0.0.1",
							Port:     8080,
							MirrorOf: []string{
								"docker.io",
							},
						},
					},
				},
			}),
		}
	})

	BeforeEach(func() {
		actuator = NewActuator(mgr)
	})

	Describe("#Reconcile", func() {
		When("purpose is 'provision'", func() {
			BeforeEach(func() {
				osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeProvision
			})

			It("should not return an error", func() {
				userData, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(userData)).To(ContainSubstring("/etc/containerd/config.toml"))
				Expect(string(userData)).To(HavePrefix("{")) // check we have ignition format
				Expect(string(userData)).To(HaveSuffix("}")) // check we have ignition format
				Expect(extensionUnits).To(BeEmpty())
				Expect(extensionFiles).To(BeEmpty())
			})

			It("network isolation files are added", func() {
				osc = osc.DeepCopy()
				osc.Spec.ProviderConfig = isolatedClusterProviderConfig

				userData, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(userData)).To(ContainSubstring("/etc/containerd/config.toml"))
				Expect(string(userData)).To(ContainSubstring("/etc/resolv.conf"))
				Expect(string(userData)).To(HavePrefix("{")) // check we have ignition format
				Expect(string(userData)).To(HaveSuffix("}")) // check we have ignition format
				Expect(extensionUnits).To(BeEmpty())
				Expect(extensionFiles).To(BeEmpty())
			})
		})

		When("purpose is 'reconcile'", func() {
			BeforeEach(func() {
				osc.Spec.Purpose = extensionsv1alpha1.OperatingSystemConfigPurposeReconcile
			})

			It("should not return an error", func() {
				userData, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())

				Expect(userData).To(BeEmpty())
				Expect(extensionUnits).To(BeNil())
				Expect(extensionFiles).To(ConsistOf(extensionsv1alpha1.File{
					Path:        "/etc/containerd/config.toml",
					Permissions: ptr.To(int32(420)),
					Content: extensionsv1alpha1.FileContent{
						Inline: &extensionsv1alpha1.FileContentInline{
							Encoding: string(extensionsv1alpha1.PlainFileCodecID),
							Data: `# Generated by os-extension-metal
version = 2
imports = ["/etc/containerd/conf.d/*.toml"]
disabled_plugins = []

[plugins."io.containerd.grpc.v1.cri".registry]
  config_path = "/etc/containerd/certs.d"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
  runtime_type = "io.containerd.runc.v2"
  `,
						},
					},
				}))
			})

			It("network isolation files are added", func() {
				osc = osc.DeepCopy()
				osc.Spec.ProviderConfig = isolatedClusterProviderConfig

				userData, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(userData)).To(BeEmpty())
				Expect(extensionUnits).To(BeEmpty())
				Expect(extensionFiles).To(ConsistOf(
					extensionsv1alpha1.File{
						Path: "/etc/systemd/resolved.conf.d/dns.conf",
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								Data: `# Generated by os-extension-metal
[Resolve]
DNS=1.1.1.1 1.0.0.1
Domain=~.
`,
							},
						},
					},
					extensionsv1alpha1.File{
						Path: "/etc/resolv.conf",
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								Data: `# Generated by os-extension-metal
nameserver 1.1.1.1
nameserver 1.0.0.1
`,
							},
						},
					},
					extensionsv1alpha1.File{
						Path:        "/etc/systemd/timesyncd.conf",
						Permissions: ptr.To(int32(0644)),
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								Data: `# Generated by os-extension-metal
[Time]
NTP=134.60.1.27 134.60.111.110
`,
							},
						},
					},
					extensionsv1alpha1.File{
						Path:        "/etc/containerd/config.toml",
						Permissions: ptr.To(int32(420)),
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								Data: `# Generated by os-extension-metal
version = 2
imports = ["/etc/containerd/conf.d/*.toml"]
disabled_plugins = []

[plugins."io.containerd.grpc.v1.cri".registry]
  config_path = "/etc/containerd/certs.d"

[plugins."io.containerd.grpc.v1.cri".containerd.runtimes.runc]
  runtime_type = "io.containerd.runc.v2"
`,
							},
						},
					},
					extensionsv1alpha1.File{
						Path: "/etc/containerd/certs.d/ghcr.io/hosts.toml",
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								Data: `server = "https://ghcr.io"

[host."https://r.metal-stack.dev"]
  capabilities = ["pull", "resolve"]
`,
							},
						},
					},
					extensionsv1alpha1.File{
						Path: "/etc/containerd/certs.d/quay.io/hosts.toml",
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								Data: `server = "https://quay.io"

[host."https://r.metal-stack.dev"]
  capabilities = ["pull", "resolve"]
`,
							},
						},
					},
					extensionsv1alpha1.File{
						Path: "/etc/containerd/certs.d/docker.io/hosts.toml",
						Content: extensionsv1alpha1.FileContent{
							Inline: &extensionsv1alpha1.FileContentInline{
								Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								Data: `server = "https://docker.io"

[host."http://localhost:8080"]
  capabilities = ["pull", "resolve"]
`,
							},
						},
					},
				))
			})
		})
	})

	When("EnsureFiles", func() {
		Describe("Ensures files", func() {
			var (
				testFile1 = extensionsv1alpha1.File{
					Path: "/etc/foo",
					Content: extensionsv1alpha1.FileContent{
						Inline: &extensionsv1alpha1.FileContentInline{
							Data: "foo",
						},
					},
				}
				testFile2 = extensionsv1alpha1.File{
					Path: "/etc/bar",
					Content: extensionsv1alpha1.FileContent{
						Inline: &extensionsv1alpha1.FileContentInline{
							Data: "bar",
						},
					},
				}
				testFile3 = extensionsv1alpha1.File{
					Path: "/etc/bar",
					Content: extensionsv1alpha1.FileContent{
						Inline: &extensionsv1alpha1.FileContentInline{
							Data: "bar different",
						},
					},
				}
			)

			It("Ensures a single file into empty base", func() {
				result := EnsureFiles([]extensionsv1alpha1.File{}, testFile1)
				Expect(result).To(ConsistOf(testFile1))
			})

			It("Ensures no file into non-empty base", func() {
				result := EnsureFiles([]extensionsv1alpha1.File{
					testFile2,
				})
				Expect(result).To(ConsistOf(testFile2))
			})

			It("Ensures a single file into non-empty base", func() {
				result := EnsureFiles([]extensionsv1alpha1.File{
					testFile2,
				}, testFile1)
				Expect(result).To(ConsistOf(testFile2, testFile1))
			})

			It("Ensures only single file is added", func() {
				result := EnsureFiles([]extensionsv1alpha1.File{
					testFile2,
				}, testFile3)
				Expect(result).To(ConsistOf(testFile3))
			})
		})
	})
})

func mustMarshal(data any) []byte {
	raw, _ := json.Marshal(data) //nolint
	return raw
}
