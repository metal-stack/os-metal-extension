package operatingsystemconfig_test

import (
	"context"
	"encoding/json"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils/test"
	"github.com/go-logr/logr"
	metalextensionv1alpha1 "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
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
				Units:   []extensionsv1alpha1.Unit{{Name: "some-unit.service", Content: pointer.String("foo")}},
				Files:   []extensionsv1alpha1.File{{Path: "/some/file", Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: "bar"}}}},
			},
		}
	})

	When("UseGardenerNodeAgent is false", func() {
		BeforeEach(func() {
			actuator = NewActuator(mgr, false)
		})

		Describe("#Reconcile", func() {
			It("should not return an error", func() {
				userData, command, unitNames, fileNames, extensionUnits, extensionFiles, err := actuator.Reconcile(ctx, log, osc)
				Expect(err).NotTo(HaveOccurred())

				Expect(userData).NotTo(BeEmpty()) // legacy logic is tested in ./generator/generator_test.go
				Expect(command).To(BeNil())
				Expect(unitNames).To(ConsistOf("some-unit.service"))
				Expect(fileNames).To(ConsistOf("/some/file"))
				Expect(extensionUnits).To(BeEmpty())
				Expect(extensionFiles).To(BeEmpty())
			})
		})
	})

	When("UseGardenerNodeAgent is true", func() {
		BeforeEach(func() {
			actuator = NewActuator(mgr, true)
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
				Context("no network isolation", func() {
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
								Permissions: pointer.Int32(0644),
								Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `
# created by os-extension-metal
imports = ["/etc/containerd/conf.d/*.toml"]

version = 2

[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["https://mirror.gcr.io"]
`}},
							},
						))
					})
				})

				Context("with network isolation", func() {
					BeforeEach(func() {
						osc.Spec.ProviderConfig = &runtime.RawExtension{
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
								}},
							),
						}
					})

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
								Permissions: pointer.Int32(0644),
								Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `# Generated by os-extension-metal
imports = ["/etc/containerd/conf.d/*.toml"]
version = 2
[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."ghcr.io"]
      endpoint = ["https://r.metal-stack.dev"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."quay.io"]
      endpoint = ["https://r.metal-stack.dev"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["http://localhost:8080"]
`}},
							},
							extensionsv1alpha1.File{
								Path:        "/etc/systemd/resolved.conf.d/dns.conf",
								Permissions: pointer.Int32(0644),
								Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `# Generated by os-extension-metal

[Resolve]
DNS=1.1.1.1 1.0.0.1
Domain=~.

`}},
							},
							extensionsv1alpha1.File{
								Path:        "/etc/resolv.conf",
								Permissions: pointer.Int32(0644),
								Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `# Generated by os-extension-metal
nameserver 1.1.1.1
nameserver 1.0.0.1
`}},
							},
							extensionsv1alpha1.File{
								Path:        "/etc/systemd/timesyncd.conf",
								Permissions: pointer.Int32(0644),
								Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `# Generated by os-extension-metal

[Time]
NTP=134.60.1.27 134.60.111.110
`}},
							},
						))
					})
				})
			})
		})
	})
})

func mustMarshal(obj runtime.Object) []byte {
	data, err := json.Marshal(obj)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return data
}
