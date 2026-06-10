DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_name = 'git_provider'
          AND column_name = 'password'
          AND character_maximum_length < 500
    ) THEN
        ALTER TABLE git_provider
            ALTER COLUMN password TYPE VARCHAR(500);
    END IF;
END $$;
