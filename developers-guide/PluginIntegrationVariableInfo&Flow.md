##Plugin variable data taken from user - plugin creation
``` 
Db table - plugin_step_variable
 
Input Variables:
    
	Name                  string 
	Format                PluginStepVariableFormatType 
	Description           string                      
	IsExposed             bool                       
	AllowEmptyValue       bool     
	ValueType             PluginStepVariableValueType                   
	DefaultValue          string                                          
	ReferenceVariableName string
	PreviousStepIndex     int                 

Output Variables:
   
	Name                  string
	Format                PluginStepVariableFormatType 
	Description           string                      
	IsExposed             bool                       
	AllowEmptyValue       bool     
	               
where, 

PluginStepVariableFormatType = ["STRING", "BOOL", "NUMBER", "DATE"]
PluginStepVariableValueType = ["NEW", "FROM_PREVIOUS_STEP", "GLOBAL"]
```

Note - Same data will be sent back in get api.

##Pipeline stage step variable data taken from user - step creation

###Case 1 - Inline Step Type

```
Db table - pipeline_stage_step_variable

Input Variables:
   
   	Name                   string
   	Format                 PipelineStageStepVariableFormatType 
	Description            string                     
	Value                  string                       
	ValueType              PipelineStageStepVariableValueType  
	PreviousStepIndex      int                                 
	ReferenceVariableName  string                             
	ReferenceVariableStage PipelineStageType 


Output Variables:

	Name                   string                         
	Format                 PipelineStageStepVariableFormatType 
	Description            string     
	
where,

PipelineStageStepVariableFormatType = ["STRING", "BOOL", "NUMBER", "DATE"]
PipelineStageStepVariableValueType = ["NEW", "FROM_PREVIOUS_STEP", "GLOBAL"]
PipelineStageType = ["PRE_CI", "POST_CI"]                    
```

Note - Same data will be sent back in get api(for both - inline & ref_plugin).

###Case 2 - Ref Plugin 

When user creates a new step with plugin, they are shown all exposed variables of that plugin.
This data comes from table `plugin_step_variable`. Once pipeline is created/updated (or step is created), these variables are saved into the table `pipeline_stage_step_variable` with the same data as in inline step type.  


##Data sent to ci runner - 

###For pipeline step variables 

``` 
Input Variables:

	Name                         string
	Format                       string 
	Value                        string                       
	VariableType                 VariableType  
	ReferenceVariableStepIndex   int                                 
	ReferenceVariableName        string                             

Output Variables:

	Name                         string
	Format                       string 
	VariableType                 VariableType   
	
where,

VariableType = ["VALUE", "REF_PRE_CI", "REF_POST_CI", "REF_GLOBAL, "REF_PLUGIN"]
```

Logic used for setting VariableType - 

```
variable -> object from db
variableData -> object to be sent to ci runner
if variable.ValueType == "NEW {
	variableData.VariableType = "VALUE"
} else if variable.ValueType == "GLOBAL" {
	variableData.VariableType = "REF_GLOBAL"
} else if variable.ValueType == "FROM_PREVIOUS_STEP {
	if variable.ReferenceVariableStage == "POST_CI" {
		variableData.VariableType = "REF_POST_CI"
	} else if variable.ReferenceVariableStage == "PRE_CI" {
		variableData.VariableType = "REF_PRE_CI"
	}
}
```

Logic used for setting Value - 

```
variable -> object from db
variableData -> object to be sent to ci runner

//below checks for setting Value field is only relevant for ref_plugin
//for inline step it will always end up using user's choice(if value == "" then defaultValue will also be = "", as no defaultValue option in inline )
if variable.Value == "" {
	//no value from user; will use default value
	variableData.Value = variable.DefaultValue
} else {
	variableData.Value = variable.Value
}

Note - here we are assuming, only validated data is saved in db(that either we will have defaultValue/value in object from db)
```




###For plugin variables(send in field - refPlugins in wfRequest)

``` 
Input Variables:

	Name                         string
	Format                       string 
	Value                        string                       
	VariableType                 VariableType  
	ReferenceVariableStepIndex   int                                 
	ReferenceVariableName        string                             

Output Variables:

	Name                         string
	Format                       string 
	VariableType                 VariableType   
	
where,

VariableType = ["VALUE", "REF_PRE_CI", "REF_POST_CI", "REF_GLOBAL, "REF_PLUGIN"]
```

Logic used for setting VariableType -

```
variable -> object from db
variableData -> object to be sent to ci runner
if variable.ValueType == "NEW {
	variableData.VariableType = "VALUE"
} else if variable.ValueType == "GLOBAL" {
	variableData.VariableType = "REF_GLOBAL"
} else if variable.ValueType == "FROM_PREVIOUS_STEP {
	variableData.VariableType = "REF_PLUGIN"
}
```


Logic used for setting Value -

```
variable -> object from db
variableData -> object to be sent to ci runner

if variable.DefaultValue == "" {
	//no default value; will use value received from user, as it must be exposed
	variableData.Value = variable.Value
} else {
    if variable.IsExposed {
	    //this value will be empty as value is set in plugin_stage_step_variable
	    //& that variable is sent in pre/post steps data
	    variableData.Value = variable.Value
    } else {
	    variableData.Value = variable.DefaultValue
    }
}

Note - here we are assuming, only validated data is saved in db(that either we will have defaultValue/value in object from db)
```