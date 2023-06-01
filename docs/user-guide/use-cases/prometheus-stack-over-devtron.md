# Prerequisite to setup Prometheus Stack over Devtron

## Introduction

Prometheus is an open-source technology designed to provide monitoring and alerting functionality for cloud-native environments, including Kubernetes. It can collect and store metrics as time-series data, recording information with a timestamp. It can also collect and record labels, which are optional key-value pairs.

### **Open Devtron dashboard and select chartstore from side panel**

### Search for Prometheus and choose the kube-prometheus-stack chart of prometheus-community repo.
![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/prometheus-stack/prome.png)


### ***We have to configure below values in values.yaml***
```yaml
1. search for kube-state-metrics and add below data

   kube-state-metrics:
    metricLabelsAllowlist:
      - pods=[*]

2. search for podMonitorSelectorNilUsesHelmValues and make it false 
 podMonitorSelectorNilUsesHelmValues: false

3. search for serviceMonitorSelectorNilUsesHelmValues and make it false
serviceMonitorSelectorNilUsesHelmValues: false

```





### **After Configuring the values.yaml click on Update and Deploy button**

Here we can see the all the resources of this stack are in healthy state . 

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/use-cases/prometheus-stack/prometheus-demo.png)

