module github.com/tmax-cloud/l2c-operator

go 1.13

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gorilla/mux v1.7.4
	github.com/operator-framework/operator-sdk v0.17.1
	github.com/spf13/pflag v1.0.5
	github.com/tektoncd/pipeline v0.15.2
	github.com/tidwall/gjson v1.6.0
	k8s.io/api v0.18.7-rc.0
	k8s.io/apimachinery v0.18.7-rc.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/kube-aggregator v0.17.3
	knative.dev/pkg v0.0.0-20200711004937-22502028e31a
	sigs.k8s.io/controller-runtime v0.6.1
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.3.2+incompatible // Required by OLM
	knative.dev/pkg => knative.dev/pkg v0.0.0-20200810223505-473bba04ee7f
	sigs.k8s.io/controller-runtime => sigs.k8s.io/controller-runtime v0.5.2
)

// Pin k8s deps to 1.17.6
replace (
	k8s.io/api => k8s.io/api v0.17.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.6
	k8s.io/client-go => k8s.io/client-go v0.17.6
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200410145947-bcb3869e6f29
)
