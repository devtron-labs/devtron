# Ingress Setup

After Devtron is installed, Devtron is accessible through service `devtron-service`.
If you want to access devtron through ingress, edit devtron-service and change the loadbalancer to ClusterIP. You can do this using `kubectl patch` command like :

```bash
kubectl patch -n devtroncd svc devtron-service -p '{"spec": {"ports": [{"port": 80,"targetPort": "devtron","protocol": "TCP","name": "devtron"}],"type": "ClusterIP","selector": {"app": "devtron"}}}'
```
 
After that create ingress by applying the ingress yaml file.
You can use [this yaml file](https://github.com/devtron-labs/devtron/blob/main/manifests/yamls/devtron-ingress.yaml) to create ingress to access devtron:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
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

You can access devtron from any host after applying this yaml. For k8s versions <1.19, [apply this yaml](https://github.com/devtron-labs/devtron/blob/main/manifests/yamls/devtron-ingress-legacy.yaml):

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
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

Optionally you also can access devtron through a specific host like :

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
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
        host: devtron.example.com
        path: /orchestrator
        pathType: ImplementationSpecific 
      - backend:
          service:
            name: devtron-service
            port:
              number: 80
       host: devtron.example.com
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

## Enable HTTPS For Devtron

Once ingress setup for devtron is done and you want to run Devtron over `https`, you need to add different annotations for different ingress controllers and load balancers.

### 1. Nginx Ingress Controller

In case of `nginx ingress controller`, add following annotations under `service.annotations` under nginx ingress controller to run devtron over `https`.

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



