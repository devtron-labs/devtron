---- DROP TABLE
DROP TABLE IF EXISTS public.release_tags;
DROP TABLE IF EXISTS public.image_comments;
DROP TABLE IF EXISTS public.image_tagging_audit;

---- DROP types
DROP TYPE IF EXISTS data_type;
DROP TYPE IF EXISTS action_type;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_image_tagging_audit;
DROP SEQUENCE IF EXISTS public.id_seq_image_comment;
DROP SEQUENCE IF EXISTS public.id_seq_image_tag;



