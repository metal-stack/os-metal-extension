package generator

import (
	"embed"

	commongen "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/metal-stack/os-metal-extension/pkg/generator/ignition"
	"k8s.io/apimachinery/pkg/util/runtime"
)

var (
	cmd = "/usr/bin/env bash %s"

	//go:embed templates/*
	templates embed.FS
)

// IgnitionGenerator is the generator which will generete the ignition userdata and provider-specific shell script parts for the cloud config downloader.
func IgnitionGenerator() commongen.Generator {
	cloudInitTemplateString, err := templates.ReadFile("templates/cloud-init.sh.template")
	runtime.Must(err)

	cloudInitTemplate, err := ostemplate.NewTemplate("cloud-init").Parse(string(cloudInitTemplateString))
	runtime.Must(err)

	return ignition.New(cloudInitTemplate, ostemplate.DefaultUnitsPath, cmd, additionalValues)
}

func additionalValues(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error) {
	return nil, nil
}
