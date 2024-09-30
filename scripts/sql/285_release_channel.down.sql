DELETE FROM devtron_resource_schema where devtron_resource_id in (select id from devtron_resource where kind in('release-channel'));

DELETE FROM devtron_resource where kind in('release-channel');