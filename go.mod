module github.com/metal-stack/os-metal-extension

go 1.14

require (
	github.com/ajeddeloh/go-json v0.0.0-20200220154158-5ae607161559 // indirect
	github.com/ajeddeloh/yaml v0.0.0-20141224210557-6b16a5714269 // indirect
	github.com/coreos/container-linux-config-transpiler v0.9.0
	github.com/coreos/ignition v0.34.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/gardener/gardener v1.6.6
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/onsi/ginkgo v1.13.0
	github.com/onsi/gomega v1.10.1
	github.com/spf13/cobra v1.0.0
	github.com/vincent-petithory/dataurl v0.0.0-20191104211930-d1553a71de50 // indirect
	go4.org v0.0.0-20200411211856-f5505b9728dd // indirect
	k8s.io/apimachinery v0.17.6
	sigs.k8s.io/controller-runtime v0.5.4
)

replace (
	github.com/ajeddeloh/yaml => github.com/ajeddeloh/yaml v0.0.0-20170912190910-6b94386aeefd // indirect
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.6
)
