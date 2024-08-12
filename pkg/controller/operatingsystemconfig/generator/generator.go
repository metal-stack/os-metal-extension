package generator

import (
	commongen "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"

	"github.com/metal-stack/os-metal-extension/pkg/controller/operatingsystemconfig/generator/ignition"
)

var (
	cmd = "/usr/bin/env bash %s"
)

// IgnitionGenerator is the generator which will generete the ignition userdata and provider-specific shell script parts.
func IgnitionGenerator() commongen.Generator {
	return ignition.New(ostemplate.DefaultUnitsPath, cmd, additionalValues)
}

func additionalValues(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error) {
	return nil, nil
}
