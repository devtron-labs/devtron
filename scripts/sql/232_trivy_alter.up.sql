/*
 * Copyright (c) 2024. Devtron Inc.
 */

UPDATE scan_tool_metadata
SET result_descriptor_template = '[
    {
        "pathToVulnerabilitiesArray": "Results.#.Vulnerabilities",
        "name": "VulnerabilityID",
        "package": "PkgName",
        "packageVersion": "InstalledVersion",
        "fixedInVersion": "FixedVersion",
        "severity": "Severity"
    }
]' where name = 'TRIVY' and version ='V1';

