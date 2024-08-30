UPDATE scan_tool_metadata SET result_descriptor_template = '[
      {
        "pathToResultArray": "Results",
        "pathToVulnerabilitiesArray": "Vulnerabilities",
        "vulnerabilityData":{
       		"name": "VulnerabilityID",
        	"package": "PkgName",
        	"packageVersion": "InstalledVersion",
        	"fixedInVersion": "FixedVersion",
        	"severity": "Severity"
        },
        "resultData":{
  			"target":"Target",
        	"class":"Class",
        	"type":"Type"
       }
      }
]',updated_on = 'now()'

WHERE name = 'TRIVY'
    AND version = 'V1'
    AND scan_target = 'IMAGE';

ALTER TABLE image_scan_execution_result
    ADD COLUMN class TEXT,
    ADD COLUMN type TEXT,
    ADD COLUMN target TEXT;