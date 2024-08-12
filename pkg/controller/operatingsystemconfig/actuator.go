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

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig"
	oscommonactuator "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/actuator"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/metal-stack/os-metal-extension/pkg/controller/operatingsystemconfig/generator"
)

type actuator struct {
	client client.Client
}

// NewActuator creates a new Actuator that updates the status of the handled OperatingSystemConfig resources.
func NewActuator(mgr manager.Manager) operatingsystemconfig.Actuator {
	return &actuator{
		client: mgr.GetClient(),
	}
}

func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, osc *extensionsv1alpha1.OperatingSystemConfig) ([]byte, *string, []string, []string, []extensionsv1alpha1.Unit, []extensionsv1alpha1.File, error) {
	cloudConfig, command, err := oscommonactuator.CloudConfigFromOperatingSystemConfig(ctx, log, a.client, osc, generator.IgnitionGenerator())
	if err != nil {
		return nil, nil, nil, nil, nil, nil, fmt.Errorf("could not generate cloud config: %w", err)
	}

	return cloudConfig, command, oscommonactuator.OperatingSystemConfigUnitNames(osc), oscommonactuator.OperatingSystemConfigFilePaths(osc), nil, nil, nil
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
