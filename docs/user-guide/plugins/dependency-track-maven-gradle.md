# Dependency Track for Maven & Gradle

Configuring `Dependency Track for NodeJs` in pre-build or post build task creates a bill of materials from Maven & Gradle projects and environments and uploads it to D-track for [Component Analysis](https://owasp.org/www-community/Component_Analysis) to identify and reduce risk in the software supply chain.


**Prerequisite**: Make sure you have set up an account in `dependency track` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Dependency track for Maven & Gradle** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| BuildToolType | String | Type of build tool your project is using. E.g., Maven, or Gradle |
| DTrackEndpoint | String | API endpoint of your dependency track account |
| DTrackProjectName | String | Name of your dependency track project |
| DTrackProjectVersion | String | Version of dependency track project |
| DTrackApiKey | String | API key of your dependency track account |
| CheckoutPath | String | Checkout path of Git material |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.