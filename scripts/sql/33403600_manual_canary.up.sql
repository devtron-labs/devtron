alter table kubernetes_resource_history add column IF NOT EXISTS version VARCHAR(64);
alter table kubernetes_resource_history add column IF NOT EXISTS action_metadata VARCHAR(512);
alter table kubernetes_resource_history add column IF NOT EXISTS resource VARCHAR(64);