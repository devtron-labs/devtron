module github.com/devtron-labs/devtron

go 1.15

require (
	github.com/Azure/azure-storage-blob-go v0.12.0
	github.com/Azure/go-autorest/autorest v0.11.19
	github.com/Azure/go-autorest/autorest/adal v0.9.13
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Pallinder/go-randomdata v1.2.0
	github.com/argoproj/argo v2.4.1+incompatible
	github.com/argoproj/argo-cd v1.2.3
	github.com/argoproj/argo-cd/v2 v2.3.3 // indirect
	github.com/aws/aws-sdk-go v1.38.49
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/casbin/casbin v1.9.1
	github.com/casbin/xorm-adapter v1.0.1-0.20190716004226-a317737a1007
	github.com/colinmarc/hdfs v1.1.4-0.20180805212432-9746310a4d31 // indirect
	github.com/coreos/go-oidc v2.2.1+incompatible
	github.com/cyphar/filepath-securejoin v0.2.2 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/devtron-labs/authenticator v0.4.19
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20210110162100-a92cc753f88e // indirect
	github.com/emicklei/go-restful v2.11.0+incompatible // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-pg/pg v6.15.1+incompatible
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.16.0 // indirect
	github.com/go-redis/cache v6.4.0+incompatible // indirect
	github.com/go-redis/redis v6.15.6+incompatible // indirect
	github.com/go-resty/resty/v2 v2.7.0
	github.com/go-xorm/xorm v0.7.9 // indirect
	github.com/gobuffalo/envy v1.7.1 // indirect
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.2.0
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/wire v0.3.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/gorilla/sessions v1.2.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-uuid v1.0.1 // indirect
	github.com/hashicorp/golang-lru v0.5.3 // indirect
	github.com/igm/sockjs-go v3.0.0+incompatible // indirect
	github.com/jcmturner/gofork v1.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/juju/testing v0.0.0-20210324180055-18c50b0c2098 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/ktrysmt/go-bitbucket v0.9.27
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/lib/pq v1.2.0
	github.com/microsoft/azure-devops-go-api/azuredevops v1.0.0-b5
	github.com/nats-io/nats.go v1.13.0
	github.com/otiai10/copy v1.0.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/posthog/posthog-go v0.0.0-20210610161230-cd4408afb35a
	github.com/prometheus/client_golang v1.11.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/rogpeppe/go-internal v1.5.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.11.0
	github.com/tidwall/sjson v1.1.7
	github.com/valyala/fasttemplate v1.1.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/xanzy/go-gitlab v0.16.1
	github.com/xeipuuv/gojsonschema v1.2.0
	go.uber.org/zap v1.19.1
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/validator.v9 v9.30.0
	gopkg.in/igm/sockjs-go.v3 v3.0.0
	gopkg.in/jcmturner/goidentity.v2 v2.0.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.1
	k8s.io/apimachinery v0.23.1
	k8s.io/client-go v11.0.1-0.20190820062731-7e43eff7c80a+incompatible
	k8s.io/helm v2.12.3+incompatible
	k8s.io/utils v0.0.0-20210930125809-cb0fa318a74b
	mellium.im/sasl v0.2.1 // indirect
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
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
