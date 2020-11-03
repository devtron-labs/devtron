--
-- PostgreSQL database dump
--

-- Dumped from database version 11.3
-- Dumped by pg_dump version 11.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: app; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app (
    id integer NOT NULL,
    app_name character varying(250) NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    team_id integer,
    app_store boolean DEFAULT false
);


ALTER TABLE public.app OWNER TO postgres;

--
-- Name: app_env_linkouts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_env_linkouts (
    id integer NOT NULL,
    app_id integer,
    environment_id integer,
    link text,
    description text,
    name character varying(100) NOT NULL,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer
);


ALTER TABLE public.app_env_linkouts OWNER TO postgres;

--
-- Name: app_env_linkouts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_env_linkouts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_env_linkouts_id_seq OWNER TO postgres;

--
-- Name: app_env_linkouts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.app_env_linkouts_id_seq OWNED BY public.app_env_linkouts.id;


--
-- Name: app_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_id_seq OWNER TO postgres;

--
-- Name: app_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.app_id_seq OWNED BY public.app.id;


--
-- Name: app_level_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_level_metrics (
    id integer NOT NULL,
    app_id integer NOT NULL,
    app_metrics boolean NOT NULL,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer,
    infra_metrics boolean DEFAULT true
);


ALTER TABLE public.app_level_metrics OWNER TO postgres;

--
-- Name: app_level_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_level_metrics_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_level_metrics_id_seq OWNER TO postgres;

--
-- Name: app_level_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.app_level_metrics_id_seq OWNED BY public.app_level_metrics.id;


--
-- Name: app_store; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_store (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    chart_repo_id integer,
    active boolean NOT NULL,
    chart_git_location character varying(250),
    created_on timestamp with time zone NOT NULL,
    updated_on timestamp with time zone NOT NULL
);


ALTER TABLE public.app_store OWNER TO postgres;

--
-- Name: app_store_application_version; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_store_application_version (
    id integer NOT NULL,
    version character varying(250),
    app_version character varying(250),
    created timestamp with time zone,
    deprecated boolean,
    description text,
    digest character varying(250),
    icon character varying(250),
    name character varying(100),
    home character varying(100),
    source character varying(250),
    values_yaml json NOT NULL,
    chart_yaml json NOT NULL,
    app_store_id integer,
    latest boolean DEFAULT false,
    created_on timestamp with time zone NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    raw_values text,
    readme text,
    created_by integer,
    updated_by integer
);


ALTER TABLE public.app_store_application_version OWNER TO postgres;

--
-- Name: app_store_application_version_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_store_application_version_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_store_application_version_id_seq OWNER TO postgres;

--
-- Name: app_store_application_version_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.app_store_application_version_id_seq OWNED BY public.app_store_application_version.id;


--
-- Name: app_store_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_store_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_store_id_seq OWNER TO postgres;

--
-- Name: app_store_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.app_store_id_seq OWNED BY public.app_store.id;


--
-- Name: app_store_version_values; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_store_version_values (
    id integer NOT NULL,
    name character varying(100),
    values_yaml text NOT NULL,
    app_store_application_version_id integer,
    deleted boolean DEFAULT false NOT NULL,
    created_by integer,
    updated_by integer,
    created_on timestamp with time zone NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    reference_type character varying(50)
);


ALTER TABLE public.app_store_version_values OWNER TO postgres;

--
-- Name: app_store_version_values_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_store_version_values_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_store_version_values_id_seq OWNER TO postgres;

--
-- Name: app_store_version_values_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.app_store_version_values_id_seq OWNED BY public.app_store_version_values.id;


--
-- Name: app_workflow; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_workflow (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    app_id integer NOT NULL,
    workflow_dag text,
    active boolean,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer
);


ALTER TABLE public.app_workflow OWNER TO postgres;

--
-- Name: app_workflow_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_workflow_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_workflow_id_seq OWNER TO postgres;

--
-- Name: app_workflow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.app_workflow_id_seq OWNED BY public.app_workflow.id;


--
-- Name: app_workflow_mapping_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.app_workflow_mapping_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.app_workflow_mapping_id_seq OWNER TO postgres;

--
-- Name: app_workflow_mapping; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_workflow_mapping (
    id integer DEFAULT nextval('public.app_workflow_mapping_id_seq'::regclass) NOT NULL,
    type character varying(100),
    component_id integer,
    parent_id integer,
    app_workflow_id integer NOT NULL,
    active boolean,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer,
    parent_type character varying(100)
);


ALTER TABLE public.app_workflow_mapping OWNER TO postgres;

--
-- Name: casbin_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.casbin_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.casbin_id_seq OWNER TO postgres;

--
-- Name: casbin_role_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.casbin_role_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.casbin_role_id_seq OWNER TO postgres;

--
-- Name: cd_workflow; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cd_workflow (
    id integer NOT NULL,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer,
    ci_artifact_id integer NOT NULL,
    pipeline_id integer NOT NULL,
    workflow_status character varying(256)
);


ALTER TABLE public.cd_workflow OWNER TO postgres;

--
-- Name: cd_workflow_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cd_workflow_config (
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

CREATE SEQUENCE public.cd_workflow_config_id_seq
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


--
-- Name: cd_workflow_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cd_workflow_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cd_workflow_id_seq OWNER TO postgres;

--
-- Name: cd_workflow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cd_workflow_id_seq OWNED BY public.cd_workflow.id;


--
-- Name: cd_workflow_runner; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cd_workflow_runner (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    workflow_type character varying(256) NOT NULL,
    executor_type character varying(256) NOT NULL,
    status character varying(256),
    pod_status character varying(256),
    message character varying(256),
    started_on timestamp with time zone,
    finished_on timestamp with time zone,
    namespace character varying(256),
    log_file_path character varying(256),
    triggered_by integer,
    cd_workflow_id integer NOT NULL
);


ALTER TABLE public.cd_workflow_runner OWNER TO postgres;

--
-- Name: cd_workflow_runner_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cd_workflow_runner_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cd_workflow_runner_id_seq OWNER TO postgres;

--
-- Name: cd_workflow_runner_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cd_workflow_runner_id_seq OWNED BY public.cd_workflow_runner.id;


--
-- Name: chart_env_config_override; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chart_env_config_override (
    id integer NOT NULL,
    chart_id integer,
    target_environment integer,
    env_override_yaml text NOT NULL,
    status character varying(50) NOT NULL,
    reviewed boolean NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    namespace character varying(250),
    latest boolean DEFAULT false NOT NULL,
    previous boolean DEFAULT false NOT NULL,
    is_override boolean
);


ALTER TABLE public.chart_env_config_override OWNER TO postgres;

--
-- Name: chart_env_config_override_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chart_env_config_override_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.chart_env_config_override_id_seq OWNER TO postgres;

--
-- Name: chart_env_config_override_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.chart_env_config_override_id_seq OWNED BY public.chart_env_config_override.id;


--
-- Name: chart_group; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chart_group (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    description text,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.chart_group OWNER TO postgres;

--
-- Name: chart_group_deployment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chart_group_deployment (
    id integer NOT NULL,
    chart_group_id integer NOT NULL,
    chart_group_entry_id integer,
    installed_app_id integer NOT NULL,
    group_installation_id character varying(250),
    deleted boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.chart_group_deployment OWNER TO postgres;

--
-- Name: chart_group_deployment_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chart_group_deployment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.chart_group_deployment_id_seq OWNER TO postgres;

--
-- Name: chart_group_deployment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.chart_group_deployment_id_seq OWNED BY public.chart_group_deployment.id;


--
-- Name: chart_group_entry; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chart_group_entry (
    id integer NOT NULL,
    app_store_values_version_id integer,
    app_store_application_version_id integer,
    chart_group_id integer,
    deleted boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.chart_group_entry OWNER TO postgres;

--
-- Name: chart_group_entry_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chart_group_entry_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.chart_group_entry_id_seq OWNER TO postgres;

--
-- Name: chart_group_entry_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.chart_group_entry_id_seq OWNED BY public.chart_group_entry.id;


--
-- Name: chart_group_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chart_group_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.chart_group_id_seq OWNER TO postgres;

--
-- Name: chart_group_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.chart_group_id_seq OWNED BY public.chart_group.id;


--
-- Name: id_seq_chart_ref; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.id_seq_chart_ref
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.id_seq_chart_ref OWNER TO postgres;

--
-- Name: chart_ref; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chart_ref (
    id integer DEFAULT nextval('public.id_seq_chart_ref'::regclass) NOT NULL,
    location character varying(250),
    version character varying(250),
    is_default boolean,
    active boolean,
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
    updated_by integer
);


ALTER TABLE public.chart_ref OWNER TO postgres;

--
-- Name: chart_repo; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.chart_repo (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    url character varying(250) NOT NULL,
    is_default boolean NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    external boolean DEFAULT false
);


ALTER TABLE public.chart_repo OWNER TO postgres;

--
-- Name: chart_repo_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.chart_repo_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.chart_repo_id_seq OWNER TO postgres;

--
-- Name: chart_repo_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.chart_repo_id_seq OWNED BY public.chart_repo.id;


--
-- Name: charts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.charts (
    id integer NOT NULL,
    app_id integer,
    chart_repo_id integer,
    chart_name character varying(250) NOT NULL,
    chart_version character varying(250) NOT NULL,
    chart_repo character varying(250) NOT NULL,
    chart_repo_url character varying(250) NOT NULL,
    git_repo_url character varying(250) NOT NULL,
    chart_location character varying(250) NOT NULL,
    status character varying(50) NOT NULL,
    active boolean NOT NULL,
    reference_template character varying(250) NOT NULL,
    values_yaml text NOT NULL,
    global_override text NOT NULL,
    environment_override text,
    release_override text NOT NULL,
    user_overrides text,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    image_descriptor_template text,
    latest boolean DEFAULT false NOT NULL,
    chart_ref_id integer NOT NULL,
    pipeline_override text DEFAULT '{}'::text NOT NULL,
    previous boolean DEFAULT false NOT NULL
);


ALTER TABLE public.charts OWNER TO postgres;

--
-- Name: charts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.charts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.charts_id_seq OWNER TO postgres;

--
-- Name: charts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.charts_id_seq OWNED BY public.charts.id;


--
-- Name: ci_artifact; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ci_artifact (
    id integer NOT NULL,
    image character varying(250),
    image_digest character varying(250),
    material_info text,
    data_source character varying(50),
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    pipeline_id integer,
    ci_workflow_id integer,
    parent_ci_artifact integer,
    scan_enabled boolean DEFAULT false NOT NULL,
    scanned boolean DEFAULT false NOT NULL
);


ALTER TABLE public.ci_artifact OWNER TO postgres;

--
-- Name: ci_artifact_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ci_artifact_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ci_artifact_id_seq OWNER TO postgres;

--
-- Name: ci_artifact_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ci_artifact_id_seq OWNED BY public.ci_artifact.id;


--
-- Name: ci_pipeline; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ci_pipeline (
    id integer NOT NULL,
    app_id integer,
    ci_template_id integer,
    name character varying(250),
    version character varying(250),
    active boolean NOT NULL,
    deleted boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    manual boolean DEFAULT false NOT NULL,
    external boolean DEFAULT false,
    docker_args text,
    parent_ci_pipeline integer,
    scan_enabled boolean DEFAULT false NOT NULL
);


ALTER TABLE public.ci_pipeline OWNER TO postgres;

--
-- Name: ci_pipeline_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ci_pipeline_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ci_pipeline_id_seq OWNER TO postgres;

--
-- Name: ci_pipeline_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ci_pipeline_id_seq OWNED BY public.ci_pipeline.id;


--
-- Name: ci_pipeline_material; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ci_pipeline_material (
    id integer NOT NULL,
    git_material_id integer,
    ci_pipeline_id integer,
    path character varying(250),
    checkout_path character varying(250),
    type character varying(250),
    value character varying(250),
    scm_id character varying(250),
    scm_name character varying(250),
    scm_version character varying(250),
    active boolean,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.ci_pipeline_material OWNER TO postgres;

--
-- Name: ci_pipeline_material_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ci_pipeline_material_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ci_pipeline_material_id_seq OWNER TO postgres;

--
-- Name: ci_pipeline_material_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ci_pipeline_material_id_seq OWNED BY public.ci_pipeline_material.id;


--
-- Name: ci_pipeline_scripts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ci_pipeline_scripts (
    id integer NOT NULL,
    name character varying(256) NOT NULL,
    index integer NOT NULL,
    ci_pipeline_id integer NOT NULL,
    script text,
    stage character varying(256),
    output_location character varying(256),
    active boolean,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer
);


ALTER TABLE public.ci_pipeline_scripts OWNER TO postgres;

--
-- Name: ci_pipeline_scripts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ci_pipeline_scripts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ci_pipeline_scripts_id_seq OWNER TO postgres;

--
-- Name: ci_pipeline_scripts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ci_pipeline_scripts_id_seq OWNED BY public.ci_pipeline_scripts.id;


--
-- Name: ci_template; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ci_template (
    id integer NOT NULL,
    app_id integer,
    docker_registry_id character varying(250),
    docker_repository character varying(250),
    dockerfile_path character varying(250),
    args text,
    before_docker_build text,
    after_docker_build text,
    template_name character varying(250),
    version character varying(250),
    active boolean,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    git_material_id integer
);


ALTER TABLE public.ci_template OWNER TO postgres;

--
-- Name: ci_template_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ci_template_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ci_template_id_seq OWNER TO postgres;

--
-- Name: ci_template_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ci_template_id_seq OWNED BY public.ci_template.id;


--
-- Name: ci_workflow; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ci_workflow (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    status character varying(50),
    pod_status character varying(50),
    message character varying(250),
    started_on timestamp with time zone,
    finished_on timestamp with time zone,
    namespace character varying(250),
    log_file_path character varying(250),
    git_triggers json,
    triggered_by integer NOT NULL,
    ci_pipeline_id integer NOT NULL,
    ci_artifact_location character varying(256)
);


ALTER TABLE public.ci_workflow OWNER TO postgres;

--
-- Name: ci_workflow_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ci_workflow_config (
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

CREATE SEQUENCE public.ci_workflow_config_id_seq
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
-- Name: ci_workflow_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ci_workflow_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ci_workflow_id_seq OWNER TO postgres;

--
-- Name: ci_workflow_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ci_workflow_id_seq OWNED BY public.ci_workflow.id;


--
-- Name: cluster; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cluster (
    id integer NOT NULL,
    cluster_name character varying(250) NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    server_url character varying(250),
    config json,
    prometheus_endpoint character varying(250),
    cd_argo_setup boolean DEFAULT false,
    p_username character varying(250),
    p_password character varying(250),
    p_tls_client_cert text,
    p_tls_client_key text
);


ALTER TABLE public.cluster OWNER TO postgres;

--
-- Name: cluster_accounts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cluster_accounts (
    id integer NOT NULL,
    account character varying(250) NOT NULL,
    config json NOT NULL,
    cluster_id integer NOT NULL,
    namespace character varying(250) NOT NULL,
    is_default boolean DEFAULT false,
    active boolean DEFAULT true NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.cluster_accounts OWNER TO postgres;

--
-- Name: cluster_accounts_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cluster_accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cluster_accounts_id_seq OWNER TO postgres;

--
-- Name: cluster_accounts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cluster_accounts_id_seq OWNED BY public.cluster_accounts.id;


--
-- Name: cluster_helm_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cluster_helm_config (
    id integer NOT NULL,
    cluster_id integer NOT NULL,
    tiller_url character varying(250),
    tiller_cert character varying,
    tiller_key character varying,
    active boolean DEFAULT true NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.cluster_helm_config OWNER TO postgres;

--
-- Name: cluster_helm_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cluster_helm_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cluster_helm_config_id_seq OWNER TO postgres;

--
-- Name: cluster_helm_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cluster_helm_config_id_seq OWNED BY public.cluster_helm_config.id;


--
-- Name: cluster_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cluster_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cluster_id_seq OWNER TO postgres;

--
-- Name: cluster_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cluster_id_seq OWNED BY public.cluster.id;


--
-- Name: cluster_installed_apps_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cluster_installed_apps_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cluster_installed_apps_id_seq OWNER TO postgres;

--
-- Name: cluster_installed_apps; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cluster_installed_apps (
    id integer DEFAULT nextval('public.cluster_installed_apps_id_seq'::regclass) NOT NULL,
    cluster_id integer,
    installed_app_id integer,
    created_by integer,
    created_on timestamp with time zone,
    updated_by integer,
    updated_on timestamp with time zone
);


ALTER TABLE public.cluster_installed_apps OWNER TO postgres;

--
-- Name: id_seq_config_map_app_level; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.id_seq_config_map_app_level
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.id_seq_config_map_app_level OWNER TO postgres;

--
-- Name: config_map_app_level; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.config_map_app_level (
    id integer DEFAULT nextval('public.id_seq_config_map_app_level'::regclass),
    app_id integer NOT NULL,
    config_map_data text,
    secret_data text,
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
    updated_by integer
);


ALTER TABLE public.config_map_app_level OWNER TO postgres;

--
-- Name: id_seq_config_map_env_level; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.id_seq_config_map_env_level
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.id_seq_config_map_env_level OWNER TO postgres;

--
-- Name: config_map_env_level; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.config_map_env_level (
    id integer DEFAULT nextval('public.id_seq_config_map_env_level'::regclass),
    app_id integer NOT NULL,
    environment_id integer NOT NULL,
    config_map_data text,
    secret_data text,
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
    updated_by integer
);


ALTER TABLE public.config_map_env_level OWNER TO postgres;

--
-- Name: id_seq_config_map_pipeline_level; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.id_seq_config_map_pipeline_level
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.id_seq_config_map_pipeline_level OWNER TO postgres;

--
-- Name: config_map_pipeline_level; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.config_map_pipeline_level (
    id integer DEFAULT nextval('public.id_seq_config_map_pipeline_level'::regclass),
    app_id integer NOT NULL,
    environment_id integer NOT NULL,
    pipeline_id integer NOT NULL,
    config_map_data text,
    secret_data text,
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
    updated_by integer
);


ALTER TABLE public.config_map_pipeline_level OWNER TO postgres;

--
-- Name: cve_policy_control_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cve_policy_control_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.cve_policy_control_id_seq OWNER TO postgres;

--
-- Name: cve_policy_control; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cve_policy_control (
    id integer DEFAULT nextval('public.cve_policy_control_id_seq'::regclass) NOT NULL,
    global boolean,
    cluster_id integer,
    env_id integer,
    app_id integer,
    cve_store_id character varying(255),
    action integer,
    severity integer,
    deleted boolean,
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
    updated_by integer
);


ALTER TABLE public.cve_policy_control OWNER TO postgres;

--
-- Name: cve_store; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cve_store (
    name character varying(255) NOT NULL,
    severity integer,
    package character varying(255),
    version character varying(255),
    fixed_version character varying(255),
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
    updated_by integer
);


ALTER TABLE public.cve_store OWNER TO postgres;

--
-- Name: db_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.db_config (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    type character varying(250) NOT NULL,
    host character varying(250) NOT NULL,
    port character varying(250) NOT NULL,
    db_name character varying(250) NOT NULL,
    user_name character varying(250) NOT NULL,
    password character varying(250) NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.db_config OWNER TO postgres;

--
-- Name: db_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.db_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.db_config_id_seq OWNER TO postgres;

--
-- Name: db_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.db_config_id_seq OWNED BY public.db_config.id;


--
-- Name: db_migration_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.db_migration_config (
    id integer NOT NULL,
    db_config_id integer NOT NULL,
    pipeline_id integer NOT NULL,
    git_material_id integer NOT NULL,
    script_source character varying(250) NOT NULL,
    migration_tool character varying(250) NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.db_migration_config OWNER TO postgres;

--
-- Name: db_migration_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.db_migration_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.db_migration_config_id_seq OWNER TO postgres;

--
-- Name: db_migration_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.db_migration_config_id_seq OWNED BY public.db_migration_config.id;


--
-- Name: deployment_group; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.deployment_group (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    status character varying(50),
    app_count integer,
    no_of_apps text,
    environment_id integer,
    ci_pipeline_id integer,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.deployment_group OWNER TO postgres;

--
-- Name: deployment_group_app; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.deployment_group_app (
    id integer NOT NULL,
    deployment_group_id integer,
    app_id integer,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.deployment_group_app OWNER TO postgres;

--
-- Name: deployment_group_app_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.deployment_group_app_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.deployment_group_app_id_seq OWNER TO postgres;

--
-- Name: deployment_group_app_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.deployment_group_app_id_seq OWNED BY public.deployment_group_app.id;


--
-- Name: deployment_group_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.deployment_group_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.deployment_group_id_seq OWNER TO postgres;

--
-- Name: deployment_group_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.deployment_group_id_seq OWNED BY public.deployment_group.id;


--
-- Name: deployment_status; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.deployment_status (
    id integer NOT NULL,
    app_name character varying(250) NOT NULL,
    status character varying(50) NOT NULL,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    app_id integer,
    env_id integer
);


ALTER TABLE public.deployment_status OWNER TO postgres;

--
-- Name: deployment_status_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.deployment_status_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.deployment_status_id_seq OWNER TO postgres;

--
-- Name: deployment_status_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.deployment_status_id_seq OWNED BY public.deployment_status.id;


--
-- Name: docker_artifact_store; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.docker_artifact_store (
    id character varying(250) NOT NULL,
    plugin_id character varying(250) NOT NULL,
    registry_url character varying(250) NOT NULL,
    registry_type character varying(250) NOT NULL,
    aws_accesskey_id character varying(250),
    aws_secret_accesskey character varying(250),
    aws_region character varying(250),
    username character varying(250),
    password character varying(250),
    is_default boolean NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.docker_artifact_store OWNER TO postgres;

--
-- Name: env_level_app_metrics; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.env_level_app_metrics (
    id integer NOT NULL,
    app_id integer NOT NULL,
    env_id integer NOT NULL,
    app_metrics boolean,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer,
    infra_metrics boolean DEFAULT true
);


ALTER TABLE public.env_level_app_metrics OWNER TO postgres;

--
-- Name: env_level_app_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.env_level_app_metrics_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.env_level_app_metrics_id_seq OWNER TO postgres;

--
-- Name: env_level_app_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.env_level_app_metrics_id_seq OWNED BY public.env_level_app_metrics.id;


--
-- Name: environment; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.environment (
    id integer NOT NULL,
    environment_name character varying(250) NOT NULL,
    cluster_id integer NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    "default" boolean DEFAULT false NOT NULL,
    namespace character varying(250),
    grafana_datasource_id integer
);


ALTER TABLE public.environment OWNER TO postgres;

--
-- Name: environment_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.environment_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.environment_id_seq OWNER TO postgres;

--
-- Name: environment_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.environment_id_seq OWNED BY public.environment.id;


--
-- Name: event; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.event (
    id integer NOT NULL,
    event_type character varying(100) NOT NULL,
    description character varying(250)
);


ALTER TABLE public.event OWNER TO postgres;

--
-- Name: event_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.event_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.event_id_seq OWNER TO postgres;

--
-- Name: event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.event_id_seq OWNED BY public.event.id;


--
-- Name: events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.events (
    id integer NOT NULL,
    namespace character varying(250),
    kind character varying(250),
    component character varying(250),
    host character varying(250),
    reason character varying(250),
    status character varying(250),
    name character varying(250),
    message character varying(250),
    resource_revision character varying(250),
    creation_time_stamp timestamp with time zone,
    uid character varying(250),
    pipeline_name character varying(250),
    release_version character varying(250),
    created_on timestamp with time zone NOT NULL,
    created_by character varying(250) NOT NULL
);


ALTER TABLE public.events OWNER TO postgres;

--
-- Name: events_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.events_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.events_id_seq OWNER TO postgres;

--
-- Name: events_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.events_id_seq OWNED BY public.events.id;


--
-- Name: external_ci_pipeline; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.external_ci_pipeline (
    id integer NOT NULL,
    ci_pipeline_id integer NOT NULL,
    access_token character varying(256) NOT NULL,
    active boolean,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer
);


ALTER TABLE public.external_ci_pipeline OWNER TO postgres;

--
-- Name: external_ci_pipeline_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.external_ci_pipeline_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.external_ci_pipeline_id_seq OWNER TO postgres;

--
-- Name: external_ci_pipeline_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.external_ci_pipeline_id_seq OWNED BY public.external_ci_pipeline.id;


--
-- Name: git_material; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.git_material (
    id integer NOT NULL,
    app_id integer,
    git_provider_id integer,
    active boolean NOT NULL,
    name character varying(250),
    url character varying(250),
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    checkout_path character varying(250)
);


ALTER TABLE public.git_material OWNER TO postgres;

--
-- Name: git_material_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.git_material_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.git_material_id_seq OWNER TO postgres;

--
-- Name: git_material_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.git_material_id_seq OWNED BY public.git_material.id;


--
-- Name: git_provider; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.git_provider (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    url character varying(250) NOT NULL,
    user_name character varying(25),
    password character varying(250),
    ssh_key character varying(250),
    access_token character varying(250),
    auth_mode character varying(250),
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.git_provider OWNER TO postgres;

--
-- Name: git_provider_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.git_provider_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.git_provider_id_seq OWNER TO postgres;

--
-- Name: git_provider_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.git_provider_id_seq OWNED BY public.git_provider.id;


--
-- Name: git_web_hook; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.git_web_hook (
    id integer NOT NULL,
    ci_material_id integer NOT NULL,
    git_material_id integer NOT NULL,
    type character varying(250),
    value character varying(250),
    active boolean,
    last_seen_hash character varying(250),
    created_on timestamp with time zone
);


ALTER TABLE public.git_web_hook OWNER TO postgres;

--
-- Name: git_web_hook_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.git_web_hook_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.git_web_hook_id_seq OWNER TO postgres;

--
-- Name: git_web_hook_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.git_web_hook_id_seq OWNED BY public.git_web_hook.id;


--
-- Name: helm_values; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.helm_values (
    app_name character varying(250) NOT NULL,
    environment character varying(250) NOT NULL,
    values_yaml text NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.helm_values OWNER TO postgres;

--
-- Name: id_seq_pconfig; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.id_seq_pconfig
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.id_seq_pconfig OWNER TO postgres;

--
-- Name: image_scan_deploy_info_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.image_scan_deploy_info_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_scan_deploy_info_id_seq OWNER TO postgres;

--
-- Name: image_scan_deploy_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.image_scan_deploy_info (
    id integer DEFAULT nextval('public.image_scan_deploy_info_id_seq'::regclass) NOT NULL,
    image_scan_execution_history_id integer[],
    scan_object_meta_id integer,
    object_type character varying(255),
    cluster_id integer,
    env_id integer,
    created_on timestamp without time zone,
    created_by integer,
    updated_on timestamp without time zone,
    updated_by integer
);


ALTER TABLE public.image_scan_deploy_info OWNER TO postgres;

--
-- Name: image_scan_execution_history_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.image_scan_execution_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_scan_execution_history_id_seq OWNER TO postgres;

--
-- Name: image_scan_execution_history; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.image_scan_execution_history (
    id integer DEFAULT nextval('public.image_scan_execution_history_id_seq'::regclass) NOT NULL,
    image character varying(255),
    execution_time timestamp with time zone,
    executed_by integer,
    image_hash character varying(255)
);


ALTER TABLE public.image_scan_execution_history OWNER TO postgres;

--
-- Name: image_scan_execution_result_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.image_scan_execution_result_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_scan_execution_result_id_seq OWNER TO postgres;

--
-- Name: image_scan_execution_result; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.image_scan_execution_result (
    id integer DEFAULT nextval('public.image_scan_execution_result_id_seq'::regclass) NOT NULL,
    image_scan_execution_history_id integer NOT NULL,
    cve_store_name character varying(255) NOT NULL
);


ALTER TABLE public.image_scan_execution_result OWNER TO postgres;

--
-- Name: image_scan_object_meta_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.image_scan_object_meta_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.image_scan_object_meta_id_seq OWNER TO postgres;

--
-- Name: image_scan_object_meta; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.image_scan_object_meta (
    id integer DEFAULT nextval('public.image_scan_object_meta_id_seq'::regclass) NOT NULL,
    name character varying(255),
    type character varying(255),
    image character varying(255),
    active boolean
);


ALTER TABLE public.image_scan_object_meta OWNER TO postgres;

--
-- Name: installed_app_versions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.installed_app_versions (
    id integer NOT NULL,
    installed_app_id integer,
    app_store_application_version_id integer,
    values_yaml json NOT NULL,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer,
    values_yaml_raw text,
    active boolean DEFAULT true,
    reference_value_id integer,
    reference_value_kind character varying(250)
);


ALTER TABLE public.installed_app_versions OWNER TO postgres;

--
-- Name: installed_app_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.installed_app_versions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.installed_app_versions_id_seq OWNER TO postgres;

--
-- Name: installed_app_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.installed_app_versions_id_seq OWNED BY public.installed_app_versions.id;


--
-- Name: installed_apps; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.installed_apps (
    id integer NOT NULL,
    app_id integer,
    environment_id integer,
    created_by integer,
    updated_by integer,
    created_on timestamp with time zone NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    active boolean DEFAULT true,
    status integer DEFAULT 0 NOT NULL
);


ALTER TABLE public.installed_apps OWNER TO postgres;

--
-- Name: installed_apps_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.installed_apps_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.installed_apps_id_seq OWNER TO postgres;

--
-- Name: installed_apps_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.installed_apps_id_seq OWNED BY public.installed_apps.id;


--
-- Name: job_event; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.job_event (
    id integer NOT NULL,
    event_trigger_time character varying(100) NOT NULL,
    name character varying(150) NOT NULL,
    status character varying(150) NOT NULL,
    message character varying(250),
    created_on timestamp with time zone,
    updated_on timestamp with time zone
);


ALTER TABLE public.job_event OWNER TO postgres;

--
-- Name: job_event_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.job_event_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.job_event_id_seq OWNER TO postgres;

--
-- Name: job_event_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.job_event_id_seq OWNED BY public.job_event.id;


--
-- Name: notification_settings; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notification_settings (
    id integer NOT NULL,
    app_id integer,
    env_id integer,
    pipeline_id integer,
    pipeline_type character varying(50) NOT NULL,
    event_type_id integer NOT NULL,
    config json NOT NULL,
    view_id integer NOT NULL,
    team_id integer
);


ALTER TABLE public.notification_settings OWNER TO postgres;

--
-- Name: notification_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notification_settings_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notification_settings_id_seq OWNER TO postgres;

--
-- Name: notification_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notification_settings_id_seq OWNED BY public.notification_settings.id;


--
-- Name: notification_settings_view; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notification_settings_view (
    id integer NOT NULL,
    config json NOT NULL,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer
);


ALTER TABLE public.notification_settings_view OWNER TO postgres;

--
-- Name: notification_settings_view_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notification_settings_view_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notification_settings_view_id_seq OWNER TO postgres;

--
-- Name: notification_settings_view_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notification_settings_view_id_seq OWNED BY public.notification_settings_view.id;


--
-- Name: notification_templates; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notification_templates (
    id integer NOT NULL,
    channel_type character varying(100) NOT NULL,
    node_type character varying(50) NOT NULL,
    event_type_id integer NOT NULL,
    template_name character varying(250) NOT NULL,
    template_payload text NOT NULL
);


ALTER TABLE public.notification_templates OWNER TO postgres;

--
-- Name: notification_templates_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notification_templates_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notification_templates_id_seq OWNER TO postgres;

--
-- Name: notification_templates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notification_templates_id_seq OWNED BY public.notification_templates.id;


--
-- Name: notifier_event_log; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.notifier_event_log (
    id integer NOT NULL,
    destination character varying(250) NOT NULL,
    source_id integer,
    pipeline_type character varying(100) NOT NULL,
    event_type_id integer NOT NULL,
    correlation_id character varying(250) NOT NULL,
    payload text,
    is_notification_sent boolean NOT NULL,
    event_time timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL
);


ALTER TABLE public.notifier_event_log OWNER TO postgres;

--
-- Name: notifier_event_log_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.notifier_event_log_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.notifier_event_log_id_seq OWNER TO postgres;

--
-- Name: notifier_event_log_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.notifier_event_log_id_seq OWNED BY public.notifier_event_log.id;


--
-- Name: pipeline; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pipeline (
    id integer NOT NULL,
    app_id integer,
    ci_pipeline_id integer,
    trigger_type character varying(250) NOT NULL,
    environment_id integer,
    deployment_template character varying(250),
    pipeline_name character varying(250) NOT NULL,
    deleted boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    pipeline_override text DEFAULT '{}'::text,
    pre_stage_config_yaml text,
    post_stage_config_yaml text,
    pre_trigger_type character varying(250),
    post_trigger_type character varying(250),
    pre_stage_config_map_secret_names text,
    post_stage_config_map_secret_names text,
    run_pre_stage_in_env boolean DEFAULT false,
    run_post_stage_in_env boolean DEFAULT false
);


ALTER TABLE public.pipeline OWNER TO postgres;

--
-- Name: pipeline_config_override; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pipeline_config_override (
    id integer NOT NULL,
    request_identifier character varying(250) NOT NULL,
    env_config_override_id integer,
    pipeline_override_yaml text NOT NULL,
    merged_values_yaml text NOT NULL,
    status character varying(50) NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    git_hash character varying(250),
    ci_artifact_id integer,
    pipeline_id integer,
    pipeline_release_counter integer,
    cd_workflow_id integer,
    deployment_type integer DEFAULT 0
);


ALTER TABLE public.pipeline_config_override OWNER TO postgres;

--
-- Name: pipeline_config_override_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pipeline_config_override_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.pipeline_config_override_id_seq OWNER TO postgres;

--
-- Name: pipeline_config_override_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pipeline_config_override_id_seq OWNED BY public.pipeline_config_override.id;


--
-- Name: pipeline_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pipeline_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.pipeline_id_seq OWNER TO postgres;

--
-- Name: pipeline_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pipeline_id_seq OWNED BY public.pipeline.id;


--
-- Name: pipeline_strategy; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pipeline_strategy (
    id integer NOT NULL,
    strategy character varying(250) NOT NULL,
    config text,
    created_by integer,
    updated_by integer,
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    deleted boolean,
    "default" boolean NOT NULL,
    pipeline_id integer NOT NULL
);


ALTER TABLE public.pipeline_strategy OWNER TO postgres;

--
-- Name: pipeline_strategy_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pipeline_strategy_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.pipeline_strategy_id_seq OWNER TO postgres;

--
-- Name: pipeline_strategy_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pipeline_strategy_id_seq OWNED BY public.pipeline_strategy.id;


--
-- Name: project_management_tool_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.project_management_tool_config (
    id integer NOT NULL,
    user_name character varying(250) NOT NULL,
    account_url character varying(250) NOT NULL,
    auth_token character varying(250) NOT NULL,
    commit_message_regex character varying(250) NOT NULL,
    final_issue_status character varying(250) NOT NULL,
    pipeline_stage character varying(250) NOT NULL,
    pipeline_id integer NOT NULL,
    created_on timestamp with time zone NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.project_management_tool_config OWNER TO postgres;

--
-- Name: project_management_tool_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.project_management_tool_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.project_management_tool_config_id_seq OWNER TO postgres;

--
-- Name: project_management_tool_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.project_management_tool_config_id_seq OWNED BY public.project_management_tool_config.id;


--
-- Name: role_group; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role_group (
    id integer NOT NULL,
    name character varying(100) NOT NULL,
    casbin_name character varying(100),
    description text,
    created_by integer,
    updated_by integer,
    created_on timestamp with time zone NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.role_group OWNER TO postgres;

--
-- Name: role_group_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.role_group_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.role_group_id_seq OWNER TO postgres;

--
-- Name: role_group_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.role_group_id_seq OWNED BY public.role_group.id;


--
-- Name: role_group_role_mapping; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role_group_role_mapping (
    id integer NOT NULL,
    role_group_id integer NOT NULL,
    role_id integer NOT NULL,
    created_by integer,
    updated_by integer,
    created_on timestamp with time zone NOT NULL,
    updated_on timestamp with time zone NOT NULL
);


ALTER TABLE public.role_group_role_mapping OWNER TO postgres;

--
-- Name: role_group_role_mapping_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.role_group_role_mapping_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.role_group_role_mapping_id_seq OWNER TO postgres;

--
-- Name: role_group_role_mapping_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.role_group_role_mapping_id_seq OWNED BY public.role_group_role_mapping.id;


--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.roles_id_seq OWNER TO postgres;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.roles (
    id integer DEFAULT nextval('public.roles_id_seq'::regclass) NOT NULL,
    role character varying(100) NOT NULL,
    team character varying(100),
    environment text,
    entity_name text,
    action character varying(100),
    created_by integer,
    created_on timestamp without time zone,
    updated_by integer,
    updated_on timestamp without time zone,
    entity character varying(100)
);


ALTER TABLE public.roles OWNER TO postgres;

--
-- Name: ses_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ses_config (
    id integer NOT NULL,
    region character varying(50) NOT NULL,
    access_key character varying(250) NOT NULL,
    secret_access_key character varying(250) NOT NULL,
    session_token character varying(250),
    from_email character varying(250) NOT NULL,
    config_name character varying(250),
    description character varying(500),
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer,
    owner_id integer,
    "default" boolean
);


ALTER TABLE public.ses_config OWNER TO postgres;

--
-- Name: ses_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ses_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ses_config_id_seq OWNER TO postgres;

--
-- Name: ses_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ses_config_id_seq OWNED BY public.ses_config.id;


--
-- Name: slack_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.slack_config (
    id integer NOT NULL,
    web_hook_url character varying(250) NOT NULL,
    config_name character varying(250) NOT NULL,
    description character varying(500),
    created_on timestamp with time zone,
    updated_on timestamp with time zone,
    created_by integer,
    updated_by integer,
    owner_id integer,
    team_id integer
);


ALTER TABLE public.slack_config OWNER TO postgres;

--
-- Name: slack_config_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.slack_config_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.slack_config_id_seq OWNER TO postgres;

--
-- Name: slack_config_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.slack_config_id_seq OWNED BY public.slack_config.id;


--
-- Name: team; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.team (
    id integer NOT NULL,
    name character varying(250) NOT NULL,
    active boolean NOT NULL,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL
);


ALTER TABLE public.team OWNER TO postgres;

--
-- Name: team_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.team_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.team_id_seq OWNER TO postgres;

--
-- Name: team_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.team_id_seq OWNED BY public.team.id;


--
-- Name: user_roles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.user_roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.user_roles_id_seq OWNER TO postgres;

--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_roles (
    id integer DEFAULT nextval('public.user_roles_id_seq'::regclass) NOT NULL,
    user_id integer NOT NULL,
    role_id integer NOT NULL,
    created_by integer,
    created_on timestamp without time zone,
    updated_by integer,
    updated_on timestamp without time zone
);


ALTER TABLE public.user_roles OWNER TO postgres;

--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_id_seq OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer DEFAULT nextval('public.users_id_seq'::regclass) NOT NULL,
    fname text,
    lname text,
    password text,
    access_token text,
    created_on timestamp without time zone,
    email_id character varying(100) NOT NULL,
    created_by integer,
    updated_by integer,
    updated_on timestamp without time zone,
    active boolean DEFAULT true NOT NULL
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: app id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app ALTER COLUMN id SET DEFAULT nextval('public.app_id_seq'::regclass);


--
-- Name: app_env_linkouts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_env_linkouts ALTER COLUMN id SET DEFAULT nextval('public.app_env_linkouts_id_seq'::regclass);


--
-- Name: app_level_metrics id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_level_metrics ALTER COLUMN id SET DEFAULT nextval('public.app_level_metrics_id_seq'::regclass);


--
-- Name: app_store id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store ALTER COLUMN id SET DEFAULT nextval('public.app_store_id_seq'::regclass);


--
-- Name: app_store_application_version id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store_application_version ALTER COLUMN id SET DEFAULT nextval('public.app_store_application_version_id_seq'::regclass);


--
-- Name: app_store_version_values id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store_version_values ALTER COLUMN id SET DEFAULT nextval('public.app_store_version_values_id_seq'::regclass);


--
-- Name: app_workflow id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_workflow ALTER COLUMN id SET DEFAULT nextval('public.app_workflow_id_seq'::regclass);


--
-- Name: cd_workflow id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow ALTER COLUMN id SET DEFAULT nextval('public.cd_workflow_id_seq'::regclass);


--
-- Name: cd_workflow_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow_config ALTER COLUMN id SET DEFAULT nextval('public.cd_workflow_config_id_seq'::regclass);


--
-- Name: cd_workflow_runner id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow_runner ALTER COLUMN id SET DEFAULT nextval('public.cd_workflow_runner_id_seq'::regclass);


--
-- Name: chart_env_config_override id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_env_config_override ALTER COLUMN id SET DEFAULT nextval('public.chart_env_config_override_id_seq'::regclass);


--
-- Name: chart_group id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group ALTER COLUMN id SET DEFAULT nextval('public.chart_group_id_seq'::regclass);


--
-- Name: chart_group_deployment id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_deployment ALTER COLUMN id SET DEFAULT nextval('public.chart_group_deployment_id_seq'::regclass);


--
-- Name: chart_group_entry id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_entry ALTER COLUMN id SET DEFAULT nextval('public.chart_group_entry_id_seq'::regclass);


--
-- Name: chart_repo id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_repo ALTER COLUMN id SET DEFAULT nextval('public.chart_repo_id_seq'::regclass);


--
-- Name: charts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.charts ALTER COLUMN id SET DEFAULT nextval('public.charts_id_seq'::regclass);


--
-- Name: ci_artifact id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_artifact ALTER COLUMN id SET DEFAULT nextval('public.ci_artifact_id_seq'::regclass);


--
-- Name: ci_pipeline id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline ALTER COLUMN id SET DEFAULT nextval('public.ci_pipeline_id_seq'::regclass);


--
-- Name: ci_pipeline_material id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_material ALTER COLUMN id SET DEFAULT nextval('public.ci_pipeline_material_id_seq'::regclass);


--
-- Name: ci_pipeline_scripts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_scripts ALTER COLUMN id SET DEFAULT nextval('public.ci_pipeline_scripts_id_seq'::regclass);


--
-- Name: ci_template id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_template ALTER COLUMN id SET DEFAULT nextval('public.ci_template_id_seq'::regclass);


--
-- Name: ci_workflow id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_workflow ALTER COLUMN id SET DEFAULT nextval('public.ci_workflow_id_seq'::regclass);


--
-- Name: ci_workflow_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_workflow_config ALTER COLUMN id SET DEFAULT nextval('public.ci_workflow_config_id_seq'::regclass);


--
-- Name: cluster id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster ALTER COLUMN id SET DEFAULT nextval('public.cluster_id_seq'::regclass);


--
-- Name: cluster_accounts id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_accounts ALTER COLUMN id SET DEFAULT nextval('public.cluster_accounts_id_seq'::regclass);


--
-- Name: cluster_helm_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_helm_config ALTER COLUMN id SET DEFAULT nextval('public.cluster_helm_config_id_seq'::regclass);


--
-- Name: db_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.db_config ALTER COLUMN id SET DEFAULT nextval('public.db_config_id_seq'::regclass);


--
-- Name: db_migration_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.db_migration_config ALTER COLUMN id SET DEFAULT nextval('public.db_migration_config_id_seq'::regclass);


--
-- Name: deployment_group id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group ALTER COLUMN id SET DEFAULT nextval('public.deployment_group_id_seq'::regclass);


--
-- Name: deployment_group_app id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group_app ALTER COLUMN id SET DEFAULT nextval('public.deployment_group_app_id_seq'::regclass);


--
-- Name: deployment_status id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_status ALTER COLUMN id SET DEFAULT nextval('public.deployment_status_id_seq'::regclass);


--
-- Name: env_level_app_metrics id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.env_level_app_metrics ALTER COLUMN id SET DEFAULT nextval('public.env_level_app_metrics_id_seq'::regclass);


--
-- Name: environment id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.environment ALTER COLUMN id SET DEFAULT nextval('public.environment_id_seq'::regclass);


--
-- Name: event id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event ALTER COLUMN id SET DEFAULT nextval('public.event_id_seq'::regclass);


--
-- Name: events id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events ALTER COLUMN id SET DEFAULT nextval('public.events_id_seq'::regclass);


--
-- Name: external_ci_pipeline id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.external_ci_pipeline ALTER COLUMN id SET DEFAULT nextval('public.external_ci_pipeline_id_seq'::regclass);


--
-- Name: git_material id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_material ALTER COLUMN id SET DEFAULT nextval('public.git_material_id_seq'::regclass);


--
-- Name: git_provider id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_provider ALTER COLUMN id SET DEFAULT nextval('public.git_provider_id_seq'::regclass);


--
-- Name: git_web_hook id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_web_hook ALTER COLUMN id SET DEFAULT nextval('public.git_web_hook_id_seq'::regclass);


--
-- Name: installed_app_versions id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_app_versions ALTER COLUMN id SET DEFAULT nextval('public.installed_app_versions_id_seq'::regclass);


--
-- Name: installed_apps id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_apps ALTER COLUMN id SET DEFAULT nextval('public.installed_apps_id_seq'::regclass);


--
-- Name: job_event id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.job_event ALTER COLUMN id SET DEFAULT nextval('public.job_event_id_seq'::regclass);


--
-- Name: notification_settings id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings ALTER COLUMN id SET DEFAULT nextval('public.notification_settings_id_seq'::regclass);


--
-- Name: notification_settings_view id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings_view ALTER COLUMN id SET DEFAULT nextval('public.notification_settings_view_id_seq'::regclass);


--
-- Name: notification_templates id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_templates ALTER COLUMN id SET DEFAULT nextval('public.notification_templates_id_seq'::regclass);


--
-- Name: notifier_event_log id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifier_event_log ALTER COLUMN id SET DEFAULT nextval('public.notifier_event_log_id_seq'::regclass);


--
-- Name: pipeline id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline ALTER COLUMN id SET DEFAULT nextval('public.pipeline_id_seq'::regclass);


--
-- Name: pipeline_config_override id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_config_override ALTER COLUMN id SET DEFAULT nextval('public.pipeline_config_override_id_seq'::regclass);


--
-- Name: pipeline_strategy id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_strategy ALTER COLUMN id SET DEFAULT nextval('public.pipeline_strategy_id_seq'::regclass);


--
-- Name: project_management_tool_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_management_tool_config ALTER COLUMN id SET DEFAULT nextval('public.project_management_tool_config_id_seq'::regclass);


--
-- Name: role_group id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_group ALTER COLUMN id SET DEFAULT nextval('public.role_group_id_seq'::regclass);


--
-- Name: role_group_role_mapping id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_group_role_mapping ALTER COLUMN id SET DEFAULT nextval('public.role_group_role_mapping_id_seq'::regclass);


--
-- Name: ses_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ses_config ALTER COLUMN id SET DEFAULT nextval('public.ses_config_id_seq'::regclass);


--
-- Name: slack_config id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.slack_config ALTER COLUMN id SET DEFAULT nextval('public.slack_config_id_seq'::regclass);


--
-- Name: team id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team ALTER COLUMN id SET DEFAULT nextval('public.team_id_seq'::regclass);



--
-- Name: app_env_linkouts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_env_linkouts_id_seq', 1, false);


--
-- Name: app_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_id_seq', 1, false);


--
-- Name: app_level_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_level_metrics_id_seq', 1, false);


--
-- Name: app_store_application_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_store_application_version_id_seq', 1, false);


--
-- Name: app_store_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_store_id_seq', 1, false);


--
-- Name: app_store_version_values_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_store_version_values_id_seq', 1, false);


--
-- Name: app_workflow_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_workflow_id_seq', 1, false);


--
-- Name: app_workflow_mapping_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.app_workflow_mapping_id_seq', 1, false);


--
-- Name: casbin_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.casbin_id_seq', 1, false);


--
-- Name: casbin_role_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.casbin_role_id_seq', 1, false);


--
-- Name: cd_workflow_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cd_workflow_config_id_seq', 1, false);


--
-- Name: cd_workflow_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cd_workflow_id_seq', 1, false);


--
-- Name: cd_workflow_runner_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cd_workflow_runner_id_seq', 1, false);


--
-- Name: chart_env_config_override_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.chart_env_config_override_id_seq', 1, false);


--
-- Name: chart_group_deployment_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.chart_group_deployment_id_seq', 1, false);


--
-- Name: chart_group_entry_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.chart_group_entry_id_seq', 1, false);


--
-- Name: chart_group_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.chart_group_id_seq', 1, false);


--
-- Name: chart_repo_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.chart_repo_id_seq', 4, true);


--
-- Name: charts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.charts_id_seq', 1, false);


--
-- Name: ci_artifact_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ci_artifact_id_seq', 1, false);


--
-- Name: ci_pipeline_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ci_pipeline_id_seq', 1, false);


--
-- Name: ci_pipeline_material_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ci_pipeline_material_id_seq', 1, false);


--
-- Name: ci_pipeline_scripts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ci_pipeline_scripts_id_seq', 1, false);


--
-- Name: ci_template_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ci_template_id_seq', 1, false);


--
-- Name: ci_workflow_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ci_workflow_config_id_seq', 1, false);


--
-- Name: ci_workflow_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ci_workflow_id_seq', 1, false);


--
-- Name: cluster_accounts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cluster_accounts_id_seq', 1, false);


--
-- Name: cluster_helm_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cluster_helm_config_id_seq', 1, false);


--
-- Name: cluster_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cluster_id_seq', 1, true);


--
-- Name: cluster_installed_apps_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cluster_installed_apps_id_seq', 1, false);


--
-- Name: cve_policy_control_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.cve_policy_control_id_seq', 3, true);


--
-- Name: db_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.db_config_id_seq', 1, false);


--
-- Name: db_migration_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.db_migration_config_id_seq', 1, false);


--
-- Name: deployment_group_app_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.deployment_group_app_id_seq', 1, false);


--
-- Name: deployment_group_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.deployment_group_id_seq', 1, false);


--
-- Name: deployment_status_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.deployment_status_id_seq', 1, false);


--
-- Name: env_level_app_metrics_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.env_level_app_metrics_id_seq', 1, false);


--
-- Name: environment_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.environment_id_seq', 1, false);


--
-- Name: event_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.event_id_seq', 3, true);


--
-- Name: events_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.events_id_seq', 1, false);


--
-- Name: external_ci_pipeline_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.external_ci_pipeline_id_seq', 1, false);


--
-- Name: git_material_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.git_material_id_seq', 1, false);


--
-- Name: git_provider_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.git_provider_id_seq', 1, false);


--
-- Name: git_web_hook_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.git_web_hook_id_seq', 1, false);


--
-- Name: id_seq_chart_ref; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.id_seq_chart_ref', 10, true);


--
-- Name: id_seq_config_map_app_level; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.id_seq_config_map_app_level', 1, false);


--
-- Name: id_seq_config_map_env_level; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.id_seq_config_map_env_level', 1, false);


--
-- Name: id_seq_config_map_pipeline_level; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.id_seq_config_map_pipeline_level', 1, false);


--
-- Name: id_seq_pconfig; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.id_seq_pconfig', 1, false);


--
-- Name: image_scan_deploy_info_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.image_scan_deploy_info_id_seq', 1, false);


--
-- Name: image_scan_execution_history_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.image_scan_execution_history_id_seq', 1, false);


--
-- Name: image_scan_execution_result_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.image_scan_execution_result_id_seq', 1, false);


--
-- Name: image_scan_object_meta_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.image_scan_object_meta_id_seq', 1, false);


--
-- Name: installed_app_versions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.installed_app_versions_id_seq', 1, false);


--
-- Name: installed_apps_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.installed_apps_id_seq', 1, false);


--
-- Name: job_event_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.job_event_id_seq', 1, false);


--
-- Name: notification_settings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notification_settings_id_seq', 1, false);


--
-- Name: notification_settings_view_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notification_settings_view_id_seq', 1, false);


--
-- Name: notification_templates_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notification_templates_id_seq', 12, true);


--
-- Name: notifier_event_log_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.notifier_event_log_id_seq', 1, false);


--
-- Name: pipeline_config_override_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pipeline_config_override_id_seq', 1, false);


--
-- Name: pipeline_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pipeline_id_seq', 1, false);


--
-- Name: pipeline_strategy_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.pipeline_strategy_id_seq', 1, false);


--
-- Name: project_management_tool_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.project_management_tool_config_id_seq', 1, false);


--
-- Name: role_group_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.role_group_id_seq', 1, false);


--
-- Name: role_group_role_mapping_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.role_group_role_mapping_id_seq', 1, false);


--
-- Name: roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.roles_id_seq', 1, true);


--
-- Name: ses_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.ses_config_id_seq', 1, false);


--
-- Name: slack_config_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.slack_config_id_seq', 1, false);


--
-- Name: team_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.team_id_seq', 1, false);


--
-- Name: user_roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.user_roles_id_seq', 1, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 2, true);


--
-- Name: app app_app_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app
    ADD CONSTRAINT app_app_name_key UNIQUE (app_name);


--
-- Name: app_env_linkouts app_env_linkouts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_env_linkouts
    ADD CONSTRAINT app_env_linkouts_pkey PRIMARY KEY (id);


--
-- Name: app_level_metrics app_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_level_metrics
    ADD CONSTRAINT app_metrics_pkey PRIMARY KEY (id);


--
-- Name: app app_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app
    ADD CONSTRAINT app_pkey PRIMARY KEY (id);


--
-- Name: app_store_application_version app_store_application_version_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store_application_version
    ADD CONSTRAINT app_store_application_version_pkey PRIMARY KEY (id);


--
-- Name: app_store app_store_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store
    ADD CONSTRAINT app_store_pkey PRIMARY KEY (id);


--
-- Name: app_store app_store_unique; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store
    ADD CONSTRAINT app_store_unique UNIQUE (name, chart_repo_id);


--
-- Name: app_store_version_values app_store_version_values_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store_version_values
    ADD CONSTRAINT app_store_version_values_pkey PRIMARY KEY (id);


--
-- Name: app_workflow_mapping app_workflow_mapping_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_workflow_mapping
    ADD CONSTRAINT app_workflow_mapping_pkey PRIMARY KEY (id);


--
-- Name: app_workflow app_workflow_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_workflow
    ADD CONSTRAINT app_workflow_pkey PRIMARY KEY (id);


--
-- Name: cd_workflow_config cd_workflow_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow_config
    ADD CONSTRAINT cd_workflow_config_pkey PRIMARY KEY (id);


--
-- Name: cd_workflow cd_workflow_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow
    ADD CONSTRAINT cd_workflow_pkey PRIMARY KEY (id);


--
-- Name: cd_workflow_runner cd_workflow_runner_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow_runner
    ADD CONSTRAINT cd_workflow_runner_pkey PRIMARY KEY (id);


--
-- Name: chart_env_config_override chart_env_config_override_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_env_config_override
    ADD CONSTRAINT chart_env_config_override_pkey PRIMARY KEY (id);


--
-- Name: chart_group_deployment chart_group_deployment_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_deployment
    ADD CONSTRAINT chart_group_deployment_pkey PRIMARY KEY (id);


--
-- Name: chart_group_entry chart_group_entry_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_entry
    ADD CONSTRAINT chart_group_entry_pkey PRIMARY KEY (id);


--
-- Name: chart_group chart_group_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group
    ADD CONSTRAINT chart_group_name_key UNIQUE (name);


--
-- Name: chart_group chart_group_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group
    ADD CONSTRAINT chart_group_pkey PRIMARY KEY (id);


--
-- Name: chart_repo chart_repo_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_repo
    ADD CONSTRAINT chart_repo_pkey PRIMARY KEY (id);


--
-- Name: charts charts_chart_name_chart_version_chart_repo_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.charts
    ADD CONSTRAINT charts_chart_name_chart_version_chart_repo_key UNIQUE (chart_name, chart_version, chart_repo);


--
-- Name: charts charts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.charts
    ADD CONSTRAINT charts_pkey PRIMARY KEY (id);


--
-- Name: ci_artifact ci_artifact_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_artifact
    ADD CONSTRAINT ci_artifact_pkey PRIMARY KEY (id);


--
-- Name: ci_pipeline_material ci_pipeline_material_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_material
    ADD CONSTRAINT ci_pipeline_material_pkey PRIMARY KEY (id);


--
-- Name: ci_pipeline ci_pipeline_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline
    ADD CONSTRAINT ci_pipeline_pkey PRIMARY KEY (id);


--
-- Name: ci_pipeline_scripts ci_pipeline_scripts_name_ci_pipeline_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_scripts
    ADD CONSTRAINT ci_pipeline_scripts_name_ci_pipeline_id_key UNIQUE (name, ci_pipeline_id);


--
-- Name: ci_pipeline_scripts ci_pipeline_scripts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_scripts
    ADD CONSTRAINT ci_pipeline_scripts_pkey PRIMARY KEY (id);


--
-- Name: ci_template ci_template_app_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_template
    ADD CONSTRAINT ci_template_app_id_key UNIQUE (app_id);


--
-- Name: ci_template ci_template_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_template
    ADD CONSTRAINT ci_template_pkey PRIMARY KEY (id);


--
-- Name: ci_workflow_config ci_workflow_config_ci_pipeline_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_workflow_config
    ADD CONSTRAINT ci_workflow_config_ci_pipeline_id_key UNIQUE (ci_pipeline_id);


--
-- Name: ci_workflow_config ci_workflow_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_workflow_config
    ADD CONSTRAINT ci_workflow_config_pkey PRIMARY KEY (id);


--
-- Name: ci_workflow ci_workflow_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_workflow
    ADD CONSTRAINT ci_workflow_pkey PRIMARY KEY (id);


--
-- Name: cluster_accounts cluster_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_accounts
    ADD CONSTRAINT cluster_accounts_pkey PRIMARY KEY (id);


--
-- Name: cluster_helm_config cluster_helm_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_helm_config
    ADD CONSTRAINT cluster_helm_config_pkey PRIMARY KEY (id);


--
-- Name: cluster_installed_apps cluster_installed_apps_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_installed_apps
    ADD CONSTRAINT cluster_installed_apps_pkey PRIMARY KEY (id);


--
-- Name: cluster cluster_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster
    ADD CONSTRAINT cluster_pkey PRIMARY KEY (id);


--
-- Name: cve_policy_control cve_policy_control_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cve_policy_control
    ADD CONSTRAINT cve_policy_control_pkey PRIMARY KEY (id);


--
-- Name: cve_store cve_store_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cve_store
    ADD CONSTRAINT cve_store_pkey PRIMARY KEY (name);


--
-- Name: db_config db_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.db_config
    ADD CONSTRAINT db_config_pkey PRIMARY KEY (id);


--
-- Name: db_migration_config db_migration_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.db_migration_config
    ADD CONSTRAINT db_migration_config_pkey PRIMARY KEY (id);


--
-- Name: deployment_group_app deployment_group_app_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group_app
    ADD CONSTRAINT deployment_group_app_pkey PRIMARY KEY (id);


--
-- Name: deployment_group deployment_group_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group
    ADD CONSTRAINT deployment_group_pkey PRIMARY KEY (id);


--
-- Name: docker_artifact_store docker_artifact_store_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.docker_artifact_store
    ADD CONSTRAINT docker_artifact_store_pkey PRIMARY KEY (id);


--
-- Name: deployment_status ds_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_status
    ADD CONSTRAINT ds_pkey PRIMARY KEY (id);


--
-- Name: env_level_app_metrics env_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.env_level_app_metrics
    ADD CONSTRAINT env_metrics_pkey PRIMARY KEY (id);


--
-- Name: environment environment_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.environment
    ADD CONSTRAINT environment_pkey PRIMARY KEY (id);


--
-- Name: event event_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT event_pkey PRIMARY KEY (id);


--
-- Name: events events_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);


--
-- Name: external_ci_pipeline external_ci_pipeline_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.external_ci_pipeline
    ADD CONSTRAINT external_ci_pipeline_pkey PRIMARY KEY (id);


--
-- Name: git_material git_material_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_material
    ADD CONSTRAINT git_material_pkey PRIMARY KEY (id);


--
-- Name: git_provider git_provider_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_provider
    ADD CONSTRAINT git_provider_name_key UNIQUE (name);


--
-- Name: git_provider git_provider_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_provider
    ADD CONSTRAINT git_provider_pkey PRIMARY KEY (id);


--
-- Name: git_provider git_provider_url_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_provider
    ADD CONSTRAINT git_provider_url_key UNIQUE (url);


--
-- Name: git_web_hook git_web_hook_ci_material_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_web_hook
    ADD CONSTRAINT git_web_hook_ci_material_id_key UNIQUE (ci_material_id);


--
-- Name: git_web_hook git_web_hook_git_material_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_web_hook
    ADD CONSTRAINT git_web_hook_git_material_id_key UNIQUE (git_material_id);


--
-- Name: git_web_hook git_web_hook_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_web_hook
    ADD CONSTRAINT git_web_hook_pkey PRIMARY KEY (id);


--
-- Name: helm_values helm_values_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.helm_values
    ADD CONSTRAINT helm_values_pkey PRIMARY KEY (app_name, environment);


--
-- Name: image_scan_deploy_info image_scan_deploy_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_deploy_info
    ADD CONSTRAINT image_scan_deploy_info_pkey PRIMARY KEY (id);


--
-- Name: image_scan_execution_history image_scan_execution_history_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_execution_history
    ADD CONSTRAINT image_scan_execution_history_pkey PRIMARY KEY (id);


--
-- Name: image_scan_execution_result image_scan_execution_result_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_execution_result
    ADD CONSTRAINT image_scan_execution_result_pkey PRIMARY KEY (id);


--
-- Name: image_scan_object_meta image_scan_object_meta_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_object_meta
    ADD CONSTRAINT image_scan_object_meta_pkey PRIMARY KEY (id);


--
-- Name: installed_app_versions installed_app_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_app_versions
    ADD CONSTRAINT installed_app_versions_pkey PRIMARY KEY (id);


--
-- Name: installed_apps installed_apps_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_apps
    ADD CONSTRAINT installed_apps_pkey PRIMARY KEY (id);


--
-- Name: job_event job_event_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.job_event
    ADD CONSTRAINT job_event_pkey PRIMARY KEY (id);


--
-- Name: notification_settings notification_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_pkey PRIMARY KEY (id);


--
-- Name: notification_settings_view notification_settings_view_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings_view
    ADD CONSTRAINT notification_settings_view_pkey PRIMARY KEY (id);


--
-- Name: notification_templates notification_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT notification_templates_pkey PRIMARY KEY (id);


--
-- Name: notifier_event_log notifier_event_log_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifier_event_log
    ADD CONSTRAINT notifier_event_log_pkey PRIMARY KEY (id);


--
-- Name: pipeline_config_override pipeline_config_override_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_config_override
    ADD CONSTRAINT pipeline_config_override_pkey PRIMARY KEY (id);


--
-- Name: pipeline pipeline_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline
    ADD CONSTRAINT pipeline_pkey PRIMARY KEY (id);


--
-- Name: pipeline_strategy pipeline_strategy_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_strategy
    ADD CONSTRAINT pipeline_strategy_pkey PRIMARY KEY (id);


--
-- Name: project_management_tool_config project_management_tool_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.project_management_tool_config
    ADD CONSTRAINT project_management_tool_config_pkey PRIMARY KEY (id);


--
-- Name: role_group role_group_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_group
    ADD CONSTRAINT role_group_pkey PRIMARY KEY (id);


--
-- Name: role_group_role_mapping role_group_role_mapping_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_group_role_mapping
    ADD CONSTRAINT role_group_role_mapping_pkey PRIMARY KEY (id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: ses_config ses_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ses_config
    ADD CONSTRAINT ses_config_pkey PRIMARY KEY (id);


--
-- Name: slack_config slack_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.slack_config
    ADD CONSTRAINT slack_config_pkey PRIMARY KEY (id);


--
-- Name: team team_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team
    ADD CONSTRAINT team_name_key UNIQUE (name);


--
-- Name: team team_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.team
    ADD CONSTRAINT team_pkey PRIMARY KEY (id);


--
-- Name: ci_artifact unique_ci_workflow_id; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_artifact
    ADD CONSTRAINT unique_ci_workflow_id UNIQUE (ci_workflow_id);


--
-- Name: event unq_event_name_type; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.event
    ADD CONSTRAINT unq_event_name_type UNIQUE (event_type);


--
-- Name: notification_templates unq_notification_template; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT unq_notification_template UNIQUE (channel_type, node_type, event_type_id);


--
-- Name: notification_settings unq_source; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT unq_source UNIQUE (app_id, env_id, pipeline_id, pipeline_type, event_type_id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: app_env_pipeline_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX app_env_pipeline_unique ON public.config_map_pipeline_level USING btree (app_id, environment_id, pipeline_id);


--
-- Name: app_env_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX app_env_unique ON public.config_map_env_level USING btree (app_id, environment_id);


--
-- Name: app_id_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX app_id_unique ON public.config_map_app_level USING btree (app_id);


--
-- Name: ds_app_name_index; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ds_app_name_index ON public.deployment_status USING btree (app_name);


--
-- Name: email_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX email_unique ON public.users USING btree (email_id);


--
-- Name: events_component; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX events_component ON public.events USING btree (component);


--
-- Name: events_creation_time_stamp; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX events_creation_time_stamp ON public.events USING btree (creation_time_stamp);


--
-- Name: events_kind; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX events_kind ON public.events USING btree (kind);


--
-- Name: events_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX events_name ON public.events USING btree (name);


--
-- Name: events_namespace; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX events_namespace ON public.events USING btree (namespace);


--
-- Name: events_reason; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX events_reason ON public.events USING btree (reason);


--
-- Name: events_resource_revision; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX events_resource_revision ON public.events USING btree (resource_revision);


--
-- Name: image_scan_deploy_info_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX image_scan_deploy_info_unique ON public.image_scan_deploy_info USING btree (scan_object_meta_id, object_type);


--
-- Name: role_unique; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX role_unique ON public.roles USING btree (role);


--
-- Name: app_env_linkouts app_env_linkouts_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_env_linkouts
    ADD CONSTRAINT app_env_linkouts_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: app_env_linkouts app_env_linkouts_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_env_linkouts
    ADD CONSTRAINT app_env_linkouts_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environment(id);


--
-- Name: app_level_metrics app_metrics_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_level_metrics
    ADD CONSTRAINT app_metrics_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: app_store_application_version app_store_application_version_app_store_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store_application_version
    ADD CONSTRAINT app_store_application_version_app_store_id_fkey FOREIGN KEY (app_store_id) REFERENCES public.app_store(id);


--
-- Name: app_store app_store_chart_repo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store
    ADD CONSTRAINT app_store_chart_repo_id_fkey FOREIGN KEY (chart_repo_id) REFERENCES public.chart_repo(id);


--
-- Name: app_store_version_values app_store_version_values_app_store_application_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_store_version_values
    ADD CONSTRAINT app_store_version_values_app_store_application_version_id_fkey FOREIGN KEY (app_store_application_version_id) REFERENCES public.app_store_application_version(id);


--
-- Name: app app_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app
    ADD CONSTRAINT app_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.team(id);


--
-- Name: app_workflow app_workflow_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_workflow
    ADD CONSTRAINT app_workflow_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: app_workflow_mapping app_workflow_mapping_app_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_workflow_mapping
    ADD CONSTRAINT app_workflow_mapping_app_workflow_id_fkey FOREIGN KEY (app_workflow_id) REFERENCES public.app_workflow(id);


--
-- Name: cd_workflow cd_workflow_ci_artifact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow
    ADD CONSTRAINT cd_workflow_ci_artifact_id_fkey FOREIGN KEY (ci_artifact_id) REFERENCES public.ci_artifact(id);


--
-- Name: cd_workflow_config cd_workflow_config_cd_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow_config
    ADD CONSTRAINT cd_workflow_config_cd_pipeline_id_fkey FOREIGN KEY (cd_pipeline_id) REFERENCES public.pipeline(id);


--
-- Name: cd_workflow cd_workflow_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow
    ADD CONSTRAINT cd_workflow_pipeline_id_fkey FOREIGN KEY (pipeline_id) REFERENCES public.pipeline(id);


--
-- Name: cd_workflow_runner cd_workflow_runner_cd_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cd_workflow_runner
    ADD CONSTRAINT cd_workflow_runner_cd_workflow_id_fkey FOREIGN KEY (cd_workflow_id) REFERENCES public.cd_workflow(id);


--
-- Name: chart_env_config_override chart_env_config_override_chart_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_env_config_override
    ADD CONSTRAINT chart_env_config_override_chart_id_fkey FOREIGN KEY (chart_id) REFERENCES public.charts(id);


--
-- Name: chart_env_config_override chart_env_config_override_target_environment_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_env_config_override
    ADD CONSTRAINT chart_env_config_override_target_environment_fkey FOREIGN KEY (target_environment) REFERENCES public.environment(id);


--
-- Name: chart_group_deployment chart_group_deployment_chart_group_entry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_deployment
    ADD CONSTRAINT chart_group_deployment_chart_group_entry_id_fkey FOREIGN KEY (chart_group_entry_id) REFERENCES public.chart_group_entry(id);


--
-- Name: chart_group_deployment chart_group_deployment_chart_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_deployment
    ADD CONSTRAINT chart_group_deployment_chart_group_id_fkey FOREIGN KEY (chart_group_id) REFERENCES public.chart_group(id);


--
-- Name: chart_group_deployment chart_group_deployment_installed_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_deployment
    ADD CONSTRAINT chart_group_deployment_installed_app_id_fkey FOREIGN KEY (installed_app_id) REFERENCES public.installed_apps(id);


--
-- Name: chart_group_entry chart_group_entry_app_store_application_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_entry
    ADD CONSTRAINT chart_group_entry_app_store_application_version_id_fkey FOREIGN KEY (app_store_application_version_id) REFERENCES public.app_store_application_version(id);


--
-- Name: chart_group_entry chart_group_entry_app_store_values_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_entry
    ADD CONSTRAINT chart_group_entry_app_store_values_version_id_fkey FOREIGN KEY (app_store_values_version_id) REFERENCES public.app_store_version_values(id);


--
-- Name: chart_group_entry chart_group_entry_chart_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.chart_group_entry
    ADD CONSTRAINT chart_group_entry_chart_group_id_fkey FOREIGN KEY (chart_group_id) REFERENCES public.chart_group(id);


--
-- Name: charts charts_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.charts
    ADD CONSTRAINT charts_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: charts charts_chart_repo_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.charts
    ADD CONSTRAINT charts_chart_repo_id_fkey FOREIGN KEY (chart_repo_id) REFERENCES public.chart_repo(id);


--
-- Name: ci_artifact ci_artifact_ci_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_artifact
    ADD CONSTRAINT ci_artifact_ci_workflow_id_fkey FOREIGN KEY (ci_workflow_id) REFERENCES public.ci_workflow(id);


--
-- Name: ci_artifact ci_artifact_parent_ci_artifact_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_artifact
    ADD CONSTRAINT ci_artifact_parent_ci_artifact_fkey FOREIGN KEY (parent_ci_artifact) REFERENCES public.ci_artifact(id);


--
-- Name: ci_artifact ci_artifact_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_artifact
    ADD CONSTRAINT ci_artifact_pipeline_id_fkey FOREIGN KEY (pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: ci_pipeline ci_pipeline_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline
    ADD CONSTRAINT ci_pipeline_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: ci_pipeline ci_pipeline_ci_template_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline
    ADD CONSTRAINT ci_pipeline_ci_template_id_fkey FOREIGN KEY (ci_template_id) REFERENCES public.ci_template(id);


--
-- Name: ci_pipeline_material ci_pipeline_material_ci_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_material
    ADD CONSTRAINT ci_pipeline_material_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: ci_pipeline_material ci_pipeline_material_git_material_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_material
    ADD CONSTRAINT ci_pipeline_material_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES public.git_material(id);


--
-- Name: ci_pipeline ci_pipeline_parent_ci_pipeline_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline
    ADD CONSTRAINT ci_pipeline_parent_ci_pipeline_fkey FOREIGN KEY (parent_ci_pipeline) REFERENCES public.ci_pipeline(id);


--
-- Name: ci_pipeline_scripts ci_pipeline_scripts_ci_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_pipeline_scripts
    ADD CONSTRAINT ci_pipeline_scripts_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: ci_template ci_template_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_template
    ADD CONSTRAINT ci_template_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: ci_template ci_template_docker_registry_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_template
    ADD CONSTRAINT ci_template_docker_registry_id_fkey FOREIGN KEY (docker_registry_id) REFERENCES public.docker_artifact_store(id);


--
-- Name: ci_template ci_template_git_material_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_template
    ADD CONSTRAINT ci_template_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES public.git_material(id);


--
-- Name: ci_workflow ci_workflow_ci_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_workflow
    ADD CONSTRAINT ci_workflow_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: ci_workflow_config ci_workflow_config_ci_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ci_workflow_config
    ADD CONSTRAINT ci_workflow_config_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: cluster_accounts cluster_accounts_cluster_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_accounts
    ADD CONSTRAINT cluster_accounts_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES public.cluster(id);


--
-- Name: cluster_helm_config cluster_helm_config_cluster_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_helm_config
    ADD CONSTRAINT cluster_helm_config_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES public.cluster(id);


--
-- Name: cluster_installed_apps cluster_installed_apps_cluster_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_installed_apps
    ADD CONSTRAINT cluster_installed_apps_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES public.cluster(id);


--
-- Name: cluster_installed_apps cluster_installed_apps_installed_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cluster_installed_apps
    ADD CONSTRAINT cluster_installed_apps_installed_app_id_fkey FOREIGN KEY (installed_app_id) REFERENCES public.installed_apps(id);


--
-- Name: config_map_app_level config_map_app_level_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config_map_app_level
    ADD CONSTRAINT config_map_app_level_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: config_map_env_level config_map_env_level_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config_map_env_level
    ADD CONSTRAINT config_map_env_level_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: config_map_env_level config_map_env_level_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config_map_env_level
    ADD CONSTRAINT config_map_env_level_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environment(id);


--
-- Name: config_map_pipeline_level config_map_pipeline_level_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config_map_pipeline_level
    ADD CONSTRAINT config_map_pipeline_level_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: config_map_pipeline_level config_map_pipeline_level_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config_map_pipeline_level
    ADD CONSTRAINT config_map_pipeline_level_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environment(id);


--
-- Name: config_map_pipeline_level config_map_pipeline_level_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config_map_pipeline_level
    ADD CONSTRAINT config_map_pipeline_level_pipeline_id_fkey FOREIGN KEY (pipeline_id) REFERENCES public.pipeline(id);


--
-- Name: cve_policy_control cve_policy_control_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cve_policy_control
    ADD CONSTRAINT cve_policy_control_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: cve_policy_control cve_policy_control_cluster_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cve_policy_control
    ADD CONSTRAINT cve_policy_control_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES public.cluster(id);


--
-- Name: cve_policy_control cve_policy_control_cve_store_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cve_policy_control
    ADD CONSTRAINT cve_policy_control_cve_store_id_fkey FOREIGN KEY (cve_store_id) REFERENCES public.cve_store(name);


--
-- Name: cve_policy_control cve_policy_control_env_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cve_policy_control
    ADD CONSTRAINT cve_policy_control_env_id_fkey FOREIGN KEY (env_id) REFERENCES public.environment(id);


--
-- Name: db_migration_config db_migration_config_db_config_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.db_migration_config
    ADD CONSTRAINT db_migration_config_db_config_id_fkey FOREIGN KEY (db_config_id) REFERENCES public.db_config(id);


--
-- Name: db_migration_config db_migration_config_git_material_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.db_migration_config
    ADD CONSTRAINT db_migration_config_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES public.git_material(id);


--
-- Name: db_migration_config db_migration_config_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.db_migration_config
    ADD CONSTRAINT db_migration_config_pipeline_id_fkey FOREIGN KEY (pipeline_id) REFERENCES public.pipeline(id);


--
-- Name: deployment_group_app deployment_group_app_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group_app
    ADD CONSTRAINT deployment_group_app_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: deployment_group_app deployment_group_app_deployment_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group_app
    ADD CONSTRAINT deployment_group_app_deployment_group_id_fkey FOREIGN KEY (deployment_group_id) REFERENCES public.deployment_group(id);


--
-- Name: deployment_group deployment_group_ci_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group
    ADD CONSTRAINT deployment_group_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: deployment_group deployment_group_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_group
    ADD CONSTRAINT deployment_group_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environment(id);


--
-- Name: deployment_status deployment_status_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_status
    ADD CONSTRAINT deployment_status_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: deployment_status deployment_status_env_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.deployment_status
    ADD CONSTRAINT deployment_status_env_id_fkey FOREIGN KEY (env_id) REFERENCES public.environment(id);


--
-- Name: env_level_app_metrics env_level_env_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.env_level_app_metrics
    ADD CONSTRAINT env_level_env_id_fkey FOREIGN KEY (env_id) REFERENCES public.environment(id);


--
-- Name: env_level_app_metrics env_metrics_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.env_level_app_metrics
    ADD CONSTRAINT env_metrics_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: environment environment_cluster_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.environment
    ADD CONSTRAINT environment_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES public.cluster(id);


--
-- Name: external_ci_pipeline external_ci_pipeline_ci_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.external_ci_pipeline
    ADD CONSTRAINT external_ci_pipeline_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: git_material git_material_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_material
    ADD CONSTRAINT git_material_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: git_material git_material_git_provider_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_material
    ADD CONSTRAINT git_material_git_provider_id_fkey FOREIGN KEY (git_provider_id) REFERENCES public.git_provider(id);


--
-- Name: git_web_hook git_web_hook_ci_material_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_web_hook
    ADD CONSTRAINT git_web_hook_ci_material_id_fkey FOREIGN KEY (ci_material_id) REFERENCES public.ci_pipeline_material(id);


--
-- Name: git_web_hook git_web_hook_git_material_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.git_web_hook
    ADD CONSTRAINT git_web_hook_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES public.git_material(id);


--
-- Name: image_scan_deploy_info image_scan_deploy_info_cluster_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_deploy_info
    ADD CONSTRAINT image_scan_deploy_info_cluster_id_fkey FOREIGN KEY (cluster_id) REFERENCES public.cluster(id);


--
-- Name: image_scan_deploy_info image_scan_deploy_info_env_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_deploy_info
    ADD CONSTRAINT image_scan_deploy_info_env_id_fkey FOREIGN KEY (env_id) REFERENCES public.environment(id);


--
-- Name: image_scan_deploy_info image_scan_deploy_info_scan_object_meta_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_deploy_info
    ADD CONSTRAINT image_scan_deploy_info_scan_object_meta_id_fkey FOREIGN KEY (scan_object_meta_id) REFERENCES public.image_scan_object_meta(id);


--
-- Name: image_scan_execution_result image_scan_execution_result_cve_store_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_execution_result
    ADD CONSTRAINT image_scan_execution_result_cve_store_name_fkey FOREIGN KEY (cve_store_name) REFERENCES public.cve_store(name);


--
-- Name: image_scan_execution_result image_scan_execution_result_image_scan_execution_history_i_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.image_scan_execution_result
    ADD CONSTRAINT image_scan_execution_result_image_scan_execution_history_i_fkey FOREIGN KEY (image_scan_execution_history_id) REFERENCES public.image_scan_execution_history(id);


--
-- Name: installed_app_versions installed_app_versions_app_store_application_version_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_app_versions
    ADD CONSTRAINT installed_app_versions_app_store_application_version_id_fkey FOREIGN KEY (app_store_application_version_id) REFERENCES public.app_store_application_version(id);


--
-- Name: installed_app_versions installed_app_versions_installed_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_app_versions
    ADD CONSTRAINT installed_app_versions_installed_app_id_fkey FOREIGN KEY (installed_app_id) REFERENCES public.installed_apps(id);


--
-- Name: installed_apps installed_apps_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_apps
    ADD CONSTRAINT installed_apps_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: installed_apps installed_apps_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.installed_apps
    ADD CONSTRAINT installed_apps_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environment(id);


--
-- Name: notification_settings notification_settings_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: notification_settings notification_settings_env_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_env_id_fkey FOREIGN KEY (env_id) REFERENCES public.environment(id);


--
-- Name: notification_templates notification_settings_event_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_templates
    ADD CONSTRAINT notification_settings_event_type_id_fkey FOREIGN KEY (event_type_id) REFERENCES public.event(id);


--
-- Name: notification_settings notification_settings_event_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_event_type_id_fkey FOREIGN KEY (event_type_id) REFERENCES public.event(id);


--
-- Name: notification_settings notification_settings_event_view_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_event_view_id_fkey FOREIGN KEY (view_id) REFERENCES public.notification_settings_view(id);


--
-- Name: notification_settings notification_settings_team_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notification_settings
    ADD CONSTRAINT notification_settings_team_id_fkey FOREIGN KEY (team_id) REFERENCES public.team(id);


--
-- Name: notifier_event_log notifier_event_log_event_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.notifier_event_log
    ADD CONSTRAINT notifier_event_log_event_type_id_fkey FOREIGN KEY (event_type_id) REFERENCES public.event(id);


--
-- Name: pipeline pipeline_app_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline
    ADD CONSTRAINT pipeline_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);


--
-- Name: pipeline pipeline_ci_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline
    ADD CONSTRAINT pipeline_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES public.ci_pipeline(id);


--
-- Name: pipeline_config_override pipeline_config_override_cd_workflow_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_config_override
    ADD CONSTRAINT pipeline_config_override_cd_workflow_id_fkey FOREIGN KEY (cd_workflow_id) REFERENCES public.cd_workflow(id);


--
-- Name: pipeline_config_override pipeline_config_override_ci_artifact_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_config_override
    ADD CONSTRAINT pipeline_config_override_ci_artifact_id_fkey FOREIGN KEY (ci_artifact_id) REFERENCES public.ci_artifact(id);


--
-- Name: pipeline_config_override pipeline_config_override_env_config_override_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_config_override
    ADD CONSTRAINT pipeline_config_override_env_config_override_id_fkey FOREIGN KEY (env_config_override_id) REFERENCES public.chart_env_config_override(id);


--
-- Name: pipeline_config_override pipeline_config_override_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_config_override
    ADD CONSTRAINT pipeline_config_override_pipeline_id_fkey FOREIGN KEY (pipeline_id) REFERENCES public.pipeline(id);


--
-- Name: pipeline pipeline_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline
    ADD CONSTRAINT pipeline_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environment(id);


--
-- Name: pipeline_strategy pipeline_strategy_pipeline_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pipeline_strategy
    ADD CONSTRAINT pipeline_strategy_pipeline_id_fkey FOREIGN KEY (pipeline_id) REFERENCES public.pipeline(id);


--
-- Name: role_group_role_mapping role_group_role_mapping_role_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_group_role_mapping
    ADD CONSTRAINT role_group_role_mapping_role_group_id_fkey FOREIGN KEY (role_group_id) REFERENCES public.role_group(id);


--
-- Name: role_group_role_mapping role_group_role_mapping_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_group_role_mapping
    ADD CONSTRAINT role_group_role_mapping_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: ses_config ses_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ses_config
    ADD CONSTRAINT ses_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id);


--
-- Name: slack_config slack_team_name_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.slack_config
    ADD CONSTRAINT slack_team_name_fkey FOREIGN KEY (team_id) REFERENCES public.team(id);


--
-- Name: user_roles user_roles_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: slack_config users_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.slack_config
    ADD CONSTRAINT users_fkey FOREIGN KEY (owner_id) REFERENCES public.users(id);


INSERT INTO "public"."chart_ref" ("id", "location", "version", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
('10', 'reference-chart_3-9-0', '3.9.0', 't', 't', 'now()', '1', 'now()', '1'),
('9', 'reference-chart_3-8-0', '3.8.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('1', 'reference-app-rolling', '2.0.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('2', 'reference-chart_3-1-0', '3.1.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('3', 'reference-chart_3-2-0', '3.2.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('4', 'reference-chart_3-3-0', '3.3.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('5', 'reference-chart_3-4-0', '3.4.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('6', 'reference-chart_3-5-0', '3.5.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('7', 'reference-chart_3-6-0', '3.6.0', 'f', 'f', 'now()', '1', 'now()', '1'),
('1', 'reference-chart_3-7-0', '3.7.0', 'f', 'f', 'now()', '1', 'now()', '1');



INSERT INTO "public"."chart_repo" ("id", "name", "url", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "external") VALUES
('1', 'default-chartmuseum', 'http://devtron-chartmuseum.devtroncd:8080/', 't', 't', 'now()', '1', 'now()', '1', 'f'),
('2', 'stable', 'https://kubernetes-charts.storage.googleapis.com', 'f', 't', 'now()', '1', 'now()', '1', 't'),
('3', 'incubator', 'https://kubernetes-charts-incubator.storage.googleapis.com', 'f', 't', 'now()', '1', 'now()', '1', 't'),
('4', 'devtron-charts', 'https://devtron-charts.s3.us-east-2.amazonaws.com/charts', 'f', 't', 'now()', '1', 'now()', '1', 't');

INSERT INTO "public"."cluster" ("id", "cluster_name", "active", "created_on", "created_by", "updated_on", "updated_by", "server_url", "config", "prometheus_endpoint", "cd_argo_setup", "p_username", "p_password", "p_tls_client_cert", "p_tls_client_key") VALUES
('1', 'default_cluster', 't', 'now()', '1', 'now()', '1', 'https://kubernetes.default.svc', '{}', NULL, 'f', NULL, NULL, NULL, NULL);

INSERT INTO "public"."cve_policy_control" ("id", "global", "cluster_id", "env_id", "app_id", "cve_store_id", "action", "severity", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 't', NULL, NULL, NULL, NULL, '1', '2', 'f', 'now()', '1', 'now()', '1'),
('2', 't', NULL, NULL, NULL, NULL, '1', '1', 'f', 'now()', '1', 'now()', '1'),
('3', 't', NULL, NULL, NULL, NULL, '1', '0', 'f', 'now()', '1', 'now()', '1');

INSERT INTO "public"."event" ("id", "event_type", "description") VALUES
('1', 'TRIGGER', ''),
('2', 'SUCCESS', ''),
('3', 'FAIL', '');

INSERT INTO "public"."notification_templates" ("id", "channel_type", "node_type", "event_type_id", "template_name", "template_payload") VALUES
('1', 'slack', 'CI', '1', 'CI trigger template', '{
    "text": ":arrow_forward: Build pipeline Triggered |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":arrow_forward: *Build Pipeline triggered*\n{{eventTime}} \n Triggered by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url": "https://github.com/devtron-labs/wp-content/uploads/2020/06/img-build-notification@2x.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Pipeline*\n{{pipelineName}}"
                }
            ]
        },
        {{#ciMaterials}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/ciMaterials}}
        {
            "type": "actions",
            "elements": [{
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "View Details"
                }
                {{#buildHistoryLink}}
                    ,
                    "url": "{{& buildHistoryLink}}"
                {{/buildHistoryLink}}
            }]
        }
    ]
}'),
('2', 'ses', 'CI', '1', 'CI trigger ses template', '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI triggered for app: {{appName}}",
 "html": "<b>CI triggered on pipeline: {{pipelineName}}</b>"
}'),
('3', 'slack', 'CI', '2', 'CI success template', '{
  "text": ":tada: Build pipeline Successful |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
  "blocks": [
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "\n"
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": ":tada: *Build Pipeline successful*\n{{eventTime}} \n Triggered by {{triggeredBy}}"
      },
      "accessory": {
        "type": "image",
        "image_url": "https://github.com/devtron-labs/wp-content/uploads/2020/06/img-build-notification@2x.png",
        "alt_text": "calendar thumbnail"
      }
    },
    {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
          "text": "*Application*\n{{appName}}"
        },
        {
          "type": "mrkdwn",
          "text": "*Pipeline*\n{{pipelineName}}"
        }
      ]
    },
    {{#ciMaterials}}
     {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
           "text": "*Branch*\n`{{appName}}/{{branch}}`"
        },
        {
          "type": "mrkdwn",
          "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
        }
      ]
    },
    {{/ciMaterials}}
    {
      "type": "actions",
      "elements": [
        {
          "type": "button",
          "text": {
            "type": "plain_text",
            "text": "View Details"
          }
          {{#buildHistoryLink}}
            ,
            "url": "{{& buildHistoryLink}}"
          {{/buildHistoryLink}}
        }
      ]
    }
  ]
}'),
('4', 'ses', 'CI', '2', 'CI success ses template', '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI success for app: {{appName}}",
 "html": "<b>CI success on pipeline: {{pipelineName}}</b><br><b>docker image: {{{dockerImageUrl}}}</b><br><b>Source: {{source}}</b><br>"
}'),
('5', 'slack', 'CI', '3', 'CI fail template', '{
    "text": ":x: Build pipeline Failed |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":x: *Build Pipeline failed*\n{{eventTime}} \n Triggered by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url": "https://github.com/devtron-labs/wp-content/uploads/2020/06/img-build-notification@2x.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Pipeline*\n{{pipelineName}}"
                }
            ]
        },
        {{#ciMaterials}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/ciMaterials}}
        {
            "type": "actions",
            "elements": [{
                "type": "button",
                "text": {
                    "type": "plain_text",
                    "text": "View Details"
                }
                  {{#buildHistoryLink}}
                    ,
                    "url": "{{& buildHistoryLink}}"
                   {{/buildHistoryLink}}
            }]
        }
    ]
}'),
('6', 'ses', 'CI', '3', 'CI failed ses template', '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CI failed for app: {{appName}}",
 "html": "<b>CI failed on pipeline: {{pipelineName}}</b><br><b>build name: {{buildName}}</b><br><b>Pod status: {{podStatus}}</b><br><b>message: {{message}}</b>"
}'),
('7', 'slack', 'CD', '1', 'CD trigger template', '{
    "text": ":arrow_forward: Deployment pipeline Triggered |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":arrow_forward: *Deployment Pipeline triggered on {{envName}}*\n{{eventTime}} \n by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url":"https://github.com/devtron-labs/wp-content/uploads/2020/06/img-deployment-notification@2x.png",
                "alt_text": "Deploy Pipeline Triggered"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}\n*Pipeline*\n{{pipelineName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Environment*\n{{envName}}\n*Stage*\n{{stage}}"
                }
            ]
        },
        {{#ciMaterials}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
             "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/ciMaterials}}
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Docker Image*\n`{{dockerImg}}`"
            }
        },
        {
            "type": "actions",
            "elements": [{
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "View Pipeline",
                        "emoji": true
                    }
                    {{#deploymentHistoryLink}}
                    ,
                    "url": "{{& deploymentHistoryLink}}"
                      {{/deploymentHistoryLink}}
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "App details",
                        "emoji": true
                    }
                    {{#appDetailsLink}}
                    ,
                    "url": "{{& appDetailsLink}}"
                      {{/appDetailsLink}}
                }
            ]
        }
    ]
}'),
('8', 'ses', 'CD', '1', 'CD trigger ses template', '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD triggered for app: {{appName}} on environment: {{environmentName}}",
 "html": "<b>CD triggered for app: {{appName}} on environment: {{environmentName}}</b> <br> <b>Docker image: {{{dockerImageUrl}}}</b> <br> <b>Source snapshot: {{source}}</b> <br> <b>pipeline: {{pipelineName}}</b>"
}'),
('9', 'slack', 'CD', '2', 'CD success template', '{
    "text": ":tada: Deployment pipeline Successful |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":tada: *Deployment Pipeline successful on {{envName}}*\n{{eventTime}} \n by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url":"https://github.com/devtron-labs/wp-content/uploads/2020/06/img-deployment-notification@2x.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}\n*Pipeline*\n{{pipelineName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Environment*\n{{envName}}\n*Stage*\n{{stage}}"
                }
            ]
        },
        {{#ciMaterials}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
             "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/ciMaterials}}
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Docker Image*\n`{{dockerImg}}`"
            }
        },
        {
            "type": "actions",
            "elements": [{
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "View Pipeline",
                        "emoji": true
                    }
                    {{#deploymentHistoryLink}}
                    ,
                    "url": "{{& deploymentHistoryLink}}"
                      {{/deploymentHistoryLink}}
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "App details",
                        "emoji": true
                    }
                    {{#appDetailsLink}}
                    ,
                    "url": "{{& appDetailsLink}}"
                      {{/appDetailsLink}}
                }
            ]
        }
    ]
}'),
('10', 'ses', 'CD', '2', 'CD success ses template', '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD success for app: {{appName}} on environment: {{environmentName}}",
 "html": "<b>CD success for app: {{appName}} on environment: {{environmentName}}</b>"
}'),
('11', 'slack', 'CD', '3', 'CD failed template', '{
    "text": ":x: Deployment pipeline Failed |  {{#ciMaterials}} Branch > {{branch}} {{/ciMaterials}} | Application > {{appName}}",
    "blocks": [{
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "\n"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": ":x: *Deployment Pipeline failed on {{envName}}*\n{{eventTime}} \n by {{triggeredBy}}"
            },
            "accessory": {
                "type": "image",
                "image_url":"https://github.com/devtron-labs/wp-content/uploads/2020/06/img-deployment-notification@2x.png",
                "alt_text": "calendar thumbnail"
            }
        },
        {
            "type": "divider"
        },
        {
            "type": "section",
            "fields": [{
                    "type": "mrkdwn",
                    "text": "*Application*\n{{appName}}\n*Pipeline*\n{{pipelineName}}"
                },
                {
                    "type": "mrkdwn",
                    "text": "*Environment*\n{{envName}}\n*Stage*\n{{stage}}"
                }
            ]
        },
        {{#ciMaterials}}
        {
        "type": "section",
        "fields": [
            {
            "type": "mrkdwn",
            "text": "*Branch*\n`{{appName}}/{{branch}}`"
            },
            {
            "type": "mrkdwn",
            "text": "*Commit*\n<{{& commitLink}}|{{commit}}>"
            }
        ]
        },
        {{/ciMaterials}}
        {
            "type": "section",
            "text": {
                "type": "mrkdwn",
                "text": "*Docker Image*\n`{{dockerImg}}`"
            }
        },
        {
            "type": "actions",
            "elements": [{
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "View Pipeline",
                        "emoji": true
                    }
                    {{#deploymentHistoryLink}}
                    ,
                    "url": "{{& deploymentHistoryLink}}"
                      {{/deploymentHistoryLink}}
                },
                {
                    "type": "button",
                    "text": {
                        "type": "plain_text",
                        "text": "App details",
                        "emoji": true
                    }
                    {{#appDetailsLink}}
                    ,
                    "url": "{{& appDetailsLink}}"
                      {{/appDetailsLink}}
                }
            ]
        }
    ]
}'),
('12', 'ses', 'CD', '3', 'CD failed ses template', '{"from": "{{fromEmail}}",
 "to": "{{toEmail}}",
 "subject": "CD failed for app: {{appName}} on environment: {{environmentName}}",
 "html": "<b>CD failed for app: {{appName}} on environment: {{environmentName}}</b>"
}');

INSERT INTO "public"."roles" ("id", "role", "team", "environment", "entity_name", "action", "created_by", "created_on", "updated_by", "updated_on", "entity") VALUES
('1', 'role:super-admin___', NULL, NULL, NULL, 'super-admin', NULL, NULL, NULL, NULL, NULL);

INSERT INTO "public"."users" ("id", "fname", "lname", "password", "access_token", "created_on", "email_id", "created_by", "updated_by", "updated_on", "active") VALUES
('1', NULL, NULL, NULL, NULL, NULL, 'system', NULL, NULL, NULL, 't'),
('2', NULL, NULL, NULL, NULL, NULL, 'admin', NULL, NULL, NULL, 't');

INSERT INTO "public"."user_roles" ("id", "user_id", "role_id", "created_by", "created_on", "updated_by", "updated_on") VALUES
('1', '2', '1', NULL, NULL, NULL, NULL);

INSERT INTO "public"."git_provider" ("id", "name", "url", "user_name", "password", "ssh_key", "access_token", "auth_mode", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'Github Public', 'github.com', NULL, NULL, NULL, NULL, 'ANONYMOUS', 't', 'now()', '1', 'now()', '1');

INSERT INTO "public"."team" ("id", "name", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'devtron', 't', 'now()', '1', 'now()', '1');

INSERT INTO "public"."environment" ("id", "environment_name", "cluster_id", "active", "created_on", "created_by", "updated_on", "updated_by", "default", "namespace", "grafana_datasource_id") VALUES
('1', 'devtron', '1', 't', 'now()', '1', 'now()', '1', 'f', 'devtron', '0');
--
-- PostgreSQL database dump complete
--
