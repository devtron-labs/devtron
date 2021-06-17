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
     created_on timestamptz NOT NULL,
     created_by INTEGER NOT NULL,
     updated_on timestamptz,
     updated_by INTEGER,
     PRIMARY KEY ("id"),
     UNIQUE(name)
);


---- Insert master data into git_host
INSERT INTO git_host (name, created_on, created_by, active, webhook_url, webhook_secret)
VALUES ('Github', NOW(), 1, 't', '/orchestrator/webhook/github', MD5(random()::text)),
       ('Bitbucket Cloud', NOW(), 1, 't', '/orchestrator/webhook/bitbucket/' || MD5(random()::text), NULL),
       ('Other', NOW(), 1, 't', NULL, NULL);


---- add column in git_provider (git_host.id)
ALTER TABLE git_provider
    ADD COLUMN git_host_id INTEGER;


---- Data migration of git_host_id in git_provider table
UPDATE git_provider
SET git_host_id = (CASE WHEN LOWER(url) like '%github.com%' THEN (select id from git_host where name = 'Github')
                        WHEN LOWER(url) like '%bitbucket.com%' THEN (select id from git_host where name = 'Bitbucket Cloud')
                        ELSE  (select id from git_host where name = 'Other')
    END)
WHERE git_host_id is NULL;


---- Add Foreign key constraint on git_host_id in Table git_provider
ALTER TABLE git_provider
    ADD CONSTRAINT git_host_id_fkey FOREIGN KEY (git_host_id) REFERENCES public.git_host(id);


---- Add NOT NULL constraint on git_host_id in Table git_provider with Not NULL
ALTER TABLE git_provider ALTER COLUMN git_host_id SET NOT NULL;