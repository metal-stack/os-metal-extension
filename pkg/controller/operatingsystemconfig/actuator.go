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
	"path/filepath"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	oscommonactuator "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/actuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
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
#Fix mis-configuration of dockerd
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
		script += fmt.Sprintf(`systemctl enable '%s' && systemctl restart '%s'
`, unit.Name, unit.Name)
	}

	return script, nil
}

func (a *actuator) handleReconcileOSC(_ *extensionsv1alpha1.OperatingSystemConfig) ([]extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	var (
		extensionUnits []extensionsv1alpha1.Unit
		extensionFiles []extensionsv1alpha1.File
	)

	filePathContainerdConfig := filepath.Join("/", "etc", "containerd", "config.toml")
	extensionFiles = append(extensionFiles, extensionsv1alpha1.File{
		Path:        filePathContainerdConfig,
		Permissions: pointer.Int32(0644),
		Content: extensionsv1alpha1.FileContent{Inline: &extensionsv1alpha1.FileContentInline{Data: `
# created by os-extension-metal
[plugins.cri.registry.mirrors]
  [plugins.cri.registry.mirrors."docker.io"]
    endpoint = ["https://mirror.gcr.io"]
`}},
	})
	extensionUnits = append(extensionUnits, extensionsv1alpha1.Unit{
		Name: "containerd.service",
		DropIns: []extensionsv1alpha1.DropIn{{
			Name: "11-exec_config.conf",
			Content: `[Service]
ExecStart=
ExecStart=` + filePathContainerdConfig + `
`,
		}},
		FilePaths: []string{filePathContainerdConfig},
	})

	return extensionUnits, extensionFiles, nil
}
