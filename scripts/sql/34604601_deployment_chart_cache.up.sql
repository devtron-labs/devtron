--- There was an issue with Deployment chart versions 4.18.0, 4.19.0, 1.0.0, 1.1.0
--- Which was fixed in PR https://github.com/devtron-labs/devtron/pull/5215/files
--- But as we cache the reference chart in DB, we need to clear that cache for those charts
UPDATE charts
SET reference_chart = NULL
WHERE id IN (
    SELECT charts.id FROM charts
    INNER JOIN chart_ref ON (charts.chart_ref_id = chart_ref.id)
    INNER JOIN app ON (charts.app_id = app.id AND app.active = true)
    WHERE charts.active = true
      AND chart_ref.version IN ('4.18.0', '4.19.0', '1.0.0', '1.1.0')
      AND chart_ref.name = 'Deployment'
);