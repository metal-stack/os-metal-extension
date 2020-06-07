package pkg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/container-linux-config-transpiler/config/types"
	oscommon "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/actuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	containerdSystemdConfig = `
[Service]
ExecStart=
ExecStart=/usr/bin/containerd --config=/etc/containerd/config.toml
`
	containerdConfig = `
# created by os-extension-metal
# This config is intentially left blank to force containerd to be started with default config
`
)

// IgnitionFromOperatingSystemConfig is responsible to transpile the gardener OperatingSystemConfig to a ignition configuration.
// This is currently done with container-linux-config-transpile v0.9.0 and creates ignition v2.2.0 compatible configuration,
// which is used by ignition 0.32.0.
// TODO
// Starting with ignition 2.0, ignition itself contains the required parsing logic, so we can use ignition directly.
// see https://github.com/coreos/ignition/blob/master/config/config.go#L38
// Therefore we must update ignition to 2.0.0 in the images and transform the gardener config to the ignition config types instead.
func IgnitionFromOperatingSystemConfig(ctx context.Context, c client.Client, config *extensionsv1alpha1.OperatingSystemConfig) ([]byte, error) {
	cfg := types.Config{}
	cfg.Systemd = types.Systemd{}
	for _, u := range config.Spec.Units {
		var contents string
		if u.Content != nil {
			contents = *u.Content
		}

		var enable bool
		if u.Enable != nil {
			enable = *u.Enable
		}

		unit := types.SystemdUnit{
			Contents: contents,
			Name:     u.Name,
			Enable:   enable,
		}
		for _, dr := range u.DropIns {
			unit.Dropins = append(unit.Dropins, types.SystemdUnitDropIn{
				Name:     dr.Name,
				Contents: dr.Content,
			})
		}
		cfg.Systemd.Units = append(cfg.Systemd.Units, unit)
	}

	cfg.Storage = types.Storage{}
	for _, f := range config.Spec.Files {
		content, err := oscommon.DataForFileContent(ctx, c, config.Namespace, &f.Content)
		if err != nil {
			return nil, err
		}

		var mode *int
		if f.Permissions != nil {
			m := int(*f.Permissions)
			mode = &m
		}

		ignitionFile := types.File{
			Path:       f.Path,
			Filesystem: "root",
			Mode:       mode,
			Contents: types.FileContents{
				Inline: string(content),
			},
		}
		cfg.Storage.Files = append(cfg.Storage.Files, ignitionFile)
	}

	if config.Spec.CRIConfig != nil {
		cri := config.Spec.CRIConfig
		if cri.Name == extensionsv1alpha1.CRINameContainerD {
			containerdSystemdConfigFile := types.File{
				Path:       "/etc/systemd/system/containerd.service.d/11-exec_config.conf",
				Filesystem: "root",
				Mode:       &types.DefaultFileMode,
				Contents: types.FileContents{
					Inline: string(containerdSystemdConfig),
				},
			}
			cfg.Storage.Files = append(cfg.Storage.Files, containerdSystemdConfigFile)

			containerdConfigFile := types.File{
				Path:       "/etc/containerd/config.toml",
				Filesystem: "root",
				Mode:       &types.DefaultFileMode,
				Contents: types.FileContents{
					Inline: string(containerdConfig),
				},
			}
			cfg.Storage.Files = append(cfg.Storage.Files, containerdConfigFile)
		}
	}

	outCfg, report := types.Convert(cfg, "", nil)
	if report.IsFatal() {
		return nil, fmt.Errorf("could not transpile ignition config: %s", report.String())
	}

	return json.Marshal(outCfg)
}
