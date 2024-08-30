package ignition

import (
	"testing"

	"github.com/flatcar/container-linux-config-transpiler/config/types"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"k8s.io/utils/ptr"
)

func TestIgnitionFromOperatingSystemConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *generator.OperatingSystemConfig
		want   types.Config
	}{
		{
			name: "simple service",
			config: &generator.OperatingSystemConfig{
				Units: []*generator.Unit{
					{
						Name:    "kubelet.service",
						Content: []byte("[Unit]\nDescription=kubelet\n[Install]\nWantedBy=multi-user.target\n[Service]\nExecStart=/bin/kubelet"),
					},
				},
			},
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
			config: &generator.OperatingSystemConfig{
				Files: []*generator.File{
					{
						Path:        "/etc/hostname",
						Content:     []byte("testhost"),
						Permissions: ptr.To(int32(0644)),
					},
					{
						Path:        "/etc/foo",
						Content:     []byte("foo"),
						Permissions: ptr.To(int32(0744)),
					},
				},
			},
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
					},
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := ignitionFromOperatingSystemConfig(tt.config)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("diff = %s", diff)
			}
		})
	}
}

func Test_ignition_Transpile(t *testing.T) {
	tests := []struct {
		name    string
		osc     *generator.OperatingSystemConfig
		want    string
		wantErr bool
	}{
		{
			name: "transpiles to ignition format 2.3.0",
			osc: &generator.OperatingSystemConfig{
				Object: &v1alpha1.OperatingSystemConfig{
					Spec: v1alpha1.OperatingSystemConfigSpec{
						Purpose: v1alpha1.OperatingSystemConfigPurposeProvision,
					},
				},
				Files: []*generator.File{
					{
						Path: "/etc/a",
					},
				},
			},
			want: `{"ignition":{"config":{},"security":{"tls":{}},"timeouts":{},"version":"2.3.0"},"networkd":{},"passwd":{},"storage":{"files":[{"filesystem":"root","overwrite":true,"path":"/etc/a","contents":{"source":"data:,","verification":{}},"mode":420}]},"systemd":{}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &IgnitionGenerator{}

			got, _, err := tr.Generate(logr.Discard(), tt.osc)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(string(got), tt.want); diff != "" {
				t.Errorf("diff = %s", diff)
			}
		})
	}
}
