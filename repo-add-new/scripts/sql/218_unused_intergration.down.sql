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
