module github.com/metal-stack/os-metal-extension

go 1.16

require (
	github.com/ajeddeloh/go-json v0.0.0-20200220154158-5ae607161559 // indirect
	github.com/ajeddeloh/yaml v0.0.0-20141224210557-6b16a5714269 // indirect
	github.com/coreos/container-linux-config-transpiler v0.9.0
	github.com/coreos/ignition v0.35.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/gardener/gardener v1.15.7
	github.com/go-logr/logr v0.1.0
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/onsi/ginkgo v1.16.1
	github.com/onsi/gomega v1.11.0
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/vincent-petithory/dataurl v0.0.0-20191104211930-d1553a71de50 // indirect
	go4.org v0.0.0-20201209231011-d4a079459e60 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	honnef.co/go/tools v0.1.3 // indirect
	k8s.io/apimachinery v0.18.10
	sigs.k8s.io/controller-runtime v0.6.3
)

replace (
	github.com/ajeddeloh/yaml => github.com/ajeddeloh/yaml v0.0.0-20170912190910-6b94386aeefd // indirect
	k8s.io/client-go => k8s.io/client-go v0.18.10
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.10
)
