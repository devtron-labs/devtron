UPDATE pipeline_status_timeline
SET
    status = REPLACE (
            status,
            'KUBECTL_APPLY_SYNCED',
            'KUBECTL APPLY SYNCED'
        );

UPDATE pipeline_status_timeline
SET
    status = REPLACE (
            status,
            'KUBECTL_APPLY_STARTED',
            'KUBECTL APPLY STARTED'
        );

UPDATE pipeline_status_timeline
SET
    status = REPLACE (
            status,
            'GIT_COMMIT',
            'GIT COMMIT'
        );