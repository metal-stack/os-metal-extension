module github.com/metal-pod/os-metal-extension

go 1.12

require (
	github.com/ajeddeloh/go-json v0.0.0-20170920214419-6a2fe990e083 // indirect
	github.com/ajeddeloh/yaml v0.0.0-20141224210557-6b16a5714269 // indirect
	github.com/alecthomas/units v0.0.0-20190717042225-c3de453c63f4 // indirect
	github.com/coreos/container-linux-config-transpiler v0.9.0
	github.com/coreos/go-systemd v0.0.0-20190617083831-1652836e9bdc // indirect
	github.com/coreos/ignition v0.32.0 // indirect
	github.com/gardener/gardener v0.0.0-20190827050434-cdafc6bd869f
	github.com/gardener/gardener-extensions v0.0.0-20190828072712-cf446972f37d
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/packr v1.25.0
	github.com/gobuffalo/packr/v2 v2.5.2
	github.com/golang/mock v1.3.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/prometheus/client_golang v1.1.0 // indirect
	github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb // indirect
	golang.org/x/net v0.0.0-20190813141303-74dc4d7220e7 // indirect
	golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	sigs.k8s.io/controller-runtime v0.2.0-beta.2
)

replace (
	github.com/ajeddeloh/yaml => github.com/ajeddeloh/yaml v0.0.0-20170912190910-6b94386aeefd // indirect
	github.com/gardener/gardener => github.com/metal-pod/gardener v0.0.0-20190827131320-58dad7be7444
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab //kubernetes-1.14.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed // kubernetes-1.14.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190313205120-d7deff9243b1 // kubernetes-1.14.0
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible // kubernetes-1.14.0
)
