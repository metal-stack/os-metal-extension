package pkg

import (
	"github.com/coreos/ignition/config/v2_2/types"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

func fromGardener(config *extensionsv1alpha1.OperatingSystemConfig) (*types.Config, error) {

	cfg := &types.Config{}
	cfg.Systemd = types.Systemd{}
	for _, u := range config.Spec.Units {
		unit := types.Unit{
			Contents: *u.Content,
			Name:     u.Name,
			Enable:   *u.Enable,
		}
		for _, dr := range u.DropIns {
			unit.Dropins = append(unit.Dropins, types.SystemdDropin{
				Name:     dr.Name,
				Contents: dr.Content,
			})
		}
		cfg.Systemd.Units = append(cfg.Systemd.Units, unit)
	}

	cfg.Storage = types.Storage{}
	for _, f := range config.Spec.Files {

		ignitionFile := types.File{
			Path: f.

		}
		cfg.Storage.Files = append(cfg.Storage.Files, ignitionFile)
	}
	return cfg, nil
}
