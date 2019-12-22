module github.com/metal-pod/os-metal-extension

go 1.13

require (
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/ajeddeloh/yaml v0.0.0-20141224210557-6b16a5714269 // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/coreos/container-linux-config-transpiler v0.9.0
	github.com/coreos/ignition v0.34.0 // indirect
	github.com/gardener/gardener v0.33.1-0.20191217084546-948979065a54
	github.com/gardener/gardener-extensions v1.1.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr v1.25.0
	github.com/gobuffalo/packr/v2 v2.5.2
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb // indirect
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	k8s.io/apimachinery v0.0.0-20191004074956-c5d2f014d689
	sigs.k8s.io/controller-runtime v0.2.2
)

replace github.com/ajeddeloh/yaml => github.com/ajeddeloh/yaml v0.0.0-20170912190910-6b94386aeefd // indirect

replace github.com/census-instrumentation/opencensus-proto => github.com/census-instrumentation/opencensus-proto v0.2.1
