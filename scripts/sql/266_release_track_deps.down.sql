ALTER TABLE devtron_resource_object_dep_relations DROP COLUMN dependency_object_identifier;
ALTER TABLE devtron_resource_object_dep_relations DROP COLUMN component_object_identifier;

UPDATE global_policy set policy_json='{
    "definitions":
    [
        {
            "operationType": "patch",
            "operationPaths":
            [
                "overview.description",
                "overview.releaseNote",
                "overview.tags",
                "overview.name",
                "overview.metadata",
                "status.config.lock",
                "status.config"
            ],
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*"
                }
            ]
        },
        {
            "operationType": "patch",
            "operationPaths":
            [
                "dependency.applications",
                "dependency.applications.image"
            ],
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                }
            ]
        },
        {
            "operationType": "patch",
            "operationPaths":
            [
                "dependency.applications.instruction"
            ],
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                }
            ]
        },
        {
            "operationType": "deploymentTrigger",
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "operationType": "delete",
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "*"
                },
                {
                    "configStatus": "rescind",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*"
                },
                {
                    "configStatus": "corrupted",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*"
                }
            ]
        }
    ],
    "consequence": "BLOCK"
}' where policy_of='RELEASE_ACTION_CHECK';
