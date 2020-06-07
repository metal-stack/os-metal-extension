package pkg

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/coreos/container-linux-config-transpiler/config/types"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

func TestIgnitionFromOperatingSystemConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *extensionsv1alpha1.OperatingSystemConfig
		want    types.Config
		wantErr bool
	}{
		{
			name: "simple service",
			config: &extensionsv1alpha1.OperatingSystemConfig{
				Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
					Units: []extensionsv1alpha1.Unit{
						{
							Name:    "kubelet.service",
							Content: strPtr("[Unit]\nDescription=kubelet\n[Install]\nWantedBy=multi-user.target\n[Service]\nExecStart=/bin/kubelet"),
							Enable:  boolPtr(true),
						},
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
							Enable:   true,
						},
					},
				},
			},
		},

		{
			name: "simple files",
			config: &extensionsv1alpha1.OperatingSystemConfig{
				Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
					Files: []extensionsv1alpha1.File{
						{
							Path: "/etc/hostname",
							Content: extensionsv1alpha1.FileContent{
								Inline: &extensionsv1alpha1.FileContentInline{
									Encoding: "",
									Data:     "testhost",
								},
							},
							Permissions: int32Ptr(0644),
						},
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
							Mode: intPtr(0644),
						},
					},
				},
			},
		},

		{
			name: "cri is enabled",
			config: &extensionsv1alpha1.OperatingSystemConfig{
				Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
					CRIConfig: &extensionsv1alpha1.CRIConfig{
						Name: "containerd",
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
									Contents: containerdSystemdConfig,
								},
							},
						},
					},
				},
				Storage: types.Storage{
					Files: []types.File{
						{
							Filesystem: "root",
							Path:       "/etc/containerd/config.toml",
							Contents: types.FileContents{
								Inline: containerdConfig,
							},
							Mode: intPtr(0644),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IgnitionFromOperatingSystemConfig(context.Background(), nil, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("IgnitionFromOperatingSystemConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			cfg, _ := types.Convert(tt.want, "", nil)
			want, err := json.Marshal(cfg)
			if err != nil {
				t.Error(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("IgnitionFromOperatingSystemConfig()\ngot:\n%v\nwant:\n %v", string(got), string(want))
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}
func intPtr(i int) *int {
	return &i
}
