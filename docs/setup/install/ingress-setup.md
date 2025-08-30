# Ingress Setup

## Introduction

If you wish to use [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) as a means to access the Devtron services available in your cluster, you can configure it either during the installation or after the installation of Devtron.  

Refer the section relevant to you:
* [During Devtron Installation](#enabling-ingress-during-devtron-installation)
* [After Devtron Installation](#configuring-ingress-after-devtron-installation)
* [Reverse Proxy with External Nginx](#reverse-proxy-setup-with-external-nginx)

If you have successfully configured Ingress, refer [Post Ingress Setup](#enable-https-for-devtron).

---

## Enabling Ingress during Devtron Installation

If you are installing Devtron, you can enable Ingress either via [set flag](#using-set-flag) or by using [ingress-values.yaml](#using-ingress-values.yaml) to specify the desired Ingress settings.

### Using set flag

You can use the `--set` flag to specify the desired Ingress settings.

Here, we have added 5 configurations you can perform depending on your requirements:
* [Only Basic Configuration](#only-basic-configuration)
* [Configuration Including Labels](#configuration-including-labels)
* [Configuration Including Annotations](#configuration-including-annotations)
* [Configuration Including TLS Settings](#configuration-including-tls-settings)
* [Comprehensive Configuration](#comprehensive-configuration)

#### Only Basic Configuration

To enable Ingress and set basic parameters, use the following command:

```bash
helm install devtron devtron/devtron-operator -n devtroncd \
  --set components.devtron.ingress.enabled=true \
  --set components.devtron.ingress.className=nginx \
  --set components.devtron.ingress.host=devtron.example.com
```

#### Configuration Including Labels

To add labels to the Ingress resource, use the following command:

```bash
helm install devtron devtron/devtron-operator -n devtroncd \
  --set components.devtron.ingress.enabled=true \
  --set components.devtron.ingress.className=nginx \
  --set components.devtron.ingress.host=devtron.example.com \
  --set components.devtron.ingress.labels.env=production
```

#### Configuration Including Annotations

To add annotations to the Ingress resource, use the following command:

```bash
helm install devtron devtron/devtron-operator -n devtroncd \
  --set components.devtron.ingress.enabled=true \
  --set components.devtron.ingress.className=nginx \
  --set components.devtron.ingress.host=devtron.example.com \
  --set components.devtron.ingress.annotations."kubernetes\.io/ingress\.class"=nginx \
  --set components.devtron.ingress.annotations."nginx\.ingress\.kubernetes\.io\/app-root"="/dashboard"
```

#### Configuration Including TLS Settings

To configure TLS settings, including `secretName` and `hosts`, use the following command:

```bash
helm install devtron devtron/devtron-operator -n devtroncd \
  --set components.devtron.ingress.enabled=true \
  --set components.devtron.ingress.className=nginx \
  --set components.devtron.ingress.host=devtron.example.com \
  --set components.devtron.ingress.tls[0].secretName=devtron-tls \
  --set components.devtron.ingress.tls[0].hosts[0]=devtron.example.com
```

#### Comprehensive Configuration

To include all the above settings in a single command, use:

```bash
helm install devtron devtron/devtron-operator -n devtroncd \
  --set components.devtron.ingress.enabled=true \
  --set components.devtron.ingress.className=nginx \
  --set components.devtron.ingress.host=devtron.example.com \
  --set components.devtron.ingress.annotations."kubernetes\.io/ingress\.class"=nginx \
  --set components.devtron.ingress.annotations."nginx\.ingress\.kubernetes\.io\/app-root"="/dashboard" \
  --set components.devtron.ingress.labels.env=production \
  --set components.devtron.ingress.pathType=ImplementationSpecific \
  --set components.devtron.ingress.tls[0].secretName=devtron-tls \
  --set components.devtron.ingress.tls[0].hosts[0]=devtron.example.com
```


### Using ingress-values.yaml

As an alternative to the [set flag](#using-set-flag) method, you can enable Ingress using `ingress-values.yaml` instead. 

Create an `ingress-values.yaml` file. You may refer the below format for an advanced ingress configuration which includes labels, annotations, secrets, and many more.

```yml
components:
  devtron:
    ingress:
      enabled: true
      className: nginx
      labels: {}
        # env: production
      annotations: {}
        # nginx.ingress.kubernetes.io/app-root: /dashboard
      pathType: ImplementationSpecific
      host: devtron.example.com
      tls: []
    #    - secretName: devtron-info-tls
    #      hosts:
    #        - devtron.example.com
```

Once you have the `ingress-values.yaml` file ready, run the following command:

```bash
helm install devtron devtron/devtron-operator -n devtroncd  --reuse-values  -f ingress-values.yaml
```

---

## Configuring Ingress after Devtron Installation

After Devtron is installed, Devtron is accessible through `devtron-service`. If you wish to access Devtron through ingress, you'll need to modify this service to use a ClusterIP instead of a LoadBalancer.

You can do this using the `kubectl patch` command:

```bash
kubectl patch -n devtroncd svc devtron-service -p '{"spec": {"ports": [{"port": 80,"targetPort": "devtron","protocol": "TCP","name": "devtron"}],"type": "ClusterIP","selector": {"app": "devtron"}}}'
```
 
Next, create ingress to access Devtron by applying the `devtron-ingress.yaml` file. The file is also available on this [link](https://github.com/devtron-labs/devtron/blob/main/manifests/yamls/devtron-ingress.yaml). You can access Devtron from any host after applying this yaml. 

```yml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations: 
    nginx.ingress.kubernetes.io/app-root: /dashboard
  labels:
    app: devtron
    release: devtron
  name: devtron-ingress
  namespace: devtroncd
spec:
  ingressClassName: nginx
  rules:
  - http:
      paths:
      - backend:
          service:
            name: devtron-service
            port:
              number: 80
        path: /orchestrator
        pathType: ImplementationSpecific 
      - backend:
          service:
            name: devtron-service
            port:
              number: 80
        path: /dashboard
        pathType: ImplementationSpecific
      - backend:
          service:
            name: devtron-service
            port:
              number: 80
        path: /grafana
        pathType: ImplementationSpecific  
```        

For k8s versions < 1.19, [apply this yaml](https://github.com/devtron-labs/devtron/blob/main/manifests/yamls/devtron-ingress-legacy.yaml):

```yml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations: 
    nginx.ingress.kubernetes.io/app-root: /dashboard
  labels:
    app: devtron
    release: devtron
  name: devtron-ingress
  namespace: devtroncd
spec:
  rules:
  - http:
      paths:
      - backend:
          serviceName: devtron-service
          servicePort: 80
        path: /orchestrator
      - backend:
          serviceName: devtron-service
          servicePort: 80
        path: /dashboard
        pathType: ImplementationSpecific  
```        

Optionally, you also can access Devtron through a specific host by running the following YAML file:

```yml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations: 
    nginx.ingress.kubernetes.io/app-root: /dashboard
  labels:
    app: devtron
    release: devtron
  name: devtron-ingress
  namespace: devtroncd
spec:
  ingressClassName: nginx
  rules:
    - host: devtron.example.com
      http:
        paths:
          - backend:
              service:
                name: devtron-service
                port:
                  number: 80
            path: /orchestrator
            pathType: ImplementationSpecific
          - backend:
              service:
                name: devtron-service
                port:
                  number: 80
            path: /dashboard
            pathType: ImplementationSpecific
          - backend:
              service:
                name: devtron-service
                port:
                  number: 80
            path: /grafana
            pathType: ImplementationSpecific
```

---

## Enable HTTPS For Devtron

Once Ingress setup for Devtron is done and you want to run Devtron over `https`, you need to add different annotations for different ingress controllers and load balancers.

### 1. Nginx Ingress Controller

In case of `nginx ingress controller`, add the following annotations under `service.annotations` under nginx ingress controller to run devtron over `https`.

(i) Amazon Web Services (AWS)

If you are using AWS cloud, add the following annotations under `service.annotations` under nginx ingress controller.

```bash
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: "http"
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "443"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: "<acm-arn-here>"
```

(ii) Digital Ocean

If you are using Digital Ocean cloud, add the following annotations under `service.annotations` under nginx ingress controller.

```bash
annotations:
  service.beta.kubernetes.io/do-loadbalancer-protocol: "http"
  service.beta.kubernetes.io/do-loadbalancer-tls-ports: "443"
  service.beta.kubernetes.io/do-loadbalancer-certificate-id: "<your-certificate-id>"
  service.beta.kubernetes.io/do-loadbalancer-redirect-http-to-https: "true"
```

### 2. AWS Application Load Balancer (AWS ALB)

In case of AWS application load balancer, add following annotations under `ingress.annotations` to run devtron over `https`.

```bash
  annotations:
    kubernetes.io/ingress.class: alb
    alb.ingress.kubernetes.io/listen-ports: '[{"HTTP": 80}, {"HTTPS": 443}]'
    alb.ingress.kubernetes.io/certificate-arn: "<acm-arn-here>"
```

### 3. Azure Application Gateway

In case of AWS application load balancer, the following annotations need to be added under `ingress.annotations` to run devtron over `https`.

```bash
 annotations:
  kubernetes.io/ingress.class: "azure/application-gateway"
  appgw.ingress.kubernetes.io/backend-protocol: "http"
  appgw.ingress.kubernetes.io/ssl-redirect: "true"
  appgw.ingress.kubernetes.io/appgw-ssl-certificate: "<name-of-appgw-installed-certificate>"
```
For an Ingress resource to be observed by AGIC (Application Gateway Ingress Controller) must be annotated with kubernetes.io/ingress.class: azure/application-gateway. Only then AGIC will work with the Ingress resource in question.

> Note: Make sure NOT to use port 80 with HTTPS and port 443 with HTTP on the Pods.

---

## Reverse Proxy Setup with External Nginx

If you are running Devtron in Kubernetes and want to expose it through an external nginx server (running on the host machine or another server), you can configure nginx to reverse proxy requests to the Devtron service.

### Service Exposure

First, expose the Devtron service using a LoadBalancer (with MetalLB) or NodePort:

```bash
# Check your Devtron service
kubectl get svc -n devtroncd devtron-service

# Example output:
# NAME              TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)        AGE
# devtron-service   LoadBalancer   10.96.10.29    172.25.0.1    80:31084/TCP   13h
```

### Nginx Configuration

Configure your external nginx to proxy all requests to the Devtron service. **Important**: Route all requests to the devtron-service, as Devtron internally manages routing to different components including the dashboard.

```nginx
server {
    listen 80;
    server_name your-domain.com;  # Replace with your domain

    # Route all requests to Devtron service
    location / {
        proxy_pass http://172.25.0.1;  # Replace with your Devtron service IP
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Disable caching for dynamic content
        proxy_cache_bypass $http_pragma;
        proxy_cache_revalidate on;
        proxy_no_cache $http_pragma $http_authorization;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
    }

    # Optional: Health check endpoint
    location /health {
        access_log off;
        proxy_pass http://172.25.0.1/health;
        proxy_set_header Host $host;
    }
}

# Define connection upgrade for WebSocket support
map $http_upgrade $connection_upgrade {
    default upgrade;
    '' close;
}
```

### HTTPS Configuration

For production environments, enable HTTPS:

```nginx
server {
    listen 443 ssl http2;
    server_name your-domain.com;

    # SSL Configuration
    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Route all requests to Devtron service
    location / {
        proxy_pass http://172.25.0.1;  # Replace with your Devtron service IP
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        
        # Timeouts
        proxy_connect_timeout       60s;
        proxy_send_timeout          60s;
        proxy_read_timeout          60s;
    }
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}
```

### Common Issues and Solutions

#### 1. API Endpoints Returning 404

**Problem**: Getting 404 errors for API endpoints like `/autocomplete`, `/api/*`, etc.

**Solution**: Ensure all requests are proxied to the devtron-service IP. Do not try to separate `/dashboard`, `/orchestrator`, `/api` paths to different services. Devtron handles internal routing.

```nginx
# ❌ Incorrect - Don't split paths
location /dashboard/ {
    proxy_pass http://172.25.0.2/;  # Dashboard service
}
location /orchestrator {
    proxy_pass http://172.25.0.1;   # Devtron service
}

# ✅ Correct - Route everything to devtron service
location / {
    proxy_pass http://172.25.0.1;   # Devtron service only
}
```

#### 2. WebSocket Connection Issues

**Problem**: Real-time features not working (live logs, terminal, etc.).

**Solution**: Ensure WebSocket support is configured:

```nginx
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection $connection_upgrade;
```

#### 3. Large Request Body Issues

**Problem**: Errors when uploading large files or making large API requests.

**Solution**: Increase nginx limits:

```nginx
client_max_body_size 100M;
proxy_request_buffering off;
```

### Verification

After configuring nginx, verify the setup:

1. **Health Check**: `curl http://your-domain.com/health` should return `{"code":200,"result":"OK"}`

2. **Dashboard Access**: `http://your-domain.com/dashboard` should load the Devtron dashboard

3. **API Access**: API endpoints like `http://your-domain.com/orchestrator/app/autocomplete` should work without 404 errors

### Alternative: Using Subdirectory

If you need to serve Devtron from a subdirectory (e.g., `your-domain.com/devtron`), you can configure it as follows:

```nginx
location /devtron/ {
    proxy_pass http://172.25.0.1/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_set_header X-Forwarded-Prefix /devtron;
    
    # Remove the /devtron prefix when forwarding to backend
    rewrite ^/devtron(.*)$ $1 break;
}
```

> **Note**: When using subdirectory configuration, ensure the Devtron installation is configured with the correct base URL to handle relative paths properly.