ALTER TABLE git_host ADD COLUMN IF NOT EXISTS display_name varchar(250);
INSERT INTO git_host (name,display_name, active,webhook_url,webhook_secret ,event_type_header,
                             secret_header,secret_validator,created_on,created_by,updated_on,updated_by)
VALUES ('Gitlab_Devtron', 'Gitlab',true, '/orchestrator/webhook/git/Gitlab_Devtron', MD5(random()::text),'X-Gitlab-Event','X-Gitlab-Token','PLAIN_TEXT',now(),1,now(),1);


