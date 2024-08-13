package ignition

import (
	"encoding/json"
	"fmt"

	"github.com/flatcar/container-linux-config-transpiler/config/types"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/apis/extensions/v1alpha1/helper"
	"github.com/go-logr/logr"
	"k8s.io/utils/ptr"
)

type ignition struct {
	log logr.Logger
}

// New creates a new IgnitionGenerator.
func New(log logr.Logger) *ignition {
	return &ignition{
		log: log,
	}
}

// Transpile transpiles the OSC into an ignition script.
func (t *ignition) Transpile(osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, error) {
	data, err := ignitionFromOperatingSystemConfig(osc)
	if err != nil {
		return nil, fmt.Errorf("unable to map osc into ignition config: %w", err)
	}

	out, report := types.Convert(data, "", nil)
	if report.IsFatal() {
		return nil, fmt.Errorf("could not transpile ignition config: %s", report.String())
	}

	return json.Marshal(out)
}

// ignitionFromOperatingSystemConfig is responsible to transpile the gardener OperatingSystemConfig to a ignition configuration.
// This is currently done with container-linux-config-transpile v0.9.0 and creates ignition v2.2.0 compatible configuration,
// which is used by ignition 0.32.0.
// TODO
// Starting with ignition 2.0, ignition itself contains the required parsing logic, so we can use ignition directly.
// see https://github.com/coreos/ignition/blob/master/config/config.go#L38
// Therefore we must update ignition to 2.0.0 in the images and transform the gardener config to the ignition config types instead.
func ignitionFromOperatingSystemConfig(osc *extensionsv1alpha1.OperatingSystemConfig) (types.Config, error) {
	cfg := types.Config{}

	cfg.Systemd = types.Systemd{}
	for _, u := range osc.Spec.Units {

		unit := types.SystemdUnit{
			Contents: ptr.Deref(u.Content, ""),
			Name:     u.Name,
			Enabled:  ptr.To(true),
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
	for _, f := range osc.Spec.Files {
		var mode *int
		if f.Permissions != nil {
			m := int(*f.Permissions)
			mode = &m
		}

		ignitionFile := types.File{
			Path:       f.Path,
			Filesystem: "root",
			Mode:       mode,
			// Contents: types.FileContents{
			// 	Inline: string(inline),
			// },
			Overwrite: ptr.To(true),
		}

		if f.Content.Inline != nil {
			inline, err := helper.Decode(f.Content.Inline.Encoding, []byte(f.Content.Inline.Data))
			if err != nil {
				return types.Config{}, fmt.Errorf("unable to decode content from osc: %w", err)
			}

			ignitionFile.Contents.Inline = string(inline)
		}

		cfg.Storage.Files = append(cfg.Storage.Files, ignitionFile)
	}

	return cfg, nil
}
