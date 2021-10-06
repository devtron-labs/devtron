---- ALTER TABLE git_provider - modify type
ALTER TABLE git_provider
ALTER COLUMN ssh_private_key TYPE varchar(250);

---- ALTER TABLE git_provider - rename column
ALTER TABLE git_provider
RENAME COLUMN ssh_private_key TO ssh_key;

---- ALTER TABLE git_material - drop column
ALTER TABLE git_material
DROP COLUMN IF EXISTS fetch_submodules
