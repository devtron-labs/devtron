INSERT INTO git_host (name, created_on, created_by, active, webhook_url, webhook_secret, event_type_header, secret_header, secret_validator)
VALUES('Gitlab', NOW(), 1, 't', '/orchestrator/webhook/git/3/'  || MD5(random()::text), NULL, 'X-Gitlab-Event', 'X-Gitlab-Token', 'URL_APPEND')
ON CONFLICT (name) 
DO 
   UPDATE SET webhook_url='/orchestrator/webhook/git/' || (select id from git_host where name='Gitlab') || '/' || MD5(random()::text),event_type_header='X-Gitlab-Event', secret_header='X-Gitlab-Token', secret_validator='URL_APPEND';

\f ','
\a
\t 
\o '/tmp/output.csv'
select id,name from git_host;
