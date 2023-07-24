ALTER SEQUENCE IF EXISTS id_seq_cluster_note RENAME to id_seq_generic_note;
ALTER SEQUENCE IF EXISTS id_seq_cluster_note_history RENAME to id_seq_generic_note_history;
AlTER TABLE cluster_note DROP CONSTRAINT cluster_note_cluster_id_fkey;
ALTER TABLE cluster_note RENAME to generic_note;
ALTER TABLE generic_note RENAME COLUMN cluster_id to identifier;
ALTER TABLE generic_note ADD COLUMN  identifier_type int4 DEFAULT 0 NOT NULL;

AlTER TABLE cluster_note_history DROP CONSTRAINT cluster_note_history_cluster_note_id_fkey;
ALTER TABLE cluster_note_history RENAME to generic_note_history;
ALTER TABLE generic_note_history ADD CONSTRAINT generic_note_history_generic_note_id_fkey
    FOREIGN KEY(note_id)
    REFERENCES public.generic_note(id);