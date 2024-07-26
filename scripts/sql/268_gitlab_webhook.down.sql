DELETE FROM git_host where name='Gitlab'
ALTER TABLE git_host  DROP COLUMN display_name;
