// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package pkg

import (
	"context"
	"fmt"

	oscommon "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/actuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/metal-stack/os-metal-extension/pkg/internal"
)

func (a *actuator) reconcile(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) ([]byte, *string, []string, error) {
	var (
		data []byte
		err  error
	)
	if config.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeProvision {
		data, err = IgnitionFromOperatingSystemConfig(ctx, a.client, config)
	} else {
		data, err = a.cloudConfigFromOperatingSystemConfig(ctx, config)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not generate cloud config: %w", err)
	}

	// if ReloadConfigFilePath is given, this is executed to actually reload the written config
	var command *string
	if path := config.Spec.ReloadConfigFilePath; path != nil {
		cmd := fmt.Sprintf("/usr/bin/env bash %s", *path)
		command = &cmd
	}

	return data, command, operatingSystemConfigUnitNames(config), nil
}

// config is brought from Gardener
func (a *actuator) cloudConfigFromOperatingSystemConfig(ctx context.Context, config *extensionsv1alpha1.OperatingSystemConfig) ([]byte, error) {
	files := make([]*internal.File, 0, len(config.Spec.Files))
	for _, file := range config.Spec.Files {
		data, err := oscommon.DataForFileContent(ctx, a.client, config.Namespace, &file.Content)
		if err != nil {
			return nil, err
		}

		files = append(files, &internal.File{Path: file.Path, Content: data, Permissions: file.Permissions})
	}

	// blacklist sctp kernel module
	if config.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeReconcile {
		files = append(files,
			&internal.File{
				Path:    "/etc/modprobe.d/sctp.conf",
				Content: []byte("install sctp /bin/true"),
			})
	}

	units := make([]*internal.Unit, 0, len(config.Spec.Units))
	for _, unit := range config.Spec.Units {
		var content []byte
		if unit.Content != nil {
			content = []byte(*unit.Content)
		}

		dropIns := make([]*internal.DropIn, 0, len(unit.DropIns))
		for _, dropIn := range unit.DropIns {
			dropIns = append(dropIns, &internal.DropIn{Name: dropIn.Name, Content: []byte(dropIn.Content)})
		}
		units = append(units, &internal.Unit{Name: unit.Name, Content: content, DropIns: dropIns})
	}

	return internal.NewCloudInitGenerator(internal.DefaultUnitsPath).
		Generate(&internal.OperatingSystemConfig{
			Bootstrap: config.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeProvision,
			Files:     files,
			Units:     units,
		})
}

func operatingSystemConfigUnitNames(config *extensionsv1alpha1.OperatingSystemConfig) []string {
	unitNames := make([]string, 0, len(config.Spec.Units))
	for _, unit := range config.Spec.Units {
		unitNames = append(unitNames, unit.Name)
	}
	return unitNames
}
