// Copyright 2023 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operatingsystemconfig

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	gardenv1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	metalextensionv1alpha1 "github.com/metal-stack/gardener-extension-provider-metal/pkg/apis/metal/v1alpha1"
	"github.com/metal-stack/os-metal-extension/pkg/controller/operatingsystemconfig/ignition"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	containerdConfig = `# Generated by os-extension-metal
version = 2
imports = ["/etc/containerd/conf.d/*.toml"]
disabled_plugins = []

[plugins."io.containerd.grpc.v1.cri".registry]
  config_path = "/etc/containerd/certs.d"
`
)

type actuator struct {
	client  client.Client
	decoder runtime.Decoder
}

// NewActuator creates a new Actuator that updates the status of the handled OperatingSystemConfig resources.
func NewActuator(mgr manager.Manager) operatingsystemconfig.Actuator {
	scheme := runtime.NewScheme()
	utilruntime.Must(gardenv1beta1.AddToScheme(scheme))
	decoder := serializer.NewCodecFactory(scheme).UniversalDecoder()

	return &actuator{
		client:  mgr.GetClient(),
		decoder: decoder,
	}
}

func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	imageProviderConfig := &metalextensionv1alpha1.ImageProviderConfig{}

	networkIsolation := &metalextensionv1alpha1.NetworkIsolation{}
	if osc.Spec.ProviderConfig != nil {
		err := decodeProviderConfig(a.decoder, osc.Spec.ProviderConfig, imageProviderConfig)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("unable to decode providerConfig")
		}
	}
	if imageProviderConfig.NetworkIsolation != nil {
		networkIsolation = imageProviderConfig.NetworkIsolation
	}

	extensionFiles := getExtensionFiles(osc, networkIsolation)

	switch purpose := osc.Spec.Purpose; purpose {
	case extensionsv1alpha1.OperatingSystemConfigPurposeProvision:
		osc := osc.DeepCopy()
		osc.Spec.Files = EnsureFiles(osc.Spec.Files, extensionFiles...)

		userData, err := ignition.New(log).Transpile(osc)
		return userData, nil, nil, err

	case extensionsv1alpha1.OperatingSystemConfigPurposeReconcile:
		return nil, nil, extensionFiles, nil
	default:
		return nil, nil, nil, fmt.Errorf("unknown purpose: %s", purpose)
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

func (a *actuator) Restore(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	return a.Reconcile(ctx, log, osc)
}

func getExtensionFiles(osc *extensionsv1alpha1.OperatingSystemConfig, networkIsolation *metalextensionv1alpha1.NetworkIsolation) []extensionsv1alpha1.File {
	var extensionFiles []extensionsv1alpha1.File

	if len(networkIsolation.RegistryMirrors) > 0 {
		// TODO: this is only required for backwards-compatibility before we started to create worker machines with DNS and NTP configuration through metal-stack
		// otherwise existing machines would lose connectivity because the GNA cleans up the dns and ntp definitions
		// references https://github.com/metal-stack/gardener-extension-provider-metal/issues/433
		//
		// can potentially be cleaned up as soon as there are no worker nodes of isolated clusters anymore that were created without dns and ntp configuration
		// ideally a point in time should be defined when we add the dns and ntp to the worker hashes to enforce the setting

		dnsFiles := additionalDNSConfFiles(networkIsolation.DNSServers)
		extensionFiles = append(extensionFiles, dnsFiles...)

		ntpFiles := additionalNTPConfFiles(networkIsolation.NTPServers)
		extensionFiles = append(extensionFiles, ntpFiles...)
	}

	if osc.Spec.CRIConfig != nil && osc.Spec.CRIConfig.Name == extensionsv1alpha1.CRINameContainerD {
		// TODO: as soon as all clusters run at least 1.31 we can remove the containerd config.toml override
		// the file will be fully managed by the GNA and latest metal-os images render the containerd default config
		if osc.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeReconcile && (osc.Spec.CRIConfig.CgroupDriver == nil || *osc.Spec.CRIConfig.CgroupDriver != extensionsv1alpha1.CgroupDriverSystemd) {
			extensionFiles = append(extensionFiles, extensionsv1alpha1.File{
				Path:        "/etc/containerd/config.toml",
				Permissions: ptr.To(int32(0644)),
				Content: extensionsv1alpha1.FileContent{
					Inline: &extensionsv1alpha1.FileContentInline{
						Encoding: string(extensionsv1alpha1.PlainFileCodecID),
						Data:     containerdConfig,
					},
				},
			})
		}

		if len(networkIsolation.RegistryMirrors) > 0 {
			extensionFiles = append(extensionFiles, additionalContainerdMirrors(networkIsolation.RegistryMirrors)...)
		}
	}

	return extensionFiles
}

// decodeProviderConfig decodes the provider config into the given struct
func decodeProviderConfig(decoder runtime.Decoder, providerConfig *runtime.RawExtension, into runtime.Object) error {
	if providerConfig == nil {
		return nil
	}

	if _, _, err := decoder.Decode(providerConfig.Raw, nil, into); err != nil {
		return fmt.Errorf("could not decode provider config: %w", err)
	}

	return nil
}

func additionalContainerdMirrors(mirrors []metalextensionv1alpha1.RegistryMirror) []extensionsv1alpha1.File {
	var files []extensionsv1alpha1.File

	for _, m := range mirrors {
		for _, of := range m.MirrorOf {
			content := fmt.Sprintf(`server = "https://%s"

[host.%q]
  capabilities = ["pull", "resolve"]
`, of, m.Endpoint)

			files = append(files, extensionsv1alpha1.File{
				Path: fmt.Sprintf("/etc/containerd/certs.d/%s/hosts.toml", of),
				Content: extensionsv1alpha1.FileContent{
					Inline: &extensionsv1alpha1.FileContentInline{
						Encoding: string(extensionsv1alpha1.PlainFileCodecID),
						Data:     content,
					},
				},
			})
		}
	}

	return files
}

func additionalDNSConfFiles(dnsServers []string) []extensionsv1alpha1.File {
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

	// TODO: in osc.Spec.Type we can get the distro "ubuntu", "debian", "nvidia", ...
	// from this information we should be able to deduce if systemd-resolved is used or not

	return []extensionsv1alpha1.File{
		{
			Path: "/etc/systemd/resolved.conf.d/dns.conf",
			Content: extensionsv1alpha1.FileContent{
				Inline: &extensionsv1alpha1.FileContentInline{
					Encoding: string(extensionsv1alpha1.PlainFileCodecID),
					Data:     systemdResolvedConfd,
				},
			},
		},
		{
			Path: "/etc/resolv.conf",
			Content: extensionsv1alpha1.FileContent{
				Inline: &extensionsv1alpha1.FileContentInline{
					Encoding: string(extensionsv1alpha1.PlainFileCodecID),
					Data:     resolvConf,
				},
			},
		},
	}
}

func additionalNTPConfFiles(ntpServers []string) []extensionsv1alpha1.File {
	if len(ntpServers) == 0 {
		return nil
	}
	ntps := strings.Join(ntpServers, " ")
	renderedContent := fmt.Sprintf(`# Generated by os-extension-metal
[Time]
NTP=%s
`, ntps)

	return []extensionsv1alpha1.File{
		{
			Path: "/etc/systemd/timesyncd.conf",
			Content: extensionsv1alpha1.FileContent{
				Inline: &extensionsv1alpha1.FileContentInline{
					Encoding: string(extensionsv1alpha1.PlainFileCodecID),
					Data:     renderedContent,
				},
			},
			Permissions: ptr.To(int32(0644)),
		},
	}
}

func EnsureFiles(base []extensionsv1alpha1.File, files ...extensionsv1alpha1.File) []extensionsv1alpha1.File {
	var res []extensionsv1alpha1.File

	res = append(res, base...)

	for _, file := range files {
		index := slices.IndexFunc(base, func(elem extensionsv1alpha1.File) bool {
			return elem.Path == file.Path
		})

		if index < 0 {
			res = append(res, file)
		} else {
			res[index] = file
		}
	}

	return res
}
