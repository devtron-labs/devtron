--
-- Name: user_attributes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE IF NOT EXISTS public.user_attributes (
                                                   email_id varchar(500) NOT NULL,
                                                   user_data json NOT NULL,
                                                   created_on timestamp with time zone,
                                                   updated_on timestamp with time zone,
                                                   created_by integer,
                                                   updated_by integer,
                                                   PRIMARY KEY ("email_id")
);


ALTER TABLE public.user_attributes OWNER TO postgres;