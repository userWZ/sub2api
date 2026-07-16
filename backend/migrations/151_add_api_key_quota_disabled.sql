-- Add OceanWay API-key-level quota switch without introducing key-level image permissions.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS quota_disabled BOOLEAN NOT NULL DEFAULT FALSE;
