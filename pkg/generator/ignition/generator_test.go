package ignition

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flatcar/container-linux-config-transpiler/config/types"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	metalextensionv1alpha1 "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
)

func TestIgnitionFromOperatingSystemConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *generator.OperatingSystemConfig
		want    types.Config
		wantErr bool
	}{
		{
			name: "simple service",
			config: &generator.OperatingSystemConfig{
				Units: []*generator.Unit{
					{
						Name:    "kubelet.service",
						Content: []byte(("[Unit]\nDescription=kubelet\n[Install]\nWantedBy=multi-user.target\n[Service]\nExecStart=/bin/kubelet")),
					},
				},
			},
			wantErr: false,
			want: types.Config{
				Systemd: types.Systemd{
					Units: []types.SystemdUnit{
						{
							Name:     "kubelet.service",
							Contents: "[Unit]\nDescription=kubelet\n[Install]\nWantedBy=multi-user.target\n[Service]\nExecStart=/bin/kubelet",
							Enabled:  pointer.Bool(true),
						},
					},
				},
			},
		},

		{
			name: "simple files",
			config: &generator.OperatingSystemConfig{
				Files: []*generator.File{
					{
						Path:              "/etc/hostname",
						TransmitUnencoded: pointer.Bool(true),
						Content:           []byte("testhost"),
						Permissions:       pointer.Int32(0644),
					},
				},
			},
			wantErr: false,
			want: types.Config{
				Storage: types.Storage{
					Files: []types.File{
						{
							Filesystem: "root",
							Path:       "/etc/hostname",
							Contents: types.FileContents{
								// FIXME here should be testhosts ???
								Inline: "testhost",
							},
							Mode: pointer.Int(0644),
						},
					},
				},
			},
		},

		{
			name: "cri is enabled",
			config: &generator.OperatingSystemConfig{
				CRI: &extensionsv1alpha1.CRIConfig{
					Name: "containerd",
				},
			},
			wantErr: false,
			want: types.Config{
				Systemd: types.Systemd{
					Units: []types.SystemdUnit{
						{
							Name: "containerd.service",
							Dropins: []types.SystemdUnitDropIn{
								{
									Name:     "11-exec_config.conf",
									Contents: containerdSystemdDropin,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "containerd with network isolation",
			config: &generator.OperatingSystemConfig{
				CRI: &extensionsv1alpha1.CRIConfig{
					Name: extensionsv1alpha1.CRINameContainerD,
				},
				Object: &extensionsv1alpha1.OperatingSystemConfig{
					Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
						DefaultSpec: extensionsv1alpha1.DefaultSpec{
							ProviderConfig: &runtime.RawExtension{
								Raw: mustMarshal(t, &metalextensionv1alpha1.ImageProviderConfig{
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
									}}),
							},
						},
					},
				},
			},
			wantErr: false,
			want: types.Config{
				Systemd: types.Systemd{
					Units: []types.SystemdUnit{
						{
							Name: "containerd.service",
							Dropins: []types.SystemdUnitDropIn{
								{
									Name:     "11-exec_config.conf",
									Contents: containerdSystemdDropin,
								},
							},
						},
					},
				},
				Storage: types.Storage{
					Files: []types.File{
						{
							Path: "/etc/systemd/resolved.conf.d/dns.conf",
							Contents: types.FileContents{
								Inline: `# Generated by os-extension-metal

[Resolve]
DNS=1.1.1.1 1.0.0.1
Domain=~.

`,
							},
						},
						{
							Path: "/etc/resolv.conf",
							Contents: types.FileContents{
								Inline: `# Generated by os-extension-metal
nameserver 1.1.1.1
nameserver 1.0.0.1
`,
							},
						},
						{
							Path: "/etc/systemd/timesyncd.conf",
							Contents: types.FileContents{
								Inline: `# Generated by os-extension-metal

[Time]
NTP=134.60.1.27 134.60.111.110
`,
							},
						},
						{
							Filesystem: "root",
							Path:       "/etc/containerd/conf.d/isolated-clusters.toml",
							Mode:       &types.DefaultFileMode,
							Contents: types.FileContents{
								Inline: `# Generated by os-extension-metal

[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."ghcr.io"]
      endpoint = ["https://r.metal-stack.dev"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."quay.io"]
      endpoint = ["https://r.metal-stack.dev"]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["http://localhost:8080"]
`,
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := ignitionFromOperatingSystemConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ignitionFromOperatingSystemConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			cfg, _ := types.Convert(tt.want, "", nil)
			want, err := json.Marshal(cfg)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("ignitionFromOperatingSystemConfig()\ngot:\n%v\nwant:\n%v", string(got), string(want))
			}
		})
	}
}

func mustMarshal(t *testing.T, obj runtime.Object) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		t.Errorf("failed to marshal object %s", err)
	}
	return data
}
