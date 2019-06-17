package pkg

import (
	"context"
	"encoding/json"
	"github.com/coreos/container-linux-config-transpiler/config/types"
	oscommon "github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig/oscommon/actuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

func (a *actuator) generateIgnitionConfig(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) ([]byte, error) {
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
        content, err := oscommon.DataForFileContent(ctx, a.client, config.Namespace, &f.Content)
		if err != nil {
			return nil, err
		}

        var mode *int
        if f.Permissions != nil {
        	m := int(*f.Permissions)
        	mode = &m
		}

		ignitionFile := types.File{
			Path: f.Path,
			Mode: mode,
			Contents: types.FileContents{
				Inline: string(content),
			},
		}
		cfg.Storage.Files = append(cfg.Storage.Files, ignitionFile)
	}

	outCfg, _ := types.Convert(cfg, "", nil)
    return json.Marshal(outCfg)
}
