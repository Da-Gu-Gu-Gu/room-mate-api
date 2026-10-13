DO $$ 
BEGIN
    CREATE TYPE role AS ENUM ('owner', 'member');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS "room_members"(
    "id" UUID NOT NULL,
    "room_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "role" role NOT NULL,
    "joined_at" DATE NOT NULL,
    "leave_at" DATE NOT NULL
);
