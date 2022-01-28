# Job and Cronjob chart

## 1. Job

A Job creates one or more Pods and will continue to retry execution of the Pods until a specified number of them successfully terminate. As pods successfully complete, the Job tracks the successful completions. When a specified number of successful completions is reached, the task (ie, Job) is complete. Deleting a Job will clean up the Pods it created. Suspeding a Job will delete its active Pods until the Job is resumed again.

## **Example:**

```yaml
kind: Job
jobConfigs:
    activeDeadlineSeconds: 120
    backoffLimit: 6
    completions: 1
    parallelism: 1
    suspend: false
    ttlSecondsAfterFinished: 100
```

| Key | Description |
| :--- | :--- |
| `activeDeadlineSeconds` | Another way to terminate a Job is by setting an active deadline. Do this by setting the activeDeadlineSeconds field of the Job to a number of seconds. The activeDeadlineSeconds applies to the duration of the job, no matter how many Pods are created. Once a Job reaches activeDeadlineSeconds, all of its running Pods are terminated and the Job status will become type: Failed with reason: DeadlineExceeded. |
| `backoffLimit` | There are situations where you want to fail a Job after some amount of retries due to a logical error in configuration etc. To do so, set backoffLimit to specify the number of retries before considering a Job as failed. The back-off limit is set by default to 6. Failed Pods associated with the Job are recreated by the Job controller with an exponential back-off delay (10s, 20s, 40s ...) capped at six minutes. The back-off count is reset when a Job's Pod is deleted or successful without any other Pods for the Job failing around that time. |
| `completions` | Jobs with fixed completion count - that is , jobs that have non null completions - can have a completion mode that is specified in completionMode. |
| `parallelism` | The requested parallelism can be set to any non-negative value. If it is unspecified, it defaults to 1. If it is specified as 0, then the Job is effectively paused until it is increased. |
| `suspend` | The suspend field is also optional. If it is set to true, all subsequent executions are suspended. This setting does not apply to already started executions. Defaults to false. |
| `ttlSecondsAfterFinished` | The TTL controller only supports Jobs for now. A cluster operator can use this feature to clean up finished Jobs (either Complete or Failed) automatically by specifying the ttlSecondsAfterFinished field of a Job, as in this example. The TTL controller will assume that a resource is eligible to be cleaned up TTL seconds after the resource has finished, in other words, when the TTL has expired. When the TTL controller cleans up a resource, it will delete it cascadingly, that is to say it will delete its dependent objects together with it. Note that when the resource is deleted, its lifecycle guarantees, such as finalizers, will be honored. |
| `kind` | As with all other Kubernetes config, a Job and cronjob needs apiVersion, kind.cronjob and job also needs a section fields which is optional . these fields specify to deploy which job (conjob or job) should be kept. by default, they are set job. |


## 2. Cronjob

A Cronjob creates Jobs on a repeating schedule , One Cronjob object is like one line of a crontab (cron table) file. It runs a job periodically on a given schedule, written in Cron format.
 Cronjobs are meant for performing regular scheduled actions such as backups, report generation, and so on. Each of those tasks should be configured to recur indefinitely (for example: once a day / week / month); you can define the point in time within that interval when the job should start.

 ## **Example:**

```yaml
kind: CronJob
cronjobConfigs:
    concurrencyPolicy: Allow
    failedJobsHistoryLimit: 1
    restartPolicy: OnFailure
    schedule: 32 8 * * *
    startingDeadlineSeconds: 100
    successfulJobsHistoryLimit: 3
    suspend: false
```

| Key | Descriptions |
| :--- | :--- |
| `concurrencyPolicy` | A CronJob is counted as missed if it has failed to be created at its scheduled time. For example, If concurrencyPolicy is set to Forbid and a CronJob was attempted to be scheduled when there was a previous schedule still running, then it would count as missed,`Acceptable values: Allow / Forbid`. |
| `failedJobsHistoryLimit` | The failedJobsHistoryLimit fields are optional. These fields specify how many completed and failed jobs should be kept. By default, they are set to 3 and 1 respectively. Setting a limit to 0 corresponds to keeping none of the corresponding kind of jobs after they finish. |
| `restartPolicy` | The spec of a Pod has a restartPolicy field with possible values Always, OnFailure, and Never. The default value is Always.The restartPolicy applies to all containers in the Pod. restartPolicy only refers to restarts of the containers by the kubelet on the same node. After containers in a Pod exit, the kubelet restarts them with an exponential back-off delay (10s, 20s, 40s, …), that is capped at five minutes. Once a container has executed for 10 minutes without any problems, the kubelet resets the restart backoff timer for that container, `Acceptable values: Always / OnFailure / Never`. |
| `schedule` | To generate Cronjob schedule expressions, you can also use web tools like https://crontab.guru/. |
| `startingDeadlineSeconds` | If startingDeadlineSeconds is set to a large value or left unset (the default) and if concurrencyPolicy is set to Allow, the jobs will always run at least once. |
| `successfulJobsHistoryLimit` | The successfulJobsHistoryLimit fields are optional. These fields specify how many completed and failed jobs should be kept. By default, they are set to 3 and 1 respectively. Setting a limit to 0 corresponds to keeping none of the corresponding kind of jobs after they finish. |
| `suspend` | The suspend field is also optional. If it is set to true, all subsequent executions are suspended. This setting does not apply to already started executions. Defaults to false. |
| `kind` | As with all other Kubernetes config, a Job and cronjob needs apiVersion, kind.cronjob and job also needs a section fields which is optional . these fields specify to deploy which job (conjob or job) should be kept. by default, they are set cronjob. |
