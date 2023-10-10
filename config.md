#application name 
APP: "orchestrator"
MODE: "PROD" 


#port no of dashbord service and url of the dashbord service
DASHBOARD_PORT: "80"  
DASHBOARD_HOST: "dashboard-service.devtroncd"  

#service name and port of argocd service and namespace  
CD_HOST: "argocd-server.devtroncd"  
CD_PORT: "80"   
CD_NAMESPACE: "devtroncd"  
GITOPS_REPO_PREFIX: "devtron"


#service url of notifier microservice it will be used if you installed notification module present in devtron stack   
EVENT_URL: "http://notifier-service.devtroncd:80/notify" 

#service url of lens microservice it will be used if you installed ci-cd module present in devtron stack to show you the deployment metrics 
LENS_URL: "http://lens-service.devtroncd:80"
LENS_TIMEOUT: "300"  

#service url of kubelink microserivce it will be used for helm deploying application 
HELM_CLIENT_URL: kubelink-service:50051


#service url of nats microservice it will be used to  pass mesage between the microservice  
NATS_SERVER_HOST: "nats://devtron-nats.devtroncd:4222"

#service url of postgres microservice it will be used to  store the data   
PG_ADDR: "postgresql-postgresql.devtroncd"
PG_PORT: "5432"
PG_USER: "postgres"
PG_DATABASE: "orchestrator"

#service url of git-sensor microservice it will be used to  store the data   
GIT_SENSOR_TIMEOUT: "300" 
GIT_SENSOR_PROTOCOL: GRPC
GIT_SENSOR_URL: git-sensor-service.devtroncd:90

#this image is used for sync all the latest helm chats present in devtron chart store 
APP_SYNC_IMAGE: "quay.io/devtron/chart-sync:0e8c785e-373-16172"

#service url of argocd-dex-server  microservice it will be used to authenticet during sso login    
DEX_HOST: "http://argocd-dex-server.devtroncd"
DEX_PORT: "5556"
DEX_RURL: "http://argocd-dex-server.devtroncd:8080/callback"
DEX_URL: "http://argocd-dex-server.devtroncd:5556/dex"
CExpirationTime: "600"
JwtExpirationTime: "120"


#service url of imagescanner microservice it will be used if you install trivy or clair from devtron-stack manager
IMAGE_SCANNER_ENDPOINT: "http://image-scanner-service.devtroncd:80"

#printing which  level in logs allowed values are (-2,-1,0)  
LOG_LEVEL: "0"  
PG_LOG_QUERY: "true"

#these are used if you are using monitering stack provided by devtron 
GRAFANA_URL: "http://%s:%s@devtron-grafana.devtroncd/grafana"
GRAFANA_HOST: "devtron-grafana.devtroncd"
GRAFANA_PORT: "80"
GRAFANA_NAMESPACE: "devtroncd"
GRAFANA_ORG_ID: "2"  

#details to be used by argocd 
#service of argocd-server it willbe used if you install gitops module in devtron-stack  
ACD_URL: "argocd-server.devtroncd"
ACD_USERNAME: "admin"
ACD_USER: "admin"
ACD_CM: "argocd-cm"
ACD_NAMESPACE: "devtroncd"
ACD_TIMEOUT: "300"
ACD_SKIP_VERIFY: "true"
GIT_WORKING_DIRECTORY: "/tmp/gitops/"

#these config are used for configure post-build workflow 
CD_LIMIT_CI_CPU: "0.5"
CD_LIMIT_CI_MEM: "3G"
CD_REQ_CI_CPU: "0.5"
CD_REQ_CI_MEM: "1G"
CD_NODE_TAINTS_KEY: "dedicated"
CD_NODE_LABEL_SELECTOR: "kubernetes.io/os=linux"
CD_WORKFLOW_SERVICE_ACCOUNT: "cd-runner"
CD_NODE_TAINTS_VALUE: "ci"



DEFAULT_CD_ARTIFACT_KEY_LOCATION: "devtron/cd-artifacts"
CD_ARTIFACT_LOCATION_FORMAT: "%d/%d.zip"
DEFAULT_CD_NAMESPACE: "devtron-cd"
DEFAULT_CD_TIMEOUT: "3600"

#this variable will use for enableing build context in devtron
ENABLE_BUILD_CONTEXT: "true"

#this image is use for creating pod in which your image build will be done and these are can be configured to modify your ci-pod limit-request and adding nodeselectors and taint  
DEFAULT_CI_IMAGE: "quay.io/devtron/ci-runner:d8d774c3-138-16238"  

WF_CONTROLLER_INSTANCE_ID: "devtron-runner"
CI_LOGS_KEY_PREFIX: "ci-artifacts"
DEFAULT_NAMESPACE: "devtron-ci"
DEFAULT_TIMEOUT: "3600"
LIMIT_CI_CPU: "0.5"
LIMIT_CI_MEM: "3G"
REQ_CI_CPU: "0.5"
REQ_CI_MEM: "1G"
CI_NODE_TAINTS_KEY: ""
CI_NODE_TAINTS_VALUE: ""
CI_NODE_LABEL_SELECTOR: ""  #e.g. "purpose=ci"
CACHE_LIMIT: "5000000000"
DEFAULT_ARTIFACT_KEY_LOCATION: "devtron/ci-artifacts"
WORKFLOW_SERVICE_ACCOUNT: "ci-runner"
CI_ARTIFACT_LOCATION_FORMAT: "%d/%d.zip"

#these config are used if you configured logs bucket or cache bucket 
DEFAULT_BUILD_LOGS_KEY_PREFIX: "devtron"

#these are used if you enable blob storage configuration
MINIO_ENDPOINT: http://devtron-minio:9000 # if minio is enabled 
BLOB_STORAGE_ENABLED: "true"
BLOB_STORAGE_PROVIDER: "S3"
BLOB_STORAGE_S3_ENDPOINT: "http://devtron-minio.devtroncd:9000"
BLOB_STORAGE_S3_ENDPOINT_INSECURE: "true"
DEFAULT_BUILD_LOGS_BUCKET: "devtron-ci-log"
DEFAULT_CACHE_BUCKET: "devtron-ci-cache"
BLOB_STORAGE_S3_BUCKET_VERSIONED: "false"
BLOB_STORAGE_S3_BUCKET_VERSIONED: "true"
DEFAULT_CACHE_BUCKET_REGION: "us-west-2"
DEFAULT_CD_LOGS_BUCKET_REGION: "us-west-2"
BLOB_STORAGE_S3_ENDPOINT: ""
BLOB_STORAGE_S3_BUCKET_VERSIONED: "true"
ECR_REPO_NAME_PREFIX: "devtron/"


EXTERNAL_CI_PAYLOAD: "{\"ciProjectDetails\":[{\"gitRepository\":\"https://github.com/srj92/getting-started-nodejs.git\",\"checkoutPath\":\"./abc\",\"commitHash\":\"239077135f8cdeeccb7857e2851348f558cb53d3\",\"commitTime\":\"2019-10-31T20:55:21+05:30\",\"branch\":\"master\",\"message\":\"Update README.md\",\"author\":\"Suraj Gupta \"}],\"dockerImage\":\"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2\",\"digest\":\"test1\",\"dataSource\":\"ext\",\"materialType\":\"git\"}"
ENFORCER_CACHE: "true"
ENFORCER_CACHE_EXPIRATION_IN_SEC: "345600"
ENFORCER_MAX_BATCH_SIZE: "1"

DEVTRON_SECRET_NAME: "devtron-secret"
