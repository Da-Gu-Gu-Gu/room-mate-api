ALTER TABLE "users"
    ALTER COLUMN "created_at" TYPE TIMESTAMPTZ
        USING ("created_at"::TIMESTAMP AT TIME ZONE 'UTC'),
    ALTER COLUMN "created_at" SET DEFAULT NOW();
