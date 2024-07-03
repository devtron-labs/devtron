INSERT INTO public.git_host (name,active,webhook_url,webhook_secret ,event_type_header,
                             secret_header,secret_validator,created_on,created_by,updated_on,updated_by)
VALUES ('Gitlab',true, '/orchestrator/webhook/git/3', MD5(random()::text),'X-Gitlab-Event','X-Gitlab-Token','PLAIN_TEXT',now(),1,now(),1)


