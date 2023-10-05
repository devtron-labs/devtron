ALTER TABLE cluster
    DROP COLUMN to_connect_with_pinniped,
    DROP COLUMN pinniped_concierge_url;

DROP TABLE IF EXISTS public.cluster_user_data;

DROP SEQUENCE IF EXISTS public.id_seq_cluster_user_data;