# Application Metrics



## CPU Usage Metrics

CPU usage is a utilization metric that shows the overall utilization of cpu by an application. For maintaining the business applications and services, we need to monitor  their cpu utilization constantly. We can do that just by looking on cpu usage metrics. Sometimes we also need to scale our application on the basis of cpu utilization. The autoscaler uses this metric to scale up the application  if the cpu utilization reaches the threshold. Devtron provides this metric for each application by default i.e. you do not need to enable “Show application metric”.

## Memory Usage Metrics

 Memory usage is a utilization metric that shows the overall utilization of memory by an application. Here by saying memory, it means RAM. Memory plays an important role in running an application, so we can use this metric to monitor memory utilization and can increase memory allocation to the application so that the application run smoothly. Autoscaler uses this metric to scale the application based on memory utilization of the application. Like CPU metrics, you'll get this application metric without enabling "Show application metrics"


 ## Throughput Metrics

 This application metrics indicates the number of request processed by an application per minute. Higher the throughput 
 i.e. number of request processed, performance of application is considered better. You’ll have to enable “Show application metrics” to see this metrics.

 ## Status Code Metrics

 Status Code Metrics

This metrics indicates the  application’s response to client’s request with a specific status code i.e 1xx(Communicate transfer protocol leve information), 2xx(Client’s request was accepted successfully), 3xx(Client must take some additional action to complete their request), 4xx(Client side error) or 5xx(Server side error).  To get this metric, you have to select the status code from drop-down in the throughput metric and you’ll get the metrics for that specific status code.


## Latency Metrics

This metric shows the latency for an application. Latency measures the delay between an action and a response. In other words we can say, latency measures how long it takes for and application to process a request. If your application experiences high latency, you can use this metric to diagnose and resolve the issue. You’ll have to enable “Show application metrics” to get latency metric for the application.


