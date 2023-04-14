DELETE FROM plugin_step_variable
WHERE plugin_step_id = (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Updater' and ps."index"=1 and ps.deleted=false);

DELETE FROM plugin_step
WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE name='Github Pull Request Updater');

DELETE FROM plugin_metadata
WHERE name = 'Github Pull Request Updater';
