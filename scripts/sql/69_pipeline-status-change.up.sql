UPDATE pipeline_status_timeline
SET status ='KUBECTL_APPLY_SYNCED'
WHERE status = 'KUBECTL APPLY SYNCED';

UPDATE pipeline_status_timeline
SET status ='KUBECTL_APPLY_STARTED'
WHERE status = 'KUBECTL APPLY STARTED';

UPDATE pipeline_status_timeline
SET status ='GIT_COMMIT'
WHERE status = 'GIT COMMIT';