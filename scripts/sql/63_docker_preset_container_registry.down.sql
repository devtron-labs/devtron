delete from docker_artifact_store where id = 'devtron-preset-container-registry';

delete from attributes where key IN ('PresetRegistryExpiryTime','PresetRegistryRepoName')