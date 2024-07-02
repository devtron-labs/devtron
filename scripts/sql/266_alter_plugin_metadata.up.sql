ALTER TABLE plugin_metadata ADD COLUMN plugin_parent_metadata_id integer;
ALTER TABLE plugin_metadata ADD COLUMN plugin_version text NOT NULL DEFAULT '1.0.0';
ALTER TABLE plugin_metadata ADD COLUMN is_deprecated bool NOT NULL default false;
ALTER TABLE plugin_metadata ADD COLUMN is_latest bool NOT NULL default true;
ALTER TABLE plugin_metadata ADD COLUMN doc_link text;

ALTER TABLE  public.plugin_metadata
    ADD CONSTRAINT plugin_metadata_plugin_parent_metadata_id_fkey
        FOREIGN KEY ("plugin_parent_metadata_id")
            REFERENCES "public"."plugin_parent_metadata" ("id");
