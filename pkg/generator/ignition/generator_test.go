package ignition

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/flatcar/container-linux-config-transpiler/config/types"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
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
							Enabled:  pointer.BoolPtr(true),
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
						TransmitUnencoded: pointer.BoolPtr(true),
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
				Storage: types.Storage{
					Files: []types.File{
						{
							Filesystem: "root",
							Path:       "/etc/containerd/config.toml",
							Contents: types.FileContents{
								Inline: containerdConfig,
							},
							Mode: pointer.Int(0644),
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
