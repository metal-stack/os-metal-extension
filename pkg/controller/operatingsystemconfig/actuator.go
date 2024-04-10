package operatingsystemconfig

import (
	"context"
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	oscommonactuator "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/actuator"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	metalextensionv1alpha1 "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/metal-stack/os-metal-extension/pkg/controller/operatingsystemconfig/generator"
)

type actuator struct {
	client               client.Client
	useGardenerNodeAgent bool
}

// NewActuator creates a new Actuator that updates the status of the handled OperatingSystemConfig resources.
func NewActuator(mgr manager.Manager, useGardenerNodeAgent bool) operatingsystemconfig.Actuator {
	return &actuator{
		client:               mgr.GetClient(),
		useGardenerNodeAgent: useGardenerNodeAgent,
	}
}

func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, *string, []string, []string, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	cloudConfig, command, err := oscommonactuator.CloudConfigFromOperatingSystemConfig(ctx, log, a.client, osc, generator.IgnitionGenerator())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("could not generate cloud config: %w", err)
	}

	switch purpose := osc.Spec.Purpose; purpose {
	case extensionsv1alpha1.OperatingSystemConfigPurposeProvision:
		if !a.useGardenerNodeAgent {
			return cloudConfig, command, oscommonactuator.OperatingSystemConfigUnitNames(osc), oscommonactuator.OperatingSystemConfigFilePaths(osc), nil, nil, nil
		}
		userData, err := a.handleProvisionOSC(ctx, osc)
		return []byte(userData), nil, nil, nil, nil, nil, err

	case extensionsv1alpha1.OperatingSystemConfigPurposeReconcile:
		extensionUnits, extensionFiles, err := a.handleReconcileOSC(osc)
		return cloudConfig, command, oscommonactuator.OperatingSystemConfigUnitNames(osc), oscommonactuator.OperatingSystemConfigFilePaths(osc), extensionUnits, extensionFiles, err

	default:
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("unknown purpose: %s", purpose)
	}
}

func (a *actuator) Delete(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.OperatingSystemConfig) error {
	return nil
}

func (a *actuator) Migrate(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) error {
	return a.Delete(ctx, log, osc)
}

func (a *actuator) ForceDelete(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) error {
	return a.Delete(ctx, log, osc)
}

func (a *actuator) Restore(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, *string, []string, []string, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	return a.Reconcile(ctx, log, osc)
}

func (a *actuator) handleProvisionOSC(ctx context.Context, osc *extensionsv1alpha1.OperatingSystemConfig) (string, error) {
	writeFilesToDiskScript, err := operatingsystemconfig.FilesToDiskScript(ctx, a.client, osc.Namespace, osc.Spec.Files)
	if err != nil {
		return "", err
	}
	writeUnitsToDiskScript := operatingsystemconfig.UnitsToDiskScript(osc.Spec.Units)

	script := `#!/bin/bash
# Fix mis-configuration of dockerd
mkdir -p /etc/docker
echo '{ "storage-driver": "devicemapper" }' > /etc/docker/daemon.json
sed -i '/Environment=DOCKER_SELINUX=--selinux-enabled=true/s/^/#/g' /run/systemd/system/docker.service

# Change existing worker to use docker registry-mirror
file="/etc/docker/daemon.json"
if [ $(jq -r 'has("registry-mirrors")' "${file}") == "false" ]; then
    contents=$(jq -M '. += {"registry-mirrors": ["https://mirror.gcr.io"]}' ${file})
    echo "${contents}" > ${file}
fi

systemctl daemon-reload
systemctl reload docker
` + writeFilesToDiskScript + `
` + writeUnitsToDiskScript + `

`

	for _, unit := range osc.Spec.Units {
		script += fmt.Sprintf(`systemctl enable '%s' && systemctl restart --no-block '%s'
`, unit.Name, unit.Name)
	}

	return script, nil
}

var decoder runtime.Decoder

func init() {
	scheme := runtime.NewScheme()
	utilruntime.Must(gardenv1beta1.AddToScheme(scheme))
	decoder = serializer.NewCodecFactory(scheme).UniversalDecoder()
}

func (a *actuator) handleReconcileOSC(osc *extensionsv1alpha1.OperatingSystemConfig) ([]extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	var (
		extensionUnits []extensionsv1alpha1.Unit
		extensionFiles []extensionsv1alpha1.File
	)

	imageProviderConfig := &metalextensionv1alpha1.ImageProviderConfig{}
	if osc != nil && osc.Spec.ProviderConfig != nil {
		if _, _, err := decoder.Decode(osc.Spec.ProviderConfig.Raw, nil, imageProviderConfig); err != nil {
			return nil, nil, fmt.Errorf("could not decode provider config: %w", err)
		}
	}

	// containerd
	filePathContainerdConfig := filepath.Join("/", "etc", "containerd", "config.toml")
	extensionUnits = append(extensionUnits, containerdUnits(filePathContainerdConfig)...)
	extensionFiles = append(extensionFiles, containerdConfigFiles(filePathContainerdConfig, imageProviderConfig.NetworkIsolation)...)

	// dns/ntp config files for network isolation
	extensionFiles = append(extensionFiles, dnsConfigFiles(imageProviderConfig.NetworkIsolation)...)
	extensionFiles = append(extensionFiles, ntpConfigFiles(imageProviderConfig.NetworkIsolation)...)

	return extensionUnits, extensionFiles, nil
}

func containerdUnits(filePathContainerdConfig string) []extensionsv1alpha1.Unit {
	return []extensionsv1alpha1.Unit{
		{

			Name: "containerd.service",
			DropIns: []extensionsv1alpha1.DropIn{{
				Name: "11-exec_config.conf",
				Content: `[Service]
ExecStart=
ExecStart=` + filePathContainerdConfig + `
`,
			}},
			FilePaths: []string{filePathContainerdConfig},
		},
	}
}

func containerdConfigFiles(filePathContainerdConfig string, networkIsolation *metalextensionv1alpha1.NetworkIsolation) []extensionsv1alpha1.File {
	file := extensionsv1alpha1.File{
		Path:        filePathContainerdConfig,
		Permissions: pointer.Int32(0644),
		Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `
# created by os-extension-metal
imports = ["/etc/containerd/conf.d/*.toml"]

version = 2

[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
      endpoint = ["https://mirror.gcr.io"]
`}},
	}

	if networkIsolation != nil && len(networkIsolation.RegistryMirrors) > 0 {
		file.Content.Inline.Data = `# Generated by os-extension-metal
imports = ["/etc/containerd/conf.d/*.toml"]
version = 2
[plugins."io.containerd.grpc.v1.cri".registry]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors]
`
		for _, m := range networkIsolation.RegistryMirrors {
			for _, of := range m.MirrorOf {
				file.Content.Inline.Data += fmt.Sprintf(`    [plugins."io.containerd.grpc.v1.cri".registry.mirrors.%q]
      endpoint = [%q]
`, of, m.Endpoint)
			}
		}
	}

	return []extensionsv1alpha1.File{file}
}

func dnsConfigFiles(networkIsolation *metalextensionv1alpha1.NetworkIsolation) []extensionsv1alpha1.File {
	if networkIsolation == nil || len(networkIsolation.DNSServers) == 0 {
		return nil
	}

	resolvConf := "# Generated by os-extension-metal\n"
	for _, ip := range networkIsolation.DNSServers {
		resolvConf += fmt.Sprintf("nameserver %s\n", ip)
	}

	return []extensionsv1alpha1.File{
		{
			Path:        "/etc/systemd/resolved.conf.d/dns.conf",
			Permissions: pointer.Int32(0644),
			Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: fmt.Sprintf(`# Generated by os-extension-metal

[Resolve]
DNS=%s
Domain=~.

`, strings.Join(networkIsolation.DNSServers, " "))}},
		},
		{
			Path:        "/etc/resolv.conf",
			Permissions: pointer.Int32(0644),
			Content:     extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: resolvConf}},
		},
	}
}

func ntpConfigFiles(networkIsolation *metalextensionv1alpha1.NetworkIsolation) []extensionsv1alpha1.File {
	if networkIsolation == nil || len(networkIsolation.NTPServers) == 0 {
		return nil
	}

	return []extensionsv1alpha1.File{{
		Path:        "/etc/systemd/timesyncd.conf",
		Permissions: pointer.Int32(0644),
		Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: fmt.Sprintf(`# Generated by os-extension-metal

[Time]
NTP=%s
`, strings.Join(networkIsolation.NTPServers, " "))}},
	}}
}
