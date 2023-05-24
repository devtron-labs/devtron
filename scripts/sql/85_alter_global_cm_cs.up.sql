ALTER TABLE global_cm_cs ADD type text;
--setting type as volume because until this release only volume type was supported
UPDATE global_cm_cs SET type='volume';

