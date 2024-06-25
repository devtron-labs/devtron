# K6 Load Testing

K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.

**Prerequisite**: Make sure you have set up an account in `k6.io` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **K6 Load Testing** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| RelativePathToScript | String | Checkout path + script path along with script name |
| PrometheusUsername | String | Username of Prometheus account |
| PrometheusApiKey | String | API key of Prometheus account |
| PrometheusRemoteWriteEndpoint | String | Remote write endpoint of Prometheus account |
| OutputType | String | `Log` or `Prometheus` |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.
