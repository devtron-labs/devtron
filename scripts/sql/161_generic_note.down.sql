ALTER SEQUENCE IF EXISTS id_seq_generic_note RENAME to id_seq_cluster_note;
ALTER SEQUENCE IF EXISTS id_seq_generic_note_history RENAME to id_seq_cluster_note_history;

ALTER TABLE generic_note RENAME COLUMN identifier to cluster_id;
DELETE FROM generic_note WHERE identifier_type = 1;
ALTER TABLE generic_note DROP COLUMN  identifier_type;
ALTER TABLE generic_note RENAME to cluster_note;

AlTER TABLE cluster_note ADD CONSTRAINT cluster_note_cluster_id_fkey
    FOREIGN KEY(cluster_id)
      REFERENCES public.cluster(id);

ALTER TABLE generic_note_history RENAME to cluster_note_history;
ALTER TABLE cluster_note_history DROP CONSTRAINT generic_note_history_generic_note_id_fkey
ALTER TABLE cluster_note_history ADD CONSTRAINT cluster_note_history_cluster_note_id_fkey
    FOREIGN KEY(note_id)
        REFERENCES public.cluster_note(id);