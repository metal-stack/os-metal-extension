package ignition

import (
	"encoding/json"
	"fmt"

	"github.com/coreos/container-linux-config-transpiler/config/types"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/utils"
)

const (
	containerdSystemdDropin = `
[Service]
ExecStart=
ExecStart=/usr/bin/containerd --config=/etc/containerd/config.toml
`
	containerdConfig = `
# created by os-extension-metal
[plugins.cri.registry.mirrors]
  [plugins.cri.registry.mirrors."docker.io"]
    endpoint = ["https://mirror.gcr.io"]
`
)

// IgnitionFromOperatingSystemConfig is responsible to transpile the gardener OperatingSystemConfig to a ignition configuration.
// This is currently done with container-linux-config-transpile v0.9.0 and creates ignition v2.2.0 compatible configuration,
// which is used by ignition 0.32.0.
// TODO
// Starting with ignition 2.0, ignition itself contains the required parsing logic, so we can use ignition directly.
// see https://github.com/coreos/ignition/blob/master/config/config.go#L38
// Therefore we must update ignition to 2.0.0 in the images and transform the gardener config to the ignition config types instead.
func IgnitionFromOperatingSystemConfig(config *generator.OperatingSystemConfig) ([]byte, error) {
	cfg := types.Config{}
	cfg.Systemd = types.Systemd{}
	for _, u := range config.Units {
		contents := string(u.Content)

		unit := types.SystemdUnit{
			Contents: contents,
			Name:     u.Name,
			Enable:   true,
		}
		for _, dr := range u.DropIns {
			unit.Dropins = append(unit.Dropins, types.SystemdUnitDropIn{
				Name:     dr.Name,
				Contents: string(dr.Content),
			})
		}
		cfg.Systemd.Units = append(cfg.Systemd.Units, unit)
	}

	cfg.Storage = types.Storage{}
	for _, f := range config.Files {
		var mode *int
		if f.Permissions != nil {
			m := int(*f.Permissions)
			mode = &m
		}

		var inline string
		if f.TransmitUnencoded != nil && *f.TransmitUnencoded {
			inline = string(f.Content)
		} else {
			inline = utils.EncodeBase64([]byte(f.Content))
		}

		ignitionFile := types.File{
			Path:       f.Path,
			Filesystem: "root",
			Mode:       mode,
			Contents: types.FileContents{
				Inline: inline,
			},
		}
		cfg.Storage.Files = append(cfg.Storage.Files, ignitionFile)
	}

	if config.CRI != nil {
		cri := config.CRI
		if cri.Name == extensionsv1alpha1.CRINameContainerD {

			containerdSystemdService := types.SystemdUnit{
				Name: "containerd.service",
				Dropins: []types.SystemdUnitDropIn{
					{
						Name:     "11-exec_config.conf",
						Contents: containerdSystemdDropin,
					},
				},
			}
			cfg.Systemd.Units = append(cfg.Systemd.Units, containerdSystemdService)

			containerdConfigFile := types.File{
				Path:       "/etc/containerd/config.toml",
				Filesystem: "root",
				Mode:       &types.DefaultFileMode,
				Contents: types.FileContents{
					Inline: containerdConfig,
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
