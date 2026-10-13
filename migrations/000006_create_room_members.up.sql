CREATE TYPE "role" AS ENUM ('owner', 'member');

CREATE TABLE IF NOT EXISTS "room_members"(
    "id" UUID NOT NULL,
    "room_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "role" role NOT NULL,
    "joined_at" DATE NOT NULL,
    "leave_at" DATE NOT NULL
);
ALTER TABLE
    "room_members" ADD PRIMARY KEY("id");