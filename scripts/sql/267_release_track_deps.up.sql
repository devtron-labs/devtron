ALTER TABLE devtron_resource_object_dep_relations ADD COLUMN IF NOT EXISTS dependency_object_identifier VARCHAR(150);
ALTER TABLE devtron_resource_object_dep_relations ADD COLUMN IF NOT EXISTS component_object_identifier VARCHAR(150);

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
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                }
            ]
        },
        {
            "operationType": "patch",
            "operationPaths":
            [
                "dependency.applications"
            ],
            "stateTo":
            {
                "configStatus": "draft",
                "dependencyArtifactStatus": "*",
                "rolloutStatus": "notDeployed",
                "lockStatus": "unLocked",
                "isReleaseTrackAppPropagationOffending": "yes"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked",
                    "isReleaseTrackAppPropagationOffending": "*"
                }
            ]
        },
        {
            "operationType": "patch",
            "operationPaths":
            [
                "dependency.applications.image"
            ],
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked",
                    "isReleaseTrackAppPropagationOffending": "*"
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
                    "lockStatus": "unLocked",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked",
                    "isReleaseTrackAppPropagationOffending": "*"
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
                    "lockStatus": "locked",
                    "isReleaseTrackAppPropagationOffending": "yes"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "partiallyDeployed",
                    "lockStatus": "locked",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "completelyDeployed",
                    "lockStatus": "locked",
                    "isReleaseTrackAppPropagationOffending": "*"
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
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "hold",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "rescind",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                },
                {
                    "configStatus": "corrupted",
                    "dependencyArtifactStatus": "*",
                    "rolloutStatus": "*",
                    "lockStatus": "*",
                    "isReleaseTrackAppPropagationOffending": "*"
                }
            ]
        }
    ],
    "consequence": "BLOCK"
}' where policy_of='RELEASE_ACTION_CHECK';



UPDATE global_policy set policy_json='{
    "definitions":
    [
        {
            "stateTo":
            {
                "configStatus": "draft",
                "dependencyArtifactStatus": "noImageSelected",
                "rolloutStatus": "notDeployed",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "noImageSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "partialImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "partialImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "draft",
                "dependencyArtifactStatus": "noImageSelected",
                "rolloutStatus": "notDeployed",
                "lockStatus": "locked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "noImageSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "draft",
                "dependencyArtifactStatus": "partialImagesSelected",
                "rolloutStatus": "notDeployed",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "partialImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "draft",
                "dependencyArtifactStatus": "partialImagesSelected",
                "rolloutStatus": "notDeployed",
                "lockStatus": "locked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "partialImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "draft",
                "dependencyArtifactStatus": "allImagesSelected",
                "rolloutStatus": "notDeployed",
                "lockStatus": "locked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "notDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "draft",
                "dependencyArtifactStatus": "allImagesSelected",
                "rolloutStatus": "notDeployed",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "partialImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "noImageSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "partialImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "readyForRelease",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "notDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "*"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "draft",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "rolloutStatus": "notDeployed",
                    "lockStatus": "locked"
                }
            ],
            "autoAction":
            {
                "configStatus": "readyForRelease",
                "lockStatus": "locked"
            }
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "notDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "notDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ],
            "autoAction":
            {
                "configStatus": "draft",
                "lockStatus": "unLocked"
            }
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "partiallyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "locked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "notDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "partiallyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ],
            "autoAction":
            {
                "configStatus": "hold",
                "lockStatus": "unLocked"
            }
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "partiallyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "*"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "hold",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "hold",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ],
            "autoAction":
            {
                "configStatus": "readyForRelease",
                "lockStatus": "locked"
            }
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "completelyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "locked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "completelyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ],
            "autoAction":
            {
                "configStatus": "hold",
                "lockStatus": "unLocked"
            }
        },
        {
            "stateTo":
            {
                "configStatus": "readyForRelease",
                "rolloutStatus": "completelyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "*"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "hold",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "unLocked"
                },
                {
                    "configStatus": "hold",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ],
            "autoAction":
            {
                "configStatus": "readyForRelease",
                "lockStatus": "locked"
            }
        },
        {
            "stateTo":
            {
                "configStatus": "hold",
                "rolloutStatus": "partiallyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "locked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "hold",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "unLocked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "hold",
                "rolloutStatus": "partiallyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "hold",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "hold",
                "rolloutStatus": "completelyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "locked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "hold",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "unLocked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "hold",
                "rolloutStatus": "completelyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "unLocked"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "hold",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "rescind",
                "rolloutStatus": "completelyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "*"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "hold",
                    "rolloutStatus": "completelyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ]
        },
        {
            "stateTo":
            {
                "configStatus": "rescind",
                "rolloutStatus": "partiallyDeployed",
                "dependencyArtifactStatus": "allImagesSelected",
                "lockStatus": "*"
            },
            "possibleFromStates":
            [
                {
                    "configStatus": "readyForRelease",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                },
                {
                    "configStatus": "hold",
                    "rolloutStatus": "partiallyDeployed",
                    "dependencyArtifactStatus": "allImagesSelected",
                    "lockStatus": "locked"
                }
            ]
        }
    ],
    "consequence": "BLOCK"
}' where policy_of='RELEASE_STATUS';
