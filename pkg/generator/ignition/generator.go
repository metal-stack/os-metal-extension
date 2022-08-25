package ignition

import (
	"fmt"
	"text/template"

	"github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/generator"
	ostemplate "github.com/gardener/gardener/extensions/pkg/controller/operatingsystemconfig/oscommon/template"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
)

// IgnitionGenerator generates cloud-init scripts.
type IgnitionGenerator struct {
	cloudInitGenerator   generator.Generator
	cmd                  string
	additionalValuesFunc func(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error)
}

// Generate generates an ignition script from the given OperatingSystemConfig.
func (t *IgnitionGenerator) Generate(data *generator.OperatingSystemConfig) ([]byte, *string, error) {
	var cmd *string
	if data.Path != nil {
		c := fmt.Sprintf(t.cmd, *data.Path)
		cmd = &c
	}

	if data.Object.Spec.Purpose == extensionsv1alpha1.OperatingSystemConfigPurposeProvision {
		data, err := IgnitionFromOperatingSystemConfig(data)
		return data, cmd, err
	}

	return t.cloudInitGenerator.Generate(data)
}

// NewIgnitionGenerator creates a new CloudInitGenerator with the given units path.
func NewIgnitionGenerator(template *template.Template, unitsPath string, cmd string, additionalValuesFunc func(*extensionsv1alpha1.OperatingSystemConfig) (map[string]interface{}, error)) *IgnitionGenerator {
	return &IgnitionGenerator{
		cloudInitGenerator:   ostemplate.NewCloudInitGenerator(template, unitsPath, cmd, additionalValuesFunc),
		cmd:                  cmd,
		additionalValuesFunc: additionalValuesFunc,
	}
}
