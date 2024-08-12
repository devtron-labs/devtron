UPDATE scan_tool_metadata
SET image_scan_descriptor_template = '[
                                        {
                                            "pathToVulnerabilitiesArray": "Results.#.Vulnerabilities",
                                            "name": "VulnerabilityID",
                                            "package": "PkgName",
                                            "packageVersion": "InstalledVersion",
                                            "fixedInVersion": "FixedVersion",
                                            "severity": "Severity"
                                        }
                                     ]', updated_on = 'now()'
WHERE name = 'TRIVY'
    AND version = 'V1'
    AND scan_target = 'IMAGE'
    AND active = true
    AND deleted = false;