alter table kubernetes_resource_history drop column IF EXISTS version ;
alter table kubernetes_resource_history drop column IF EXISTS action_metadata;
alter table kubernetes_resource_history drop column IF EXISTS resource ;