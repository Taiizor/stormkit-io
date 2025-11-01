-- =====================================
-- 0002_alter_domains_add_columns.up.sql
-- =====================================

ALTER TABLE skitapi.domains ADD COLUMN IF NOT EXISTS custom_cert_value TEXT;
ALTER TABLE skitapi.domains ADD COLUMN IF NOT EXISTS custom_cert_key TEXT;

-- ===============================================
-- 0003_alter_deployments_add_status_checks.up.sql
-- ===============================================

ALTER TABLE skitapi.deployments ADD COLUMN IF NOT EXISTS status_checks TEXT;
ALTER TABLE skitapi.deployments ADD COLUMN IF NOT EXISTS status_checks_passed BOOLEAN;
ALTER TABLE skitapi.deployments ADD COLUMN IF NOT EXISTS is_immutable BOOLEAN;

UPDATE skitapi.deployments SET is_immutable = TRUE;

-- =======================================
-- 0004_create_analytics_agg_tables.up.sql
-- =======================================

CREATE TABLE IF NOT EXISTS skitapi.analytics_visitors_agg_200 (
    aggregate_date date NOT NULL,
    domain_id bigint NOT NULL,
    unique_visitors bigint NOT NULL,
    total_visitors bigint NOT NULL,
    PRIMARY KEY (aggregate_date, domain_id)
);

CREATE TABLE IF NOT EXISTS skitapi.analytics_visitors_agg_404 (
    aggregate_date date NOT NULL,
    domain_id bigint NOT NULL,
    unique_visitors bigint NOT NULL,
    total_visitors bigint NOT NULL,
    PRIMARY KEY (aggregate_date, domain_id)
);

CREATE TABLE IF NOT EXISTS skitapi.analytics_visitors_agg_hourly_200 (
    aggregate_date timestamp without time zone NOT NULL,
    domain_id bigint NOT NULL,
    unique_visitors bigint NOT NULL,
    total_visitors bigint NOT NULL,
    PRIMARY KEY (aggregate_date, domain_id)
);

CREATE TABLE IF NOT EXISTS skitapi.analytics_visitors_agg_hourly_404 (
    aggregate_date timestamp without time zone NOT NULL,
    domain_id bigint NOT NULL,
    unique_visitors bigint NOT NULL,
    total_visitors bigint NOT NULL,
    PRIMARY KEY (aggregate_date, domain_id)
);

CREATE TABLE IF NOT EXISTS skitapi.analytics_referrers (
    aggregate_date date NOT NULL,
    referrer text NOT NULL,
    request_path text NOT NULL,
    domain_id bigint NOT NULL,
    visit_count bigint NOT NULL,
    PRIMARY KEY (aggregate_date, referrer, request_path, domain_id)
);

CREATE TABLE IF NOT EXISTS skitapi.analytics_visitors_by_countries (
    aggregate_date date NOT NULL,
    domain_id bigint NOT NULL,
    country_iso_code text NOT NULL,
    visit_count bigint NOT NULL,
    PRIMARY KEY (aggregate_date, country_iso_code, domain_id)
);

DO $$
BEGIN
  BEGIN
    -- We keep these inserts here because the if the foreign keys exist, the migration 
    -- is already completed and it will move to the catch block and skip these inserts.

    -- Insert the initial data into 200 and 304 table

    INSERT INTO skitapi.analytics_visitors_agg_200 (aggregate_date, domain_id, unique_visitors, total_visitors)
        SELECT DATE(a.request_timestamp) as req_date, domain_id, COUNT(DISTINCT a.visitor_ip) AS unique_visitors, COUNT(a.domain_id) as total_visitors
        FROM skitapi.analytics a
        WHERE a.response_code = ANY(array[200, 304])
        GROUP BY req_date, a.domain_id
        ORDER BY req_date DESC
    ON CONFLICT (aggregate_date, domain_id) DO UPDATE SET unique_visitors = EXCLUDED.unique_visitors, total_visitors = EXCLUDED.total_visitors;

    -- Insert the initial data into 404 table

    INSERT INTO skitapi.analytics_visitors_agg_404 (aggregate_date, domain_id, unique_visitors, total_visitors)
        SELECT DATE(a.request_timestamp) as req_date, domain_id, COUNT(DISTINCT a.visitor_ip) AS unique_visitors, COUNT(a.domain_id) as total_visitors
        FROM skitapi.analytics a
        WHERE a.response_code = 404
        GROUP BY req_date, a.domain_id
        ORDER BY req_date DESC
    ON CONFLICT (aggregate_date, domain_id) DO UPDATE SET unique_visitors = EXCLUDED.unique_visitors, total_visitors = EXCLUDED.total_visitors;

    -- Insert the initial data into 200 and 304 hourly table

    INSERT INTO skitapi.analytics_visitors_agg_hourly_200 (aggregate_date, domain_id, unique_visitors, total_visitors)
        SELECT TO_CHAR(DATE_TRUNC('hour', a.request_timestamp), 'YYYY-MM-DD HH24:MI')::timestamp as agg_date, domain_id, COUNT(DISTINCT a.visitor_ip) AS unique_visitors, COUNT(a.domain_id) as total_visitors
        FROM skitapi.analytics a
        WHERE a.response_code = ANY(array[200, 304]) AND a.request_timestamp >= current_date - interval '3 days'
        GROUP BY agg_date, a.domain_id
        ORDER BY agg_date DESC
    ON CONFLICT (aggregate_date, domain_id) DO UPDATE SET unique_visitors = EXCLUDED.unique_visitors, total_visitors = EXCLUDED.total_visitors;

    -- Insert the initial data into 404 hourly table

    INSERT INTO skitapi.analytics_visitors_agg_hourly_404 (aggregate_date, domain_id, unique_visitors, total_visitors)
        SELECT TO_CHAR(DATE_TRUNC('hour', a.request_timestamp), 'YYYY-MM-DD HH24:MI')::timestamp as agg_date, domain_id, COUNT(DISTINCT a.visitor_ip) AS unique_visitors, COUNT(a.domain_id) as total_visitors
        FROM skitapi.analytics a
        WHERE a.response_code = 404 AND a.request_timestamp >= current_date - interval '3 days'
        GROUP BY agg_date, a.domain_id
        ORDER BY agg_date DESC
    ON CONFLICT (aggregate_date, domain_id) DO UPDATE SET unique_visitors = EXCLUDED.unique_visitors, total_visitors = EXCLUDED.total_visitors;

    -- Insert the initial data for referrers table

    INSERT INTO skitapi.analytics_referrers (aggregate_date, referrer, request_path, domain_id, visit_count)
        SELECT DATE(a.request_timestamp) as req_date, regexp_replace(regexp_replace(COALESCE(a.referrer, ''), '^https?://(www\.)?', ''), '/$', '') as referrer_domain, a.request_path, a.domain_id, COUNT(a.domain_id) AS total_count
        FROM skitapi.analytics a
        WHERE a.response_code IN (200, 304) AND a.request_timestamp >= current_date - interval '30 days'
        GROUP BY req_date, referrer_domain, a.request_path, a.domain_id
    ON CONFLICT (aggregate_date, referrer, request_path, domain_id) DO UPDATE SET visit_count = EXCLUDED.visit_count;

    -- Insert the initial data for analytics table

    INSERT INTO skitapi.analytics_visitors_by_countries (aggregate_date, country_iso_code, domain_id, visit_count)
        SELECT DATE(a.request_timestamp) as agg_date, a.country_iso_code, a.domain_id, COUNT(DISTINCT a.visitor_ip) AS count
        FROM skitapi.analytics a
        WHERE a.response_code IN (200, 304) AND a.country_iso_code IS NOT NULL AND a.request_timestamp >= current_date - interval '31 days'
        GROUP BY a.country_iso_code, agg_date, domain_id
    ON CONFLICT (aggregate_date, country_iso_code, domain_id) DO UPDATE SET visit_count = EXCLUDED.visit_count;

  EXCEPTION
    WHEN duplicate_table THEN  -- postgres raises duplicate_table at surprising times. Ex.: for UNIQUE constraints.
    WHEN duplicate_object THEN
      RAISE NOTICE 'Table constraint already exists';
  END;
END $$;

DROP TABLE IF EXISTS skitapi.analytics_archive;

-- ================================
-- 0005_create_volumes_table.up.sql
-- ================================

CREATE TABLE IF NOT EXISTS skitapi.stormkit_config (
    config_id serial primary key NOT NULL,
    volumes_config jsonb NULL,
    updated_at timestamp without time zone NULL,
    created_at timestamp without time zone DEFAULT (now() AT TIME ZONE 'UTC'::text) NOT NULL
);

CREATE TABLE IF NOT EXISTS skitapi.volumes (
    file_id bigserial primary key NOT NULL,
    file_name text NOT NULL,
    file_path text NOT NULL,
    file_size bigint NOT NULL,
    is_public boolean NOT NULL,
    env_id bigint NOT NULL,
    updated_at timestamp without time zone NULL,
    created_at timestamp without time zone DEFAULT (now() AT TIME ZONE 'UTC'::text) NOT NULL
);

DO $$
BEGIN
  BEGIN

    ALTER TABLE ONLY skitapi.volumes
        ADD CONSTRAINT volumes_env_id_fkey FOREIGN KEY (env_id) REFERENCES skitapi.apps_build_conf(env_id) ON DELETE CASCADE;

    ALTER TABLE ONLY skitapi.volumes
        ADD CONSTRAINT volumes_file_name_env_id_key UNIQUE (file_name, env_id);

  EXCEPTION
    WHEN duplicate_table THEN  -- postgres raises duplicate_table at surprising times. Ex.: for UNIQUE constraints.
    WHEN duplicate_object THEN
      RAISE NOTICE 'Table constraint already exists';
  END;
END $$;

-- ===============================================
-- 0006_alter_outbound_webhooks_and_mailer_setup.up.sql
-- ===============================================

ALTER TABLE skitapi.app_outbound_webhooks ALTER COLUMN trigger_when TYPE TEXT USING trigger_when::TEXT;
UPDATE skitapi.app_outbound_webhooks SET trigger_when = 'on_deploy_success' WHERE trigger_when = 'on_deploy';

ALTER TABLE skitapi.apps_build_conf ADD COLUMN IF NOT EXISTS mailer_conf jsonb;

CREATE TABLE IF NOT EXISTS skitapi.mailer (
    email_id bigserial primary key NOT NULL,
    email_to text NOT NULL,
    email_from text NOT NULL,
    email_subject text NOT NULL,
    email_body text NOT NULL,
    env_id bigint NOT NULL,
    created_at timestamp without time zone DEFAULT (now() AT TIME ZONE 'UTC'::text) NOT NULL
);

DO $$
BEGIN
  BEGIN

    ALTER TABLE ONLY skitapi.mailer
        ADD CONSTRAINT mailer_env_id_fkey FOREIGN KEY (env_id) REFERENCES skitapi.apps_build_conf(env_id) ON DELETE CASCADE;

  EXCEPTION
    WHEN duplicate_table THEN  -- postgres raises duplicate_table at surprising times. Ex.: for UNIQUE constraints.
    WHEN duplicate_object THEN
      RAISE NOTICE 'Table constraint already exists';
  END;
END $$;

-- =======================================================================
-- 0007_alter_deployments_add_webhook_event_and_auto_deploy_commits.up.sql
-- =======================================================================

ALTER TABLE skitapi.deployments ADD COLUMN IF NOT EXISTS webhook_event JSONB;
ALTER TABLE skitapi.apps_build_conf ADD COLUMN IF NOT EXISTS auto_deploy_commits TEXT;

-- =====================================================
-- 0008_alter_function_triggers_add_webhook_event.up.sql
-- =====================================================

CREATE TABLE IF NOT EXISTS skitapi.function_trigger_logs (
    ftl_id serial primary key NOT NULL,
    trigger_id bigint NOT NULL,
    request jsonb NOT NULL,  -- includes information such as payload, method, path etc...
    response jsonb NOT NULL, -- includes information such as status, response body etc...
    created_at timestamp without time zone DEFAULT (now() AT TIME ZONE 'UTC'::text) NOT NULL
);

ALTER TABLE skitapi.function_triggers DROP CONSTRAINT IF EXISTS unique_config;
ALTER TABLE skitapi.function_triggers DROP COLUMN IF EXISTS timezone;
ALTER TABLE skitapi.function_triggers DROP COLUMN IF EXISTS app_id;
ALTER TABLE skitapi.function_triggers ALTER COLUMN cron TYPE text;

DO
$$
    BEGIN
        ALTER TABLE skitapi.function_triggers RENAME COLUMN options TO trigger_options;
		ALTER TABLE skitapi.function_triggers RENAME COLUMN status TO trigger_status;
        ALTER TABLE skitapi.function_triggers RENAME COLUMN id TO trigger_id;
    EXCEPTION
        WHEN undefined_column THEN RAISE NOTICE 'column does not exist';
    END;
$$;


DO $$
BEGIN
  BEGIN

    ALTER TABLE ONLY skitapi.function_trigger_logs
        ADD CONSTRAINT function_trigger_logs_trigger_id_fkey FOREIGN KEY (trigger_id) REFERENCES skitapi.function_triggers(trigger_id) ON DELETE CASCADE;

  EXCEPTION
    WHEN duplicate_table THEN  -- postgres raises duplicate_table at surprising times. Ex.: for UNIQUE constraints.
    WHEN duplicate_object THEN
      RAISE NOTICE 'Table constraint already exists';
  END;
END $$;

-- ===============================
-- 0009_domain_ping_updates.up.sql
-- ===============================

ALTER TABLE skitapi.domains ADD COLUMN IF NOT EXISTS last_ping JSONB;
ALTER TABLE skitapi.stormkit_config ADD COLUMN IF NOT EXISTS config_data JSONB;

DO $$
BEGIN
    UPDATE skitapi.stormkit_config
        SET config_data = jsonb_build_object('volumes', volumes_config);
    EXCEPTION
        WHEN undefined_column THEN
            RAISE NOTICE 'Column does not exist: %', SQLERRM;
END $$;

ALTER TABLE skitapi.stormkit_config DROP COLUMN IF EXISTS volumes_config;

-- =============================
-- 0010_volumes_file_type.up.sql
-- =============================

ALTER TABLE skitapi.volumes ADD COLUMN IF NOT EXISTS file_metadata JSONB;

-- =====================================
-- 0011_snippet_unique_constraint.up.sql
-- =====================================

ALTER TABLE skitapi.snippets ADD COLUMN IF NOT EXISTS snippet_content_hash text NULL;
CREATE UNIQUE INDEX IF NOT EXISTS snippets_snippet_content_hash_key ON skitapi.snippets USING btree (env_id, snippet_content_hash);

-- =================================================
-- 0012_remove_unnecessary_columns_and_tables.up.sql
-- =================================================

DO $$
BEGIN
    ALTER TABLE skitapi.user_referrals DROP CONSTRAINT IF EXISTS idx_user_referrals_details;
    ALTER TABLE skitapi.user_referrals DROP CONSTRAINT IF EXISTS idx_user_referrals_invited_by;
    ALTER TABLE skitapi.user_referrals DROP CONSTRAINT IF EXISTS user_referrals_invited_by_users_id_fkey;
    ALTER TABLE skitapi.user_referrals DROP CONSTRAINT IF EXISTS user_referrals_display_name_provider_key;
    ALTER TABLE skitapi.apps_build_conf DROP CONSTRAINT IF EXISTS apps_build_conf_domain_unique_key;
    ALTER TABLE skitapi.apps_build_conf DROP CONSTRAINT IF EXISTS idx_apps_build_conf_domain;

    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table does not exist: %', SQLERRM;
END $$;

DROP TABLE IF EXISTS skitapi.user_referrals;

ALTER TABLE skitapi.deployments DROP COLUMN IF EXISTS is_v2;
ALTER TABLE skitapi.deployments DROP COLUMN IF EXISTS lambda_version;
ALTER TABLE skitapi.deployments DROP COLUMN IF EXISTS percentage;
ALTER TABLE skitapi.deployments DROP COLUMN IF EXISTS is_serverless;
ALTER TABLE skitapi.deployments DROP COLUMN IF EXISTS s3_key_prefix;
ALTER TABLE skitapi.deployments DROP COLUMN IF EXISTS s3_bucket_name;
ALTER TABLE skitapi.deployments DROP COLUMN IF EXISTS stormkit_config;
ALTER TABLE skitapi.apps DROP COLUMN IF EXISTS stormkit_config_hash;
ALTER TABLE skitapi.apps_build_conf DROP COLUMN IF EXISTS domain_name;
ALTER TABLE skitapi.apps_build_conf DROP COLUMN IF EXISTS domain_verified;
ALTER TABLE skitapi.apps_build_conf DROP COLUMN IF EXISTS domain_verified_at;
ALTER TABLE skitapi.apps_build_conf DROP COLUMN IF EXISTS domain_token;
ALTER TABLE skitapi.apps_build_conf DROP COLUMN IF EXISTS snippets;
ALTER TABLE skitapi.apps_build_conf DROP COLUMN IF EXISTS proxy_enabled;

-- ======================
-- 0013_last_login.up.sql
-- ======================

ALTER TABLE skitapi.users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITHOUT TIME ZONE NULL;

-- =====================
-- 0014_auth_wall.up.sql
-- =====================

CREATE TABLE IF NOT EXISTS skitapi.auth_wall (
    login_id bigserial primary key NOT NULL,
    login_email text NOT NULL,
    login_password text NOT NULL,
    last_login_at timestamp without time zone,
    env_id bigint NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS auth_wall_env_id_login_email ON skitapi.auth_wall USING btree (env_id, login_email);

ALTER TABLE skitapi.apps_build_conf ADD COLUMN IF NOT EXISTS auth_wall_conf jsonb;

DO $$
BEGIN
  BEGIN

    ALTER TABLE ONLY skitapi.auth_wall
        ADD CONSTRAINT auth_wall_env_id_fkey FOREIGN KEY (env_id) REFERENCES skitapi.apps_build_conf(env_id) ON UPDATE CASCADE ON DELETE CASCADE;

  EXCEPTION
    WHEN duplicate_table THEN  -- postgres raises duplicate_table at surprising times. Ex.: for UNIQUE constraints.
    WHEN duplicate_object THEN
      RAISE NOTICE 'Table constraint already exists';
  END;
END $$;

DROP INDEX IF EXISTS skitapi.idx_apps_api_key;
DROP INDEX IF EXISTS skitapi.idx_apps_build_conf_api_key;
DROP INDEX IF EXISTS skitapi.idx_apps_api_key_unique;
DROP INDEX IF EXISTS skitapi.idx_apps_build_conf_api_key_unique;

ALTER TABLE skitapi.apps DROP COLUMN IF EXISTS api_key;
ALTER TABLE skitapi.apps_build_conf DROP COLUMN IF EXISTS api_key;