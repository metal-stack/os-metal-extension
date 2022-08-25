// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package generator

import (
	"embed"

	commongen "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/metal-stack/os-metal-extension/pkg/generator/ignition"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var cmd = "/usr/bin/env bash %s"
var ignitionGenerator commongen.Generator

func additionalValues(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error) {
	return nil, nil
}

//go:embed templates/*
var templates embed.FS

func init() {
	cloudInitTemplateString, err := templates.ReadFile("templates/cloud-init.sh.template")
	runtime.Must(err)

	cloudInitTemplate, err := ostemplate.NewTemplate("cloud-init").Parse(string(cloudInitTemplateString))
	runtime.Must(err)
	ignitionGenerator = ignition.New(cloudInitTemplate, ostemplate.DefaultUnitsPath, cmd, additionalValues)
}

// IgnitionGenerator is the generator which will genereta the cloud init yaml
func IgnitionGenerator() commongen.Generator {
	return ignitionGenerator
}
