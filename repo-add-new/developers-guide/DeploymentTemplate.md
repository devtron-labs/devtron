#Global Deployment Template

## Create (db table - charts) :- 

1. charts.Values = values.yaml(from ref chart) +  charts.global_override
2. charts.GlobalOverride = charts.global_override + global_override(different for different cases, find details below)  
3. charts.ReleaseOverrides = release-values.yaml(from ref chart)
4. charts.PipelineOverrides = pipeline-values.yaml(from ref chart)
5. charts.ImageDescriptorTemplate =  image_descriptor_template.json (from ref chart)

### GlobalOverride values from UI - 

1. Case 1, new app = app-values.yaml + env-values.yaml (both from ref chart)
2. Case 2, old app = Value of charts.global_override (from db)

#Env Template

## Create (db table - chart_env_config_override) :- 

1. Case 1, env is overridden 
- chart_env_config_override.env_override_yaml = chart_env_config_override.env_override_yaml(from UI)
- In this case chart_env_config_override.env_override_yaml is used for release 
2. Case 2, env is not overridden 
- chart_env_config_override.env_override_yaml = "{}"
- In this case charts.global_override is used for release

#Pipeline Strategy

Config is stored in pipeline_strategy.Config

#Deployment Triggered - 

## Create (db table - pipeline_config_override) :- 

1. pipeline_config_override.pipeline_override_yaml = rendered value of image descriptor template
2. pipeline_config_override.merged_values_yaml = All values merged which are going to be committed on git (Values - charts.global_override + chart_env_config_override.env_override_yaml + cm/secrets(converted to json) + pipeline_strategy.config)