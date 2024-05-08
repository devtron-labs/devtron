UPDATE pipeline_status_timeline
SET status ='KUBECTL APPLY SYNCED'
WHERE status = 'KUBECTL_APPLY_SYNCED';

UPDATE pipeline_status_timeline
SET status ='KUBECTL APPLY STARTED'
WHERE status = 'KUBECTL_APPLY_STARTED';

UPDATE pipeline_status_timeline
SET status ='GIT COMMIT'
WHERE status = 'GIT_COMMIT';