package ignition

import (
	"encoding/json"
	"fmt"
	"strings"

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
	"k8s.io/utils/ptr"
)

// IgnitionGenerator generates cloud-init scripts.
type IgnitionGenerator struct {
	cloudInitGenerator generator.Generator
	cmd                string
}

// New creates a new IgnitionGenerator with the given units path.
func New(unitsPath string, cmd string, additionalValuesFunc func(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error)) *IgnitionGenerator {
	return &IgnitionGenerator{
		cloudInitGenerator: ostemplate.NewCloudInitGenerator(nil, unitsPath, cmd, additionalValuesFunc),
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

	imageProviderConfig := &metalextensionv1alpha1.ImageProviderConfig{}
	networkIsolation := &metalextensionv1alpha1.NetworkIsolation{}
	if config.Object != nil && config.Object.Spec.ProviderConfig != nil {
		err := decodeProviderConfig(config.Object.Spec.ProviderConfig, imageProviderConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to decode providerConfig")
		}
	}
	if imageProviderConfig.NetworkIsolation != nil {
		networkIsolation = imageProviderConfig.NetworkIsolation
	}

	cfg.Systemd = types.Systemd{}
	for _, u := range config.Units {
		contents := string(u.Content)

		unit := types.SystemdUnit{
			Contents: contents,
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

	dnsFiles := additionalDNSConfFiles(networkIsolation.DNSServers)
	cfg.Storage.Files = append(cfg.Storage.Files, dnsFiles...)

	ntpFiles := additionalNTPConfFiles(networkIsolation.NTPServers)
	cfg.Storage.Files = append(cfg.Storage.Files, ntpFiles...)

	if config.CRI != nil {
		cri := config.CRI
		if cri.Name == extensionsv1alpha1.CRINameContainerD && len(networkIsolation.RegistryMirrors) > 0 {
			cfg.Storage.Files = append(cfg.Storage.Files, additionalContainterdConfigFile(networkIsolation.RegistryMirrors))
		}
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

func additionalContainterdConfigFile(mirrors []metalextensionv1alpha1.RegistryMirror) types.File {
	content := `# Generated by os-extension-metal
version = 2

[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
`
	for _, m := range mirrors {
		for _, of := range m.MirrorOf {
			content += fmt.Sprintf(`    [plugins."io.containerd.grpc.v1.cri".registry.mirrors.%q]
      endpoint = [%q]
`, of, m.Endpoint)
		}
	}

	return types.File{
		Path:       "/etc/containerd/conf.d/isolated-cluster.toml",
		Filesystem: "root",
		Mode:       &types.DefaultFileMode,
		Contents: types.FileContents{
			Inline: content,
		},
	}

}

func additionalDNSConfFiles(dnsServers []string) []types.File {
	if len(dnsServers) == 0 {
		return nil
	}
	resolveDNS := strings.Join(dnsServers, " ")
	systemdResolvedConfd := fmt.Sprintf(`# Generated by os-extension-metal

[Resolve]
DNS=%s
Domain=~.

`, resolveDNS)
	resolvConf := "# Generated by os-extension-metal\n"
	for _, ip := range dnsServers {
		resolvConf += fmt.Sprintf("nameserver %s\n", ip)
	}

	return []types.File{
		{
			Path: "/etc/systemd/resolved.conf.d/dns.conf",
			Contents: types.FileContents{
				Inline: systemdResolvedConfd,
			},
		},
		{
			Path: "/etc/resolv.conf",
			Contents: types.FileContents{
				Inline: resolvConf,
			},
		},
	}
}

func additionalNTPConfFiles(ntpServers []string) []types.File {
	if len(ntpServers) == 0 {
		return nil
	}
	ntps := strings.Join(ntpServers, " ")
	renderedContent := fmt.Sprintf(`# Generated by os-extension-metal

[Time]
NTP=%s
`, ntps)

	return []types.File{
		{
			Path: "/etc/systemd/timesyncd.conf",
			Contents: types.FileContents{
				Inline: renderedContent,
			},
		},
	}
}
