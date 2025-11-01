-- ==========================================================
-- create user_id column in teams table
-- rename column to client_package_size in deployments table
-- ==========================================================

ALTER TABLE skitapi.teams ADD COLUMN IF NOT EXISTS user_id BIGINT NULL;

DO $$
BEGIN
  BEGIN

    ALTER TABLE ONLY skitapi.teams
        ADD CONSTRAINT teams_user_id_fkey FOREIGN KEY (user_id) REFERENCES skitapi.users(user_id) ON DELETE CASCADE;

  EXCEPTION
    WHEN duplicate_table THEN  -- postgres raises duplicate_table at surprising times. Ex.: for UNIQUE constraints.
    WHEN duplicate_object THEN
      RAISE NOTICE 'Table constraint already exists';
  END;
END $$;

UPDATE skitapi.teams t SET user_id = (
    SELECT
        user_id
    FROM
        skitapi.team_members tm
    WHERE
        tm.team_id = t.team_id AND
        tm.member_role = 'owner'
);

ALTER TABLE skitapi.teams ALTER COLUMN user_id SET NOT NULL;

DO $$
BEGIN
  BEGIN

    ALTER TABLE skitapi.deployments RENAME COLUMN s3_total_size_in_bytes TO client_package_size;
    ALTER TABLE skitapi.deployments ALTER COLUMN client_package_size TYPE BIGINT;

  EXCEPTION
    WHEN undefined_column THEN
  END;
END $$;