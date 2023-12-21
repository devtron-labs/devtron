## Automated Test suite integration in the CI step using devtron-ci.yaml

Users can run the test case using the Devtron dashboard or by including the test cases in the devtron.ci.yaml file in the source git repository. For reference, check: [https://github.com/kumarnishant/getting-started-nodejs/blob/master/devtron-ci.yaml](https://github.com/kumarnishant/getting-started-nodejs/blob/master/devtron-ci.yaml)

The test cases given in the script will run before the test cases given in the devtron.ci.yaml

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/creating-application/workflow-ci-pipeline/yaml.jpg)

| Field | Description |
| :--- | :--- |
| `version` | specify the version of yaml |
| `appliesTo` | applies the changes to a specified branch |
| `type` | branch type on which changes are to be applied, it can be **BRANCH\_FIXED** or **TAG\_PATTERN** |
| `value` | branch name on which changes are to be applied, it can take a value as the name of branch \(“master”\) or as a regular expression \("%d.%d.%d-rc"\) |
| `script` | A script which you want to execute, you can also execute the docker commands here |
| `beforeDockerBuildStages` | script to run before the docker build step |
| `afterDockerBuildStages` | script to run after the docker build step |
| `outputLocation` | The location where you want to see the output of the report of Test cases |