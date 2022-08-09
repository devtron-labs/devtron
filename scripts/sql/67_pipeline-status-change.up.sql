UPDATE pipeline_status_timeline
SET
    status = REPLACE (
            status,
            'KUBECTL APPLY SYNCED',
            'KUBECTL_APPLY_SYNCED'
        );

UPDATE pipeline_status_timeline
SET
    status = REPLACE (
            status,
            'KUBECTL APPLY STARTED',
            'KUBECTL_APPLY_STARTED'
        );