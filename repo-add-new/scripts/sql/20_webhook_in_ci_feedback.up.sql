--
-- Name: webhook_event_data_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.webhook_event_data_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE CACHE 1;


--
-- Name: webhook_event_data; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.webhook_event_data
(
    id           INTEGER                NOT NULL DEFAULT nextval('webhook_event_data_id_seq'::regclass),
    git_host_id  INTEGER                NOT NULL,
    event_type   character varying(250) NOT NULL,
    payload_json JSON                   NOT NULL,
    created_on   timestamptz            NOT NULL,
    PRIMARY KEY ("id")
);


---- Add Foreign key constraint on git_host_id in Table webhook_event_data
ALTER TABLE webhook_event_data
    ADD CONSTRAINT webhook_event_data_ghid_fkey FOREIGN KEY (git_host_id) REFERENCES public.git_host(id);
