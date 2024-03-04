# Semgrep

Semgrep is a fast, open source, static analysis engine for finding bugs, detecting dependency vulnerabilities, and enforcing code standards.

**Prerequisite**: Make sure you have set up an account in `Semgrep` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Semgrep** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| SemgrepAppToken | String | App token of Semgrep. If it is provided, this token will be used, otherwise it will be picked from Global Secret. |
| PrefixAppNameInSemgrepBranchName | Bool | Enter either `true` or `false` accordingly whether you want app name to be reflected with a branch name. If it is `true`, it will add app name with branch name. E.g., {SemgrepAppName}-{branchName} |
| UseCommitAsSemgrepBranchName | Bool | Enter either `true` or `false` accordingly whether you want app name to be reflected with commit hash. If it is `true`, it will add app name with commit hash. E.g., {SemgrepAppName}-{CommitHash}. |
| SemgrepAppName | String | App name for Semgrep. If it is provided, and `PrefixAppNameInSemgrepBranchName` is true, then this will be prefixed with branch name/commit hash.|
| ExtraCommandArguments | String | Extra command arguments for Semgrep CI command. E.g., Input: --json --dry-run. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.