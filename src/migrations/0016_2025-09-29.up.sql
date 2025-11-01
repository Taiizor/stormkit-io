-- Drop feature flags
DROP INDEX IF EXISTS skitapi.feature_flags_ff_name_unique_key;
DROP INDEX IF EXISTS skitapi.idx_feature_flags_app_id;

DO $$ 
BEGIN
	IF EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'skitapi' AND table_name = 'feature_flags') THEN
		ALTER TABLE ONLY skitapi.feature_flags DROP CONSTRAINT IF EXISTS feature_flags_app_id_fkey;
		ALTER TABLE ONLY skitapi.feature_flags DROP CONSTRAINT IF EXISTS feature_flags_env_id_fkey;
	END IF;
END $$;

DROP TABLE IF EXISTS skitapi.feature_flags;

-- Drop stripe client id from users and add metadata column
DROP INDEX IF EXISTS skitapi.idx_users_stripe_client_id;
ALTER TABLE ONLY skitapi.users DROP COLUMN IF EXISTS stripe_client_id;
ALTER TABLE ONLY skitapi.users ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Drop user stats: our business model is no longer based on deployments
DROP INDEX IF EXISTS skitapi.idx_userid_user_stats;
DROP TABLE IF EXISTS skitapi.user_stats;
