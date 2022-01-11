# Ingress setup for devtron installation

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
      - backend:
          service:
            name: devtron-service
            port:
              number: 80
        path: /grafana
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
       Host: devtron.example.com
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

