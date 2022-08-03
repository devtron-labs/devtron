Insert into docker_artifact_store(id, plugin_id, registry_url, registry_type, username, password, active, is_default, created_on, created_by, updated_on, updated_by, connection) values('devtron-preset-container-registry', 'cd.go.artifact.docker.registry', 'ttl.sh', 'other','a','a', 't', 'f', now(), 1, now(), 1, 'secure')

Insert into attributes(key,value,active,created_on,created_by,updated_on) values('PresetRegistryRepoName','devtron-preset-registry-repo',true,now(),1,now());

Insert into attributes(key,value,active,created_on,created_by,updated_on) values('PresetRegistryExpiryTime','86400',true,now(),1,now());

