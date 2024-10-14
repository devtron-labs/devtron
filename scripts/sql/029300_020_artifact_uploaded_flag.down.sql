---
--- Drop is_artifact_uploaded flag from ci_workflow and cd_workflow_runner tables
---

ALTER TABLE public.ci_workflow
    DROP COLUMN IF EXISTS is_artifact_uploaded;
ALTER TABLE public.cd_workflow_runner
    DROP COLUMN IF EXISTS is_artifact_uploaded,
    DROP COLUMN IF EXISTS cd_artifact_location;

--
-- Name: ci_workflow_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE IF NOT EXISTS public.ci_workflow_config (
       id integer NOT NULL,
       ci_timeout bigint,
       min_cpu character varying(250),
       max_cpu character varying(250),
       min_mem character varying(250),
       max_mem character varying(250),
       min_storage character varying(250),
       max_storage character varying(250),
       min_eph_storage character varying(250),
       max_eph_storage character varying(250),
       ci_cache_bucket character varying(250),
       ci_cache_region character varying(250),
       ci_image character varying(250),
       wf_namespace character varying(250),
       logs_bucket character varying(250),
       ci_pipeline_id integer NOT NULL,
       ci_artifact_location_format character varying(256)
);


ALTER TABLE public.ci_workflow_config OWNER TO postgres;

--
-- Name: ci_workflow_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE IF NOT EXISTS public.ci_workflow_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ci_workflow_config_id_seq OWNER TO postgres;

--
-- Name: ci_workflow_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ci_workflow_config_id_seq OWNED BY public.ci_workflow_config.id;

--
-- Name: cd_workflow_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE IF NOT EXISTS public.cd_workflow_config (
       id integer NOT NULL,
       cd_timeout integer,
       min_cpu character varying(256),
       max_cpu character varying(256),
       min_mem character varying(256),
       max_mem character varying(256),
       min_storage character varying(256),
       max_storage character varying(256),
       min_eph_storage character varying(256),
       max_eph_storage character varying(256),
       cd_cache_bucket character varying(256),
       cd_cache_region character varying(256),
       cd_image character varying(256),
       wf_namespace character varying(256),
       cd_pipeline_id integer,
       logs_bucket character varying(256),
       cd_artifact_location_format character varying(256)
);


ALTER TABLE public.cd_workflow_config OWNER TO postgres;

--
-- Name: cd_workflow_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE IF NOT EXISTS public.cd_workflow_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cd_workflow_config_id_seq OWNER TO postgres;

--
-- Name: cd_workflow_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cd_workflow_config_id_seq OWNED BY public.cd_workflow_config.id;