# GoLang-migrate

Migrate reads migrations from sources file and applies them in correct order to a database.

**Prerequisite**: Make sure you have SQL files in format used by the golang-migrate tool.

**official-documentation**: https://github.com/golang-migrate/migrate
**postgres-example**: https://github.com/golang-migrate/migrate/tree/master/database/postgres

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage). or 
2. Click **+ Add task**.
3. Select **GoLang-migrate** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| DB_TYPE | String | Currently this plugin support postgres,mongodb,mongodb+srv,mysql,sqlserver. |
| DB_HOST | String | The hostname, service endpoint or IP address of the database server. |
| DB_PORT | String | The port number on which the database server is listening. |
| DB_NAME | String | The name of the specific database instance you want to connect to. |
| DB_USER | String | The username required to authenticate to the database.|
| DB_PASSWORD | String | The password required to authenticate to the database. |
| SCRIPT_LOCATION | String | Location of SQL files that need to be run on desired database. |
| MIGRATE_IMAGE | String | Docker image of golang-migrate default:migrate/migrate. |
| MIGRATE_TO_VERSION | String | migrate to which version of sql script need to be run on desired database(default: 0 is for all files in directory). |
| PARAM | String | extra params that runs with db queries. |
| POST_COMMAND | String | post commands that runs at the end of script. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.


### Notes
- Use `in-cluster/ Execute tasks in application environment` feature in `pre-deploy` or `post-deploy`, in case when the database service is not reachable or accessible from devtron cluster.
- In case the `DB_TYPE` is not supported for your database, then use `POST_COMMAND` as 

   ```
   POST_COMMAND: 
        docker run -v $PWD:$PWD $MIGRATE_IMAGE   -path $PWD/$SCRIPT_LOCATION -database <myDB-connection-string>" goto $MIGRATE_TO_VERSION;
   ```
- use `DB_PASSWORD` with `scope-variable` feature for more security.
