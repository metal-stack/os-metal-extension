module github.com/metal-stack/os-metal-extension

go 1.13

require (
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/ajeddeloh/yaml v0.0.0-20141224210557-6b16a5714269 // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/coreos/container-linux-config-transpiler v0.9.0
	github.com/coreos/ignition v0.34.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/gardener/gardener v0.35.1-0.20200128130120-5b69a02f511a
	github.com/gardener/gardener-extensions v1.3.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr v1.30.1
	github.com/gobuffalo/packr/v2 v2.7.1
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb // indirect
	go4.org v0.0.0-20200104003542-c7e774b10ea0 // indirect
	k8s.io/apimachinery v0.16.6
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	github.com/ajeddeloh/yaml => github.com/ajeddeloh/yaml v0.0.0-20170912190910-6b94386aeefd // indirect
	github.com/census-instrumentation/opencensus-proto => github.com/census-instrumentation/opencensus-proto v0.2.1
	k8s.io/client-go => k8s.io/client-go v0.16.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.6
)
