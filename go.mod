module github.com/metal-pod/os-metal-extension

go 1.12

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/Masterminds/sprig v2.18.0+incompatible // indirect
	github.com/appscode/jsonpatch v0.0.0-20190108182946-7c0e3b262f30 // indirect
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/evanphx/json-patch v4.1.0+incompatible // indirect
	github.com/gardener/controller-manager-library v0.0.0-20190531111244-4db8db4aed9b // indirect
	github.com/gardener/external-dns-management v0.0.0-20190617090046-9aae9a268f22 // indirect
	github.com/gardener/gardener v0.0.0-20190614160235-a872956ad019
	github.com/gardener/gardener-extensions v0.0.0-20190617062402-946f575a6489
	github.com/gardener/machine-controller-manager v0.0.0-20190613181923-2c7567353864 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gobuffalo/packr v1.26.0
	github.com/gobuffalo/packr/v2 v2.4.0
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20190129154638-5b532d6fd5ef // indirect
	github.com/golang/mock v1.3.1 // indirect
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/google/uuid v1.1.1 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/spf13/cobra v0.0.5
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20190308202827-9d24e82272b4 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver v0.0.0-20190606210616-f848dc7be4a4 // indirect
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver v0.0.0-20190606205144-71ebb8303503 // indirect
	k8s.io/client-go v10.0.0+incompatible // indirect
	k8s.io/helm v2.14.1+incompatible // indirect
	k8s.io/klog v0.3.3 // indirect
	k8s.io/kube-aggregator v0.0.0-20190606205516-445c23e3c4b2 // indirect
	k8s.io/kube-openapi v0.0.0-20190202092118-df6fb93e6113 // indirect
	k8s.io/utils v0.0.0-20190607212802-c55fbcfc754a // indirect
	sigs.k8s.io/controller-runtime v0.1.11
	sigs.k8s.io/testing_frameworks v0.1.1 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190606204050-af9c91bd2759
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190606210616-f848dc7be4a4
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go => k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190606205516-445c23e3c4b2
)
