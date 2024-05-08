---- DROP table
DROP TABLE IF EXISTS "cluster_note_history";
DROP TABLE IF EXISTS "cluster_note";

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_cluster_note;
DROP SEQUENCE IF EXISTS public.id_seq_cluster_note_history;