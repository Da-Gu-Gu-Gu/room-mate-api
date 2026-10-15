ALTER TABLE "users"
    ALTER COLUMN "created_at" DROP DEFAULT,
    ALTER COLUMN "created_at" TYPE DATE
        USING (("created_at" AT TIME ZONE 'UTC')::DATE);
