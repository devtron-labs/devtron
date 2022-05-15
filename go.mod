module github.com/devtron-labs/devtron

go 1.15

require (
	github.com/Azure/azure-storage-blob-go v0.12.0
	github.com/Azure/go-autorest/autorest v0.11.19
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/argoproj/argo v2.4.1+incompatible
	github.com/argoproj/argo-cd v1.2.3
	github.com/argoproj/argo-cd/v2 v2.3.3
	github.com/argoproj/argo-workflows/v3 v3.3.5
	github.com/argoproj/gitops-engine v0.6.2 // indirect
	github.com/argoproj/pkg v0.11.1-0.20211203175135-36c59d8fafe0 // indirect
	github.com/aws/aws-sdk-go v1.42.50
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/casbin/casbin v1.9.1
	github.com/casbin/xorm-adapter v1.0.1-0.20190716004226-a317737a1007
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/davecgh/go-spew v1.1.1
	github.com/devtron-labs/authenticator v0.4.19
	github.com/elazarl/goproxy v0.0.0-20210110162100-a92cc753f88e // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-pg/pg v6.15.1+incompatible
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-redis/cache v6.4.0+incompatible // indirect
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-xorm/xorm v0.7.9 // indirect
	github.com/gobuffalo/envy v1.7.1 // indirect
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.7
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/wire v0.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/sessions v1.2.1
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/igm/sockjs-go v3.0.0+incompatible // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/juju/testing v0.0.0-20210324180055-18c50b0c2098 // indirect
	github.com/ktrysmt/go-bitbucket v0.9.32
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lib/pq v1.10.4
	github.com/microsoft/azure-devops-go-api/azuredevops v1.0.0-b5
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d
	github.com/otiai10/copy v1.0.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/posthog/posthog-go v0.0.0-20210610161230-cd4408afb35a
	github.com/prometheus/client_golang v1.12.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/satori/go.uuid v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.14.0
	github.com/tidwall/sjson v1.2.4
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xanzy/go-gitlab v0.55.1
	github.com/xeipuuv/gojsonschema v1.2.0
	go.uber.org/zap v1.21.0
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.28.0
	gopkg.in/go-playground/validator.v9 v9.30.0
	gopkg.in/igm/sockjs-go.v3 v3.0.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.3
	k8s.io/apimachinery v0.23.3
	k8s.io/client-go v11.0.1-0.20190820062731-7e43eff7c80a+incompatible
	k8s.io/helm v2.12.3+incompatible
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	mellium.im/sasl v0.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0
	xorm.io/core v0.7.2 // indirect
)

replace (
	github.com/devtron-labs/authenticator => github.com/nishant-d/authenticator v0.4.20
	github.com/go-check/check => github.com/go-check/check v0.0.0-20180628173108-788fd7840127
	k8s.io/api => k8s.io/api v0.23.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.23.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.23.1
	k8s.io/apiserver => k8s.io/apiserver v0.23.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.23.1
	k8s.io/client-go => k8s.io/client-go v0.23.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.23.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.23.1
	k8s.io/code-generator => k8s.io/code-generator v0.23.1
	k8s.io/component-base => k8s.io/component-base v0.23.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.23.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.23.1
	k8s.io/cri-api => k8s.io/cri-api v0.23.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.23.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.23.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.23.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.23.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.23.1
	k8s.io/kubectl => k8s.io/kubectl v0.23.1
	k8s.io/kubelet => k8s.io/kubelet v0.23.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.23.1
	k8s.io/metrics => k8s.io/metrics v0.23.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.23.1
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.23.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.23.1
)
