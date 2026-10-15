-- Roll back the timestamps added by 000008.
ALTER TABLE "room_members"
    DROP COLUMN IF EXISTS "updated_at",
    DROP COLUMN IF EXISTS "created_at";

ALTER TABLE "room_invitations"
    DROP COLUMN IF EXISTS "updated_at",
    DROP COLUMN IF EXISTS "created_at";

ALTER TABLE "users-themes"
    DROP COLUMN IF EXISTS "updated_at",
    DROP COLUMN IF EXISTS "created_at";

ALTER TABLE "themes"
    DROP COLUMN IF EXISTS "updated_at",
    DROP COLUMN IF EXISTS "created_at";

ALTER TABLE "rooms"
    DROP COLUMN IF EXISTS "updated_at",
    ALTER COLUMN "created_at" DROP DEFAULT,
    ALTER COLUMN "created_at" TYPE DATE
        USING (("created_at" AT TIME ZONE 'UTC')::DATE);

ALTER TABLE "users"
    DROP COLUMN IF EXISTS "updated_at";
