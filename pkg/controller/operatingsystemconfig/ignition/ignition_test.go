package ignition

import (
	"testing"

	"github.com/flatcar/container-linux-config-transpiler/config/types"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
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
							Content: ptr.To("[Unit]\nDescription=kubelet\n[Install]\nWantedBy=multi-user.target\n[Service]\nExecStart=/bin/kubelet"),
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
							Enabled:  ptr.To(true),
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
									Data: "testhost",
								},
							},
							Permissions: ptr.To(int32(0644)),
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
							Mode:      ptr.To(0644),
							Overwrite: ptr.To(true),
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
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
