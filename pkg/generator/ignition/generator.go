package ignition

import (
	"encoding/json"
	"fmt"
	"text/template"

	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	metalextensionv1alpha1 "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/flatcar/container-linux-config-transpiler/config/types"
	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/pointer"
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

// IgnitionGenerator generates cloud-init scripts.
type IgnitionGenerator struct {
	cloudInitGenerator generator.Generator
	cmd                string
}

// New creates a new IgnitionGenerator with the given units path.
func New(template *template.Template, unitsPath string, cmd string, additionalValuesFunc func(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error)) *IgnitionGenerator {
	return &IgnitionGenerator{
		cloudInitGenerator: ostemplate.NewCloudInitGenerator(template, unitsPath, cmd, additionalValuesFunc),
		cmd:                cmd,
	}
}

// Generate generates an ignition script from the given OperatingSystemConfig.
func (t *IgnitionGenerator) Generate(logr logr.Logger, config *generator.OperatingSystemConfig) ([]byte, *string, error) {
	if config.Object.Spec.Purpose != extensionsv1alpha1.OperatingSystemConfigPurposeProvision {
		return t.cloudInitGenerator.Generate(logr, config)
	}

	var cmd *string
	if config.Path != nil {
		c := fmt.Sprintf(t.cmd, *config.Path)
		cmd = &c
	}

	data, err := ignitionFromOperatingSystemConfig(config)
	return data, cmd, err
}

// ignitionFromOperatingSystemConfig is responsible to transpile the gardener OperatingSystemConfig to a ignition configuration.
// This is currently done with container-linux-config-transpile v0.9.0 and creates ignition v2.2.0 compatible configuration,
// which is used by ignition 0.32.0.
// TODO
// Starting with ignition 2.0, ignition itself contains the required parsing logic, so we can use ignition directly.
// see https://github.com/coreos/ignition/blob/master/config/config.go#L38
// Therefore we must update ignition to 2.0.0 in the images and transform the gardener config to the ignition config types instead.
func ignitionFromOperatingSystemConfig(config *generator.OperatingSystemConfig) ([]byte, error) {
	cfg := types.Config{}
	cfg.Systemd = types.Systemd{}
	for _, u := range config.Units {
		contents := string(u.Content)

		unit := types.SystemdUnit{
			Contents: contents,
			Name:     u.Name,
			Enabled:  pointer.Bool(true),
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

		ignitionFile := types.File{
			Path:       f.Path,
			Filesystem: "root",
			Mode:       mode,
			Contents: types.FileContents{
				Inline: string(f.Content),
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

	if config.Object != nil && config.Object.Spec.ProviderConfig != nil {
		networkIsolation := &metalextensionv1alpha1.NetworkIsolation{}
		// This does not work because NetworkIsolation is not a RuntimeObject because it is missing ObjectMeta and TypeMeta
		err := decodeProviderConfig(config.Object.Spec.ProviderConfig, networkIsolation)
		if err != nil {
			return nil, fmt.Errorf("unable to decode providerConfig")
		}

		fmt.Errorf("networkIsolation:%#v", networkIsolation)
	}

	outCfg, report := types.Convert(cfg, "", nil)
	if report.IsFatal() {
		return nil, fmt.Errorf("could not transpile ignition config: %s", report.String())
	}

	return json.Marshal(outCfg)
}

// decodeProviderConfig decodes the provider config into the given struct
func decodeProviderConfig(providerConfig *runtime.RawExtension, into runtime.Object) error {
	if providerConfig == nil {
		return nil
	}

	if _, _, err := getGardenerDecoder().Decode(providerConfig.Raw, nil, into); err != nil {
		return fmt.Errorf("could not decode provider config: %w", err)
	}

	return nil
}

var (
	decoder runtime.Decoder
)

// getGardenerDecoder returns a decoder to decode Gardener objects
func getGardenerDecoder() runtime.Decoder {
	if decoder == nil {
		scheme := runtime.NewScheme()
		utilruntime.Must(gardenv1beta1.AddToScheme(scheme))
		decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
	}
	return decoder
}
