module github.com/metal-pod/os-metal-extension

go 1.13

require (
	git.apache.org/thrift.git v0.0.0-20180902110319-2566ecd5d999 // indirect
	github.com/Azure/go-autorest v11.5.0+incompatible // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/ajeddeloh/yaml v0.0.0-20141224210557-6b16a5714269 // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/census-instrumentation/opencensus-proto v0.2.1 // indirect
	github.com/coreos/container-linux-config-transpiler v0.9.0
	github.com/coreos/ignition v0.34.0 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/gardener/gardener v0.35.0
	github.com/gardener/gardener-extensions v1.1.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr v1.25.0
	github.com/gobuffalo/packr/v2 v2.5.2
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/onsi/ginkgo v1.10.1
	github.com/onsi/gomega v1.7.0
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb // indirect
	go4.org v0.0.0-20200104003542-c7e774b10ea0 // indirect
	k8s.io/api v0.16.6
	k8s.io/apimachinery v0.16.6
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	github.com/ajeddeloh/yaml => github.com/ajeddeloh/yaml v0.0.0-20170912190910-6b94386aeefd // indirect
	github.com/census-instrumentation/opencensus-proto => github.com/census-instrumentation/opencensus-proto v0.2.1
	github.com/gardener/gardener-extensions => github.com/metal-pod/gardener-extensions v0.0.0-20200122160011-3274fdbed6b0
	k8s.io/client-go => k8s.io/client-go v0.16.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.6
)
