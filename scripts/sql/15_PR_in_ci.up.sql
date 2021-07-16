--
-- Name: git_host_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.git_host_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: git_host; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.git_host (
     id INTEGER NOT NULL DEFAULT nextval('git_host_id_seq'::regclass),
     name character varying(250) NOT NULL,
     active bool NOT NULL,
     webhook_url character varying(500),
     webhook_secret character varying(250),
     event_type_header character varying(250),
     secret_header character varying(250),
     secret_validator character varying(250),
     created_on timestamptz NOT NULL,
     created_by INTEGER NOT NULL,
     updated_on timestamptz,
     updated_by INTEGER,
     PRIMARY KEY ("id"),
     UNIQUE(name)
);


---- Insert master data into git_host
INSERT INTO git_host (name, created_on, created_by, active, webhook_url, webhook_secret, event_type_header, secret_header, secret_validator)
VALUES ('Github', NOW(), 1, 't', '/orchestrator/webhook/git/1', MD5(random()::text), 'X-GitHub-Event', 'X-Hub-Signature' , 'SHA-1'),
       ('Bitbucket Cloud', NOW(), 1, 't', '/orchestrator/webhook/git/2/' || MD5(random()::text), NULL, 'X-Event-Key', NULL, 'URL_APPEND');


---- add column in git_provider (git_host.id)
ALTER TABLE git_provider
    ADD COLUMN git_host_id INTEGER;



---- Add Foreign key constraint on git_host_id in Table git_provider
ALTER TABLE git_provider
    ADD CONSTRAINT git_host_id_fkey FOREIGN KEY (git_host_id) REFERENCES public.git_host(id);