---- ALTER TABLE git_provider - modify type
ALTER TABLE git_provider
ALTER COLUMN ssh_key TYPE text;

---- ALTER TABLE git_provider - rename column
ALTER TABLE git_provider
RENAME COLUMN ssh_key TO ssh_private_key;

---- ALTER TABLE git_material - add column
ALTER TABLE git_material
ADD COLUMN fetch_submodules bool NOT NULL DEFAULT FALSE;