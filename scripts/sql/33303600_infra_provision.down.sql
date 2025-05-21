BEGIN;
-- Drop Table: chart_category
DROP TABLE IF EXISTS "public"."chart_category_mapping";

-- Drop Sequence: id_seq_chart_category
DROP SEQUENCE IF EXISTS id_seq_chart_category_mapping;


-- Drop Table: chart_category
DROP TABLE IF EXISTS "public"."chart_category";

-- Drop Sequence: id_seq_chart_category
DROP SEQUENCE IF EXISTS id_seq_chart_category;


DROP TABLE IF EXISTS "public"."infrastructure_installation";

DROP SEQUENCE IF EXISTS id_seq_infrastructure_installation;

DROP TABLE IF EXISTS "public"."infrastructure_installation_versions";

DROP SEQUENCE IF EXISTS id_seq_infrastructure_installation_versions;

COMMIT;