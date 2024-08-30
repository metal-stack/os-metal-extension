package ignition

import (
	"testing"

	"github.com/flatcar/container-linux-config-transpiler/config/types"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
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
						{
							Path: "/etc/foo",
							Content: extensionsv1alpha1.FileContent{
								Inline: &extensionsv1alpha1.FileContentInline{
									Data:     "foo",
									Encoding: string(extensionsv1alpha1.PlainFileCodecID),
								},
							},
							Permissions: ptr.To(int32(0744)),
						},
						{
							Path: "/etc/bar",
							Content: extensionsv1alpha1.FileContent{
								Inline: &extensionsv1alpha1.FileContentInline{
									Data:     "YmFy",
									Encoding: string(extensionsv1alpha1.B64FileCodecID),
								},
							},
							Permissions: ptr.To(int32(0744)),
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
								Inline: "testhost",
							},
							Mode:      ptr.To(0644),
							Overwrite: ptr.To(true),
						},
						{
							Filesystem: "root",
							Path:       "/etc/foo",
							Contents: types.FileContents{
								Inline: "foo",
							},
							Mode:      ptr.To(0744),
							Overwrite: ptr.To(true),
						},
						{
							Filesystem: "root",
							Path:       "/etc/bar",
							Contents: types.FileContents{
								Inline: "bar",
							},
							Mode:      ptr.To(0744),
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
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

func Test_ignition_Transpile(t *testing.T) {
	tests := []struct {
		name    string
		osc     *extensionsv1alpha1.OperatingSystemConfig
		want    string
		wantErr bool
	}{
		{
			name: "transpiles to ignition format 2.3.0",
			osc: &extensionsv1alpha1.OperatingSystemConfig{
				Spec: extensionsv1alpha1.OperatingSystemConfigSpec{
					Files: []extensionsv1alpha1.File{
						{
							Path: "/etc/a",
						},
					},
				},
			},
			want: `{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.3.0"},"networkd":{},"passwd":{},"storage":{"files":[{"filesystem":"root","overwrite":true,"path":"/etc/a","contents":{"source":"data:,","verification":{}},"mode":420}]},"systemd":{}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &ignition{
				log: logr.Discard(),
			}
			got, err := tr.Transpile(tt.osc)
			if (err != nil) != tt.wantErr {
				t.Errorf("ignition.Transpile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(string(got), tt.want); diff != "" {
				t.Errorf("ignition.Transpile() diff = %s", diff)
			}
		})
	}
}
