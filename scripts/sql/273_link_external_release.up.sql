CREATE UNIQUE INDEX "unique_deployment_app_name"
    ON pipeline(deployment_app_name,environment_id,deleted);

ALTER TABLE pipeline
    ADD COLUMN release_mode VARCHAR(256) DEFAULT 'create';

ALTER TABLE deployment_config
    ADD COLUMN release_mode VARCHAR(256) DEFAULT 'create';
