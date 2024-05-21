### /pkg - consists of all business logic 

#### /pkg/devtronResource/history 
package for handling all task runs(deployments etc.) history related business logic. 
Currently, has cdPipeline's deployment history related code (for getting history for a pipeline and getting history configuration for comparing).
Uses pkg/pipeline/history to get the configuration data as old apis and used /pkg/devtronResource/read for getting the run source data (currently only release supported). 
