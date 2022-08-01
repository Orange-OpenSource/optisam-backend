--
-- PostgreSQL database dump
--

/**********
--
-- Acccount Database 
-- 
*********/

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: rdsAdmin
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO "rdsAdmin";

--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: rdsAdmin
--

COMMENT ON SCHEMA public IS 'standard public schema';


--
-- Name: audit_status; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.audit_status AS ENUM (
    'DELETED',
    'UPDATED'
);


ALTER TYPE public.audit_status OWNER TO optisam;

--
-- Name: scope_types; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.scope_types AS ENUM (
    'GENERIC',
    'SPECIFIC'
);


ALTER TYPE public.scope_types OWNER TO optisam;


--
-- Name: group_ownership; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.group_ownership (
    group_id integer NOT NULL,
    user_id character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now()
);


ALTER TABLE public.group_ownership OWNER TO optisam;

--
-- Name: groups; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.groups (
    id integer NOT NULL,
    name character varying NOT NULL,
    fully_qualified_name public.ltree,
    scopes text[],
    parent_id integer,
    created_by character varying,
    created_on timestamp without time zone DEFAULT now()
);


ALTER TABLE public.groups OWNER TO optisam;

--
-- Name: groups_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.groups_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.groups_id_seq OWNER TO optisam;

--
-- Name: groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.groups_id_seq OWNED BY public.groups.id;


--
-- Name: roles; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.roles (
    user_role character varying NOT NULL
);


ALTER TABLE public.roles OWNER TO optisam;

--
-- Name: scopes; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.scopes (
    scope_code character varying NOT NULL,
    scope_name character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now(),
    created_by character varying,
    scope_type public.scope_types DEFAULT 'GENERIC'::public.scope_types NOT NULL
);


ALTER TABLE public.scopes OWNER TO optisam;

--
-- Name: users; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.users (
    username character varying NOT NULL,
    first_name character varying NOT NULL,
    last_name character varying,
    role character varying,
    password character varying NOT NULL,
    locale character varying NOT NULL,
    cont_failed_login smallint DEFAULT 0 NOT NULL,
    created_on timestamp without time zone DEFAULT now(),
    last_login timestamp without time zone,
    first_login boolean DEFAULT false,
    profile_pic bytea
);


ALTER TABLE public.users OWNER TO optisam;

--
-- Name: users_audit; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.users_audit (
    id integer NOT NULL,
    username character varying NOT NULL,
    first_name character varying NOT NULL,
    last_name character varying NOT NULL,
    role character varying NOT NULL,
    locale character varying NOT NULL,
    cont_failed_login smallint DEFAULT 0 NOT NULL,
    created_on timestamp without time zone NOT NULL,
    last_login timestamp without time zone,
    operation public.audit_status,
    updated_by character varying NOT NULL,
    updated_on timestamp without time zone DEFAULT now()
);


ALTER TABLE public.users_audit OWNER TO optisam;

--
-- Name: users_audit_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.users_audit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.users_audit_id_seq OWNER TO optisam;

--
-- Name: users_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.users_audit_id_seq OWNED BY public.users_audit.id;


--
-- Name: groups id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.groups ALTER COLUMN id SET DEFAULT nextval('public.groups_id_seq'::regclass);


--
-- Name: users_audit id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.users_audit ALTER COLUMN id SET DEFAULT nextval('public.users_audit_id_seq'::regclass);


--
-- Name: group_ownership group_ownership_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.group_ownership
    ADD CONSTRAINT group_ownership_pkey PRIMARY KEY (group_id, user_id);


--
-- Name: groups groups_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_pkey PRIMARY KEY (id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (user_role);


--
-- Name: scopes scopes_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.scopes
    ADD CONSTRAINT scopes_pkey PRIMARY KEY (scope_code);


--
-- Name: users_audit users_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.users_audit
    ADD CONSTRAINT users_audit_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (username);


--
-- Name: fully_qualified_name_gist_idx; Type: INDEX; Schema: public; Owner: optisam
--

CREATE INDEX fully_qualified_name_gist_idx ON public.groups USING gist (fully_qualified_name);


--
-- Name: group_ownership group_ownership_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.group_ownership
    ADD CONSTRAINT group_ownership_group_id_fkey FOREIGN KEY ("group_id") REFERENCES public.groups(id) ON DELETE CASCADE;


--
-- Name: group_ownership group_ownership_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.group_ownership
    ADD CONSTRAINT group_ownership_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(username) ON DELETE CASCADE;


--
-- Name: groups groups_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(username);


--
-- Name: groups groups_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.groups
    ADD CONSTRAINT groups_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.groups(id) ON DELETE CASCADE;


--
-- Name: scopes scopes_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.scopes
    ADD CONSTRAINT scopes_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.users(username);


--
-- Name: users_audit users_audit_role_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.users_audit
    ADD CONSTRAINT users_audit_role_fkey FOREIGN KEY (role) REFERENCES public.roles(user_role);


--
-- Name: users users_role_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_role_fkey FOREIGN KEY (role) REFERENCES public.roles(user_role);


/**********
--
-- Application Database 
-- 
*********/

--
-- Name: job_status; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.job_status AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'RETRY',
    'RUNNING'
);


ALTER TYPE public.job_status OWNER TO optisam;


--
-- Name: applications; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.applications (
    application_id character varying NOT NULL,
    application_name character varying NOT NULL,
    application_version character varying NOT NULL,
    application_owner character varying NOT NULL,
    application_domain character varying NOT NULL,
    scope character varying NOT NULL,
    obsolescence_risk character varying,
    created_on timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.applications OWNER TO optisam;

--
-- Name: applications_instances; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.applications_instances (
    application_id character varying NOT NULL,
    instance_id character varying NOT NULL,
    instance_environment character varying NOT NULL,
    products text[],
    equipments text[],
    scope character varying NOT NULL
);


ALTER TABLE public.applications_instances OWNER TO optisam;

--
-- Name: domain_criticity; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.domain_criticity (
    critic_id integer NOT NULL,
    scope character varying NOT NULL,
    domain_critic_id integer NOT NULL,
    domains text[] NOT NULL,
    created_by character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now()
);


ALTER TABLE public.domain_criticity OWNER TO optisam;

--
-- Name: domain_criticity_critic_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.domain_criticity_critic_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.domain_criticity_critic_id_seq OWNER TO optisam;

--
-- Name: domain_criticity_critic_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.domain_criticity_critic_id_seq OWNED BY public.domain_criticity.critic_id;


--
-- Name: domain_criticity_meta; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.domain_criticity_meta (
    domain_critic_id integer NOT NULL,
    domain_critic_name character varying NOT NULL
);


ALTER TABLE public.domain_criticity_meta OWNER TO optisam;

--
-- Name: domain_criticity_meta_domain_critic_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.domain_criticity_meta_domain_critic_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.domain_criticity_meta_domain_critic_id_seq OWNER TO optisam;

--
-- Name: domain_criticity_meta_domain_critic_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.domain_criticity_meta_domain_critic_id_seq OWNED BY public.domain_criticity_meta.domain_critic_id;


--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


ALTER TABLE public.gorp_migrations OWNER TO optisam;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.jobs (
    job_id integer NOT NULL,
    type character varying NOT NULL,
    status public.job_status DEFAULT 'PENDING'::public.job_status NOT NULL,
    data jsonb NOT NULL,
    comments character varying,
    start_time timestamp without time zone,
    end_time timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    retry_count integer DEFAULT 0,
    meta_data jsonb NOT NULL
);


ALTER TABLE public.jobs OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.jobs_job_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.jobs_job_id_seq OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.jobs_job_id_seq OWNED BY public.jobs.job_id;


--
-- Name: maintenance_level_meta; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.maintenance_level_meta (
    maintenance_level_id integer NOT NULL,
    maintenance_level_name character varying NOT NULL
);


ALTER TABLE public.maintenance_level_meta OWNER TO optisam;

--
-- Name: maintenance_level_meta_maintenance_level_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.maintenance_level_meta_maintenance_level_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.maintenance_level_meta_maintenance_level_id_seq OWNER TO optisam;

--
-- Name: maintenance_level_meta_maintenance_level_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.maintenance_level_meta_maintenance_level_id_seq OWNED BY public.maintenance_level_meta.maintenance_level_id;


--
-- Name: maintenance_time_criticity; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.maintenance_time_criticity (
    maintenance_critic_id integer NOT NULL,
    scope character varying NOT NULL,
    level_id integer NOT NULL,
    start_month integer NOT NULL,
    end_month integer NOT NULL,
    created_by character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now()
);


ALTER TABLE public.maintenance_time_criticity OWNER TO optisam;

--
-- Name: maintenance_time_criticity_maintenance_critic_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.maintenance_time_criticity_maintenance_critic_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.maintenance_time_criticity_maintenance_critic_id_seq OWNER TO optisam;

--
-- Name: maintenance_time_criticity_maintenance_critic_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.maintenance_time_criticity_maintenance_critic_id_seq OWNED BY public.maintenance_time_criticity.maintenance_critic_id;


--
-- Name: risk_matrix; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.risk_matrix (
    configuration_id integer NOT NULL,
    scope character varying NOT NULL,
    created_by character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.risk_matrix OWNER TO optisam;

--
-- Name: risk_matrix_config; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.risk_matrix_config (
    configuration_id integer NOT NULL,
    domain_critic_id integer NOT NULL,
    maintenance_level_id integer NOT NULL,
    risk_id integer NOT NULL
);


ALTER TABLE public.risk_matrix_config OWNER TO optisam;

--
-- Name: risk_matrix_configuration_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.risk_matrix_configuration_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.risk_matrix_configuration_id_seq OWNER TO optisam;

--
-- Name: risk_matrix_configuration_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.risk_matrix_configuration_id_seq OWNED BY public.risk_matrix.configuration_id;


--
-- Name: risk_meta; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.risk_meta (
    risk_id integer NOT NULL,
    risk_name character varying NOT NULL
);


ALTER TABLE public.risk_meta OWNER TO optisam;

--
-- Name: risk_meta_risk_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.risk_meta_risk_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.risk_meta_risk_id_seq OWNER TO optisam;

--
-- Name: risk_meta_risk_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.risk_meta_risk_id_seq OWNED BY public.risk_meta.risk_id;


--
-- Name: domain_criticity critic_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.domain_criticity ALTER COLUMN critic_id SET DEFAULT nextval('public.domain_criticity_critic_id_seq'::regclass);


--
-- Name: domain_criticity_meta domain_critic_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.domain_criticity_meta ALTER COLUMN domain_critic_id SET DEFAULT nextval('public.domain_criticity_meta_domain_critic_id_seq'::regclass);


--
-- Name: jobs job_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs ALTER COLUMN job_id SET DEFAULT nextval('public.jobs_job_id_seq'::regclass);


--
-- Name: maintenance_level_meta maintenance_level_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.maintenance_level_meta ALTER COLUMN maintenance_level_id SET DEFAULT nextval('public.maintenance_level_meta_maintenance_level_id_seq'::regclass);


--
-- Name: maintenance_time_criticity maintenance_critic_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.maintenance_time_criticity ALTER COLUMN maintenance_critic_id SET DEFAULT nextval('public.maintenance_time_criticity_maintenance_critic_id_seq'::regclass);


--
-- Name: risk_matrix configuration_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix ALTER COLUMN configuration_id SET DEFAULT nextval('public.risk_matrix_configuration_id_seq'::regclass);


--
-- Name: risk_meta risk_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_meta ALTER COLUMN risk_id SET DEFAULT nextval('public.risk_meta_risk_id_seq'::regclass);


--
-- Name: applications_instances applications_instances_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.applications_instances
    ADD CONSTRAINT applications_instances_pkey PRIMARY KEY (instance_id, scope);


--
-- Name: applications applications_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.applications
    ADD CONSTRAINT applications_pkey PRIMARY KEY (application_id, scope);


--
-- Name: domain_criticity_meta domain_criticity_meta_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.domain_criticity_meta
    ADD CONSTRAINT domain_criticity_meta_pkey PRIMARY KEY (domain_critic_id);


--
-- Name: domain_criticity domain_criticity_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.domain_criticity
    ADD CONSTRAINT domain_criticity_pkey PRIMARY KEY (critic_id);


--
-- Name: domain_criticity domain_criticity_scope_domain_critic_id_key; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.domain_criticity
    ADD CONSTRAINT domain_criticity_scope_domain_critic_id_key UNIQUE (scope, domain_critic_id);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (job_id);


--
-- Name: maintenance_level_meta maintenance_level_meta_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.maintenance_level_meta
    ADD CONSTRAINT maintenance_level_meta_pkey PRIMARY KEY (maintenance_level_id);


--
-- Name: maintenance_time_criticity maintenance_time_criticity_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.maintenance_time_criticity
    ADD CONSTRAINT maintenance_time_criticity_pkey PRIMARY KEY (maintenance_critic_id);


--
-- Name: maintenance_time_criticity maintenance_time_criticity_scope_level_id_key; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.maintenance_time_criticity
    ADD CONSTRAINT maintenance_time_criticity_scope_level_id_key UNIQUE (scope, level_id);


--
-- Name: risk_matrix_config risk_matrix_config_configuration_id_domain_critic_id_mainte_key; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix_config
    ADD CONSTRAINT risk_matrix_config_configuration_id_domain_critic_id_mainte_key UNIQUE (configuration_id, domain_critic_id, maintenance_level_id);


--
-- Name: risk_matrix risk_matrix_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix
    ADD CONSTRAINT risk_matrix_pkey PRIMARY KEY (configuration_id);


--
-- Name: risk_matrix risk_matrix_scope_key; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix
    ADD CONSTRAINT risk_matrix_scope_key UNIQUE (scope);


--
-- Name: risk_meta risk_meta_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_meta
    ADD CONSTRAINT risk_meta_pkey PRIMARY KEY (risk_id);


--
-- Name: scope_index; Type: INDEX; Schema: public; Owner: optisam
--

CREATE INDEX scope_index ON public.applications USING btree (scope);


--
-- Name: domain_criticity domain_criticity_domain_critic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.domain_criticity
    ADD CONSTRAINT domain_criticity_domain_critic_id_fkey FOREIGN KEY (domain_critic_id) REFERENCES public.domain_criticity_meta(domain_critic_id);


--
-- Name: maintenance_time_criticity maintenance_time_criticity_level_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.maintenance_time_criticity
    ADD CONSTRAINT maintenance_time_criticity_level_id_fkey FOREIGN KEY (level_id) REFERENCES public.maintenance_level_meta(maintenance_level_id);


--
-- Name: risk_matrix_config risk_matrix_config_configuration_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix_config
    ADD CONSTRAINT risk_matrix_config_configuration_id_fkey FOREIGN KEY (configuration_id) REFERENCES public.risk_matrix(configuration_id) ON DELETE CASCADE;


--
-- Name: risk_matrix_config risk_matrix_config_domain_critic_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix_config
    ADD CONSTRAINT risk_matrix_config_domain_critic_id_fkey FOREIGN KEY (domain_critic_id) REFERENCES public.domain_criticity_meta(domain_critic_id);


--
-- Name: risk_matrix_config risk_matrix_config_maintenance_level_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix_config
    ADD CONSTRAINT risk_matrix_config_maintenance_level_id_fkey FOREIGN KEY (maintenance_level_id) REFERENCES public.maintenance_level_meta(maintenance_level_id);


--
-- Name: risk_matrix_config risk_matrix_config_risk_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.risk_matrix_config
    ADD CONSTRAINT risk_matrix_config_risk_id_fkey FOREIGN KEY (risk_id) REFERENCES public.risk_meta(risk_id);


/**********
--
-- DPS Database 
-- 
*********/


--
-- Name: data_type; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.data_type AS ENUM (
    'DATA',
    'METADATA',
    'GLOBALDATA'
);


ALTER TYPE public.data_type OWNER TO optisam;

--
-- Name: deletion_type; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.deletion_type AS ENUM (
    'ACQRIGHTS',
    'INVENTORY_PARK',
    'WHOLE_INVENTORY'
);


ALTER TYPE public.deletion_type OWNER TO optisam;

--
-- Name: job_status; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.job_status AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'RETRY',
    'RUNNING'
);


ALTER TYPE public.job_status OWNER TO optisam;

--
-- Name: scope_types; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.scope_types AS ENUM (
    'GENERIC',
    'SPECIFIC'
);


ALTER TYPE public.scope_types OWNER TO optisam;

--
-- Name: upload_status; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.upload_status AS ENUM (
    'COMPLETED',
    'FAILED',
    'INPROGRESS',
    'PARTIAL',
    'PENDING',
    'PROCESSED',
    'SUCCESS',
    'UPLOADED'
);


ALTER TYPE public.upload_status OWNER TO optisam;

--
-- Name: upload_status_old; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.upload_status_old AS ENUM (
    'PENDING',
    'SUCCESS',
    'FAILED',
    'INPROGRESS',
    'PARTIAL'
);


ALTER TYPE public.upload_status_old OWNER TO optisam;


--
-- Name: core_factor_logs; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.core_factor_logs (
    upload_id integer NOT NULL,
    file_name character varying DEFAULT ''::character varying NOT NULL,
    uploaded_on timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.core_factor_logs OWNER TO optisam;

--
-- Name: core_factor_logs_upload_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.core_factor_logs_upload_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.core_factor_logs_upload_id_seq OWNER TO optisam;

--
-- Name: core_factor_logs_upload_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.core_factor_logs_upload_id_seq OWNED BY public.core_factor_logs.upload_id;


--
-- Name: core_factor_references; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.core_factor_references (
    id integer NOT NULL,
    manufacturer character varying DEFAULT ''::character varying NOT NULL,
    model character varying DEFAULT ''::character varying NOT NULL,
    core_factor character varying DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.core_factor_references OWNER TO optisam;

--
-- Name: deletion_audit; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.deletion_audit (
    id integer NOT NULL,
    scope character varying NOT NULL,
    deletion_type public.deletion_type NOT NULL,
    status public.upload_status DEFAULT 'INPROGRESS'::public.upload_status NOT NULL,
    reason character varying DEFAULT ''::character varying,
    created_by character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now() NOT NULL,
    updated_on timestamp without time zone
);


ALTER TABLE public.deletion_audit OWNER TO optisam;

--
-- Name: deletion_audit_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.deletion_audit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.deletion_audit_id_seq OWNER TO optisam;

--
-- Name: deletion_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.deletion_audit_id_seq OWNED BY public.deletion_audit.id;


--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


ALTER TABLE public.gorp_migrations OWNER TO optisam;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.jobs (
    job_id integer NOT NULL,
    type character varying NOT NULL,
    status public.job_status DEFAULT 'PENDING'::public.job_status NOT NULL,
    data jsonb NOT NULL,
    comments character varying,
    start_time timestamp without time zone,
    end_time timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    retry_count integer DEFAULT 0,
    meta_data jsonb NOT NULL
);


ALTER TABLE public.jobs OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.jobs_job_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.jobs_job_id_seq OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.jobs_job_id_seq OWNED BY public.jobs.job_id;


--
-- Name: uploaded_data_files; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.uploaded_data_files (
    upload_id integer NOT NULL,
    scope character varying NOT NULL,
    data_type public.data_type,
    file_name character varying NOT NULL,
    status public.upload_status DEFAULT 'PENDING'::public.upload_status NOT NULL,
    uploaded_by character varying NOT NULL,
    uploaded_on timestamp without time zone DEFAULT now() NOT NULL,
    total_records integer DEFAULT 0 NOT NULL,
    success_records integer DEFAULT 0 NOT NULL,
    failed_records integer DEFAULT 0 NOT NULL,
    comments character varying DEFAULT ''::character varying,
    updated_on timestamp without time zone,
    gid integer DEFAULT 0 NOT NULL,
    scope_type public.scope_types DEFAULT 'GENERIC'::public.scope_types,
    analysis_id character varying
);


ALTER TABLE public.uploaded_data_files OWNER TO optisam;

--
-- Name: uploaded_data_files_upload_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.uploaded_data_files_upload_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.uploaded_data_files_upload_id_seq OWNER TO optisam;

--
-- Name: uploaded_data_files_upload_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.uploaded_data_files_upload_id_seq OWNED BY public.uploaded_data_files.upload_id;


--
-- Name: core_factor_logs upload_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.core_factor_logs ALTER COLUMN upload_id SET DEFAULT nextval('public.core_factor_logs_upload_id_seq'::regclass);


--
-- Name: deletion_audit id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.deletion_audit ALTER COLUMN id SET DEFAULT nextval('public.deletion_audit_id_seq'::regclass);


--
-- Name: jobs job_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs ALTER COLUMN job_id SET DEFAULT nextval('public.jobs_job_id_seq'::regclass);


--
-- Name: uploaded_data_files upload_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.uploaded_data_files ALTER COLUMN upload_id SET DEFAULT nextval('public.uploaded_data_files_upload_id_seq'::regclass);


--
-- Name: core_factor_logs core_factor_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.core_factor_logs
    ADD CONSTRAINT core_factor_logs_pkey PRIMARY KEY (upload_id);


--
-- Name: core_factor_references core_factor_references_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.core_factor_references
    ADD CONSTRAINT core_factor_references_pkey PRIMARY KEY (id);


--
-- Name: deletion_audit deletion_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.deletion_audit
    ADD CONSTRAINT deletion_audit_pkey PRIMARY KEY (id, scope, deletion_type);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (job_id);


--
-- Name: uploaded_data_files uploaded_data_files_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.uploaded_data_files
    ADD CONSTRAINT uploaded_data_files_pkey PRIMARY KEY (upload_id, file_name);


/**********
--
-- Product Database 
-- 
*********/


--
-- Name: job_status; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.job_status AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'RETRY',
    'RUNNING'
);


ALTER TYPE public.job_status OWNER TO optisam;


--
-- Name: acqrights; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.acqrights (
    sku character varying NOT NULL,
    swidtag character varying NOT NULL,
    product_name character varying NOT NULL,
    product_editor character varying NOT NULL,
    scope character varying NOT NULL,
    metric character varying NOT NULL,
    num_licenses_acquired integer DEFAULT 0 NOT NULL,
    num_licences_computed integer DEFAULT 0 NOT NULL,
    num_licences_maintainance integer DEFAULT 0 NOT NULL,
    avg_unit_price numeric(15,2) DEFAULT 0 NOT NULL,
    avg_maintenance_unit_price numeric(15,2) DEFAULT 0 NOT NULL,
    total_purchase_cost numeric(15,2) DEFAULT 0 NOT NULL,
    total_computed_cost numeric(15,2) DEFAULT 0 NOT NULL,
    total_maintenance_cost numeric(15,2) DEFAULT 0 NOT NULL,
    total_cost numeric(15,2) DEFAULT 0 NOT NULL,
    created_on timestamp without time zone DEFAULT now() NOT NULL,
    created_by character varying NOT NULL,
    updated_on timestamp without time zone DEFAULT now() NOT NULL,
    updated_by character varying,
    start_of_maintenance timestamp without time zone,
    end_of_maintenance timestamp without time zone,
    version character varying NOT NULL,
    comment character varying DEFAULT ''::character varying,
    last_purchased_order character varying DEFAULT ''::character varying NOT NULL,
    support_number character varying DEFAULT ''::character varying NOT NULL,
    maintenance_provider character varying DEFAULT ''::character varying NOT NULL,
    ordering_date timestamp without time zone,
    corporate_sourcing_contract character varying DEFAULT ''::character varying NOT NULL,
    software_provider character varying DEFAULT ''::character varying NOT NULL,
    file_name character varying DEFAULT ''::character varying NOT NULL,
    file_data bytea,
    repartition boolean DEFAULT false NOT NULL
);


ALTER TABLE public.acqrights OWNER TO optisam;

--
-- Name: aggregated_rights; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.aggregated_rights (
    sku character varying NOT NULL,
    aggregation_id integer NOT NULL,
    metric character varying NOT NULL,
    ordering_date timestamp without time zone,
    corporate_sourcing_contract character varying DEFAULT ''::character varying NOT NULL,
    software_provider character varying DEFAULT ''::character varying NOT NULL,
    scope character varying NOT NULL,
    num_licenses_acquired integer DEFAULT 0 NOT NULL,
    num_licences_computed integer DEFAULT 0 NOT NULL,
    num_licences_maintenance integer DEFAULT 0 NOT NULL,
    avg_unit_price numeric(15,2) DEFAULT 0 NOT NULL,
    avg_maintenance_unit_price numeric(15,2) DEFAULT 0 NOT NULL,
    total_purchase_cost numeric(15,2) DEFAULT 0 NOT NULL,
    total_computed_cost numeric(15,2) DEFAULT 0 NOT NULL,
    total_maintenance_cost numeric(15,2) DEFAULT 0 NOT NULL,
    total_cost numeric(15,2) DEFAULT 0 NOT NULL,
    start_of_maintenance timestamp without time zone,
    end_of_maintenance timestamp without time zone,
    last_purchased_order character varying DEFAULT ''::character varying NOT NULL,
    support_number character varying DEFAULT ''::character varying NOT NULL,
    maintenance_provider character varying DEFAULT ''::character varying NOT NULL,
    comment character varying DEFAULT ''::character varying,
    created_on timestamp without time zone DEFAULT now() NOT NULL,
    created_by character varying NOT NULL,
    updated_on timestamp without time zone DEFAULT now() NOT NULL,
    updated_by character varying,
    file_name character varying DEFAULT ''::character varying NOT NULL,
    file_data bytea,
    repartition boolean DEFAULT false NOT NULL
);


ALTER TABLE public.aggregated_rights OWNER TO optisam;

--
-- Name: aggregations; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.aggregations (
    id integer NOT NULL,
    aggregation_name character varying NOT NULL,
    scope character varying NOT NULL,
    product_editor character varying NOT NULL,
    products text[] NOT NULL,
    swidtags text[] NOT NULL,
    created_on timestamp without time zone DEFAULT now() NOT NULL,
    created_by character varying NOT NULL,
    updated_on timestamp without time zone,
    updated_by character varying
);


ALTER TABLE public.aggregations OWNER TO optisam;

--
-- Name: aggregations_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.aggregations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.aggregations_id_seq OWNER TO optisam;

--
-- Name: aggregations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.aggregations_id_seq OWNED BY public.aggregations.id;


--
-- Name: dashboard_audit; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.dashboard_audit (
    id integer NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('UTC'::text, now()) NOT NULL,
    next_update_at timestamp with time zone,
    updated_by character varying NOT NULL,
    scope character varying NOT NULL
);


ALTER TABLE public.dashboard_audit OWNER TO optisam;

--
-- Name: dashboard_audit_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.dashboard_audit_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.dashboard_audit_id_seq OWNER TO optisam;

--
-- Name: dashboard_audit_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.dashboard_audit_id_seq OWNED BY public.dashboard_audit.id;


--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


ALTER TABLE public.gorp_migrations OWNER TO optisam;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.jobs (
    job_id integer NOT NULL,
    type character varying NOT NULL,
    status public.job_status DEFAULT 'PENDING'::public.job_status NOT NULL,
    data jsonb NOT NULL,
    comments character varying,
    start_time timestamp without time zone,
    end_time timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    retry_count integer DEFAULT 0,
    meta_data jsonb NOT NULL
);


ALTER TABLE public.jobs OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.jobs_job_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.jobs_job_id_seq OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.jobs_job_id_seq OWNED BY public.jobs.job_id;


--
-- Name: overall_computed_licences; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.overall_computed_licences (
    sku character varying NOT NULL,
    swidtags character varying NOT NULL,
    scope character varying NOT NULL,
    product_names character varying NOT NULL,
    aggregation_name character varying DEFAULT ''::character varying NOT NULL,
    metrics character varying NOT NULL,
    num_computed_licences integer DEFAULT 0 NOT NULL,
    num_acquired_licences integer DEFAULT 0 NOT NULL,
    total_cost numeric(15,2) DEFAULT 0 NOT NULL,
    purchase_cost numeric(15,2) DEFAULT 0 NOT NULL,
    computed_cost numeric(15,2) DEFAULT 0 NOT NULL,
    delta_number integer DEFAULT 0 NOT NULL,
    delta_cost numeric(15,2) DEFAULT 0 NOT NULL,
    avg_unit_price numeric(15,2) DEFAULT 0 NOT NULL,
    computed_details character varying NOT NULL,
    metic_not_defined boolean DEFAULT false,
    not_deployed boolean DEFAULT false,
    editor character varying NOT NULL
);


ALTER TABLE public.overall_computed_licences OWNER TO optisam;

--
-- Name: products; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.products (
    swidtag character varying NOT NULL,
    product_name character varying DEFAULT ''::character varying NOT NULL,
    product_version character varying DEFAULT ''::character varying NOT NULL,
    product_edition character varying DEFAULT ''::character varying NOT NULL,
    product_category character varying DEFAULT ''::character varying NOT NULL,
    product_editor character varying DEFAULT ''::character varying NOT NULL,
    scope character varying NOT NULL,
    option_of character varying DEFAULT ''::character varying NOT NULL,
    aggregation_id integer DEFAULT 0 NOT NULL,
    aggregation_name character varying DEFAULT ''::character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now() NOT NULL,
    created_by character varying NOT NULL,
    updated_on timestamp without time zone DEFAULT now() NOT NULL,
    updated_by character varying
);


ALTER TABLE public.products OWNER TO optisam;

--
-- Name: products_applications; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.products_applications (
    swidtag character varying NOT NULL,
    application_id character varying NOT NULL,
    scope character varying NOT NULL
);


ALTER TABLE public.products_applications OWNER TO optisam;

--
-- Name: products_equipments; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.products_equipments (
    swidtag character varying NOT NULL,
    equipment_id character varying NOT NULL,
    num_of_users integer,
    scope character varying NOT NULL,
    allocated_metric character varying DEFAULT ''::character varying NOT NULL
);


ALTER TABLE public.products_equipments OWNER TO optisam;

--
-- Name: aggregations id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.aggregations ALTER COLUMN id SET DEFAULT nextval('public.aggregations_id_seq'::regclass);


--
-- Name: dashboard_audit id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.dashboard_audit ALTER COLUMN id SET DEFAULT nextval('public.dashboard_audit_id_seq'::regclass);


--
-- Name: jobs job_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs ALTER COLUMN job_id SET DEFAULT nextval('public.jobs_job_id_seq'::regclass);


--
-- Name: acqrights acqrights_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.acqrights
    ADD CONSTRAINT acqrights_pkey PRIMARY KEY (sku, scope);


--
-- Name: aggregated_rights aggregated_rights_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.aggregated_rights
    ADD CONSTRAINT aggregated_rights_pkey PRIMARY KEY (sku, scope);


--
-- Name: aggregations aggregations_aggregation_name_scope_key; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.aggregations
    ADD CONSTRAINT aggregations_aggregation_name_scope_key UNIQUE (aggregation_name, scope);


--
-- Name: aggregations aggregations_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.aggregations
    ADD CONSTRAINT aggregations_pkey PRIMARY KEY (id);


--
-- Name: dashboard_audit dashboard_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.dashboard_audit
    ADD CONSTRAINT dashboard_audit_pkey PRIMARY KEY (scope);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (job_id);


--
-- Name: overall_computed_licences overall_computed_licences_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.overall_computed_licences
    ADD CONSTRAINT overall_computed_licences_pkey PRIMARY KEY (sku, swidtags, scope);


--
-- Name: products_applications products_applications_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.products_applications
    ADD CONSTRAINT products_applications_pkey PRIMARY KEY (swidtag, application_id, scope);


--
-- Name: products_equipments products_equipments_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.products_equipments
    ADD CONSTRAINT products_equipments_pkey PRIMARY KEY (swidtag, equipment_id, scope);


--
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (swidtag, scope);


--
-- Name: aggregated_rights aggregated_rights_aggregation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.aggregated_rights
    ADD CONSTRAINT aggregated_rights_aggregation_id_fkey FOREIGN KEY (aggregation_id) REFERENCES public.aggregations(id) ON DELETE CASCADE;


--
-- Name: products_applications products_applications_swidtag_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.products_applications
    ADD CONSTRAINT products_applications_swidtag_fkey FOREIGN KEY (swidtag, scope) REFERENCES public.products(swidtag, scope) ON DELETE CASCADE;


--
-- Name: products_equipments products_equipments_swidtag_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.products_equipments
    ADD CONSTRAINT products_equipments_swidtag_fkey FOREIGN KEY (swidtag, scope) REFERENCES public.products(swidtag, scope) ON DELETE CASCADE;

/**********
--
-- Report Database 
-- 
*********/

--
-- Name: job_status; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.job_status AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'RETRY',
    'RUNNING'
);


ALTER TYPE public.job_status OWNER TO optisam;

--
-- Name: report_status; Type: TYPE; Schema: public; Owner: optisam
--

CREATE TYPE public.report_status AS ENUM (
    'PENDING',
    'COMPLETED',
    'FAILED',
    'RUNNING'
);


ALTER TYPE public.report_status OWNER TO optisam;


--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


ALTER TABLE public.gorp_migrations OWNER TO optisam;

--
-- Name: jobs; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.jobs (
    job_id integer NOT NULL,
    type character varying NOT NULL,
    status public.job_status DEFAULT 'PENDING'::public.job_status NOT NULL,
    data jsonb NOT NULL,
    comments character varying,
    start_time timestamp without time zone,
    end_time timestamp without time zone,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    retry_count integer DEFAULT 0,
    meta_data jsonb
);


ALTER TABLE public.jobs OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.jobs_job_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.jobs_job_id_seq OWNER TO optisam;

--
-- Name: jobs_job_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.jobs_job_id_seq OWNED BY public.jobs.job_id;


--
-- Name: report; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.report (
    report_id integer NOT NULL,
    report_type_id integer NOT NULL,
    scope character varying NOT NULL,
    report_metadata jsonb NOT NULL,
    report_data json,
    report_status public.report_status DEFAULT 'PENDING'::public.report_status NOT NULL,
    created_by character varying NOT NULL,
    created_on timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.report OWNER TO optisam;

--
-- Name: report_report_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.report_report_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.report_report_id_seq OWNER TO optisam;

--
-- Name: report_report_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.report_report_id_seq OWNED BY public.report.report_id;


--
-- Name: report_type; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.report_type (
    report_type_id integer NOT NULL,
    report_type_name character varying NOT NULL
);


ALTER TABLE public.report_type OWNER TO optisam;

--
-- Name: report_type_report_type_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.report_type_report_type_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.report_type_report_type_id_seq OWNER TO optisam;

--
-- Name: report_type_report_type_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.report_type_report_type_id_seq OWNED BY public.report_type.report_type_id;


--
-- Name: jobs job_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs ALTER COLUMN job_id SET DEFAULT nextval('public.jobs_job_id_seq'::regclass);


--
-- Name: report report_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.report ALTER COLUMN report_id SET DEFAULT nextval('public.report_report_id_seq'::regclass);


--
-- Name: report_type report_type_id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.report_type ALTER COLUMN report_type_id SET DEFAULT nextval('public.report_type_report_type_id_seq'::regclass);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (job_id);


--
-- Name: report report_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT report_pkey PRIMARY KEY (report_id);


--
-- Name: report_type report_type_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.report_type
    ADD CONSTRAINT report_type_pkey PRIMARY KEY (report_type_id);


--
-- Name: report report_report_type_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT report_report_type_id_fkey FOREIGN KEY (report_type_id) REFERENCES public.report_type(report_type_id);


/**********
--
-- Simulation Database 
-- 
*********/

--
-- Name: config_data; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.config_data (
    metadata_id integer NOT NULL,
    attribute_value character varying NOT NULL,
    json_data jsonb NOT NULL
);


ALTER TABLE public.config_data OWNER TO optisam;

--
-- Name: config_master; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.config_master (
    id integer NOT NULL,
    name character varying NOT NULL,
    equipment_type character varying NOT NULL,
    status integer NOT NULL,
    created_by character varying NOT NULL,
    created_on timestamp without time zone NOT NULL,
    updated_by character varying NOT NULL,
    updated_on timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.config_master OWNER TO optisam;

--
-- Name: config_master_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.config_master_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.config_master_id_seq OWNER TO optisam;

--
-- Name: config_master_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.config_master_id_seq OWNED BY public.config_master.id;


--
-- Name: config_metadata; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.config_metadata (
    id integer NOT NULL,
    config_id integer NOT NULL,
    equipment_type character varying NOT NULL,
    attribute_name character varying NOT NULL,
    config_filename character varying NOT NULL
);


ALTER TABLE public.config_metadata OWNER TO optisam;

--
-- Name: config_metadata_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.config_metadata_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.config_metadata_id_seq OWNER TO optisam;

--
-- Name: config_metadata_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.config_metadata_id_seq OWNED BY public.config_metadata.id;


--
-- Name: gorp_migrations; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.gorp_migrations (
    id text NOT NULL,
    applied_at timestamp with time zone
);


ALTER TABLE public.gorp_migrations OWNER TO optisam;

--
-- Name: status; Type: TABLE; Schema: public; Owner: optisam
--

CREATE TABLE public.status (
    id integer NOT NULL,
    text character varying
);


ALTER TABLE public.status OWNER TO optisam;

--
-- Name: status_id_seq; Type: SEQUENCE; Schema: public; Owner: optisam
--

CREATE SEQUENCE public.status_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.status_id_seq OWNER TO optisam;

--
-- Name: status_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: optisam
--

ALTER SEQUENCE public.status_id_seq OWNED BY public.status.id;


--
-- Name: config_master id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.config_master ALTER COLUMN id SET DEFAULT nextval('public.config_master_id_seq'::regclass);


--
-- Name: config_metadata id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.config_metadata ALTER COLUMN id SET DEFAULT nextval('public.config_metadata_id_seq'::regclass);


--
-- Name: status id; Type: DEFAULT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.status ALTER COLUMN id SET DEFAULT nextval('public.status_id_seq'::regclass);


--
-- Name: config_master config_master_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.config_master
    ADD CONSTRAINT config_master_pkey PRIMARY KEY (id);


--
-- Name: config_metadata config_metadata_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.config_metadata
    ADD CONSTRAINT config_metadata_pkey PRIMARY KEY (id);


--
-- Name: gorp_migrations gorp_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.gorp_migrations
    ADD CONSTRAINT gorp_migrations_pkey PRIMARY KEY (id);


--
-- Name: status status_pkey; Type: CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.status
    ADD CONSTRAINT status_pkey PRIMARY KEY (id);


--
-- Name: config_data config_data_metadata_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.config_data
    ADD CONSTRAINT config_data_metadata_id_fkey FOREIGN KEY (metadata_id) REFERENCES public.config_metadata(id) ON DELETE CASCADE;


--
-- Name: config_master config_master_status_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.config_master
    ADD CONSTRAINT config_master_status_fkey FOREIGN KEY (status) REFERENCES public.status(id);


--
-- Name: config_metadata config_metadata_config_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: optisam
--

ALTER TABLE ONLY public.config_metadata
    ADD CONSTRAINT config_metadata_config_id_fkey FOREIGN KEY (config_id) REFERENCES public.config_master(id) ON DELETE CASCADE;

