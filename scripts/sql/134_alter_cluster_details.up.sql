-- Sequence and defined type for cluster_note
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_note;

-- Sequence and defined type for cluster_note_history
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_note_history;

-- cluster_note Table Definition
CREATE TABLE IF NOT EXISTS public.cluster_note
(
    "id"                        int4         NOT NULL DEFAULT nextval('id_seq_cluster_note'::regclass),
    "cluster_id"                int4         NOT NULL,
    "description"               TEXT         NOT NULL,
    "created_on"                timestamptz  NOT NULL,
    "created_by"                int4         NOT NULL,
    "updated_on"                timestamptz,
    "updated_by"                int4,
    PRIMARY KEY ("id"),
    CONSTRAINT cluster_note_cluster_id_fkey
    FOREIGN KEY(cluster_id)
    REFERENCES public.cluster(id)
);

-- cluster_note_history Table Definition
CREATE TABLE IF NOT EXISTS public.cluster_note_history
(
    "id"                        int4         NOT NULL DEFAULT nextval('id_seq_cluster_note_history'::regclass),
    "note_id"                   int4         NOT NULL,
    "description"               TEXT         NOT NULL,
    "created_on"                timestamptz  NOT NULL,
    "created_by"                int4         NOT NULL,
    "updated_on"                timestamptz,
    "updated_by"                int4,
    PRIMARY KEY ("id"),
    CONSTRAINT cluster_note_history_cluster_note_id_fkey
    FOREIGN KEY(note_id)
    REFERENCES public.cluster_note(id)
);