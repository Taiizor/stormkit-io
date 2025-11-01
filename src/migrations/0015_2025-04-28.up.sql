ALTER TABLE skitapi.apps ALTER COLUMN repo DROP NOT NULL;
ALTER TABLE skitapi.deployments ALTER COLUMN branch DROP NOT NULL;
ALTER TABLE skitapi.deployments ALTER COLUMN checkout_repo DROP NOT NULL;
