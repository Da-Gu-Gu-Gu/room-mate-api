CREATE TYPE "status" AS ENUM ('pending', 'accepted', 'rejected');

CREATE TABLE IF NOT EXISTS "room_invitations"(
    "id" UUID NOT NULL,
    "room_id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "status" status NOT NULL,
    "invited_by" UUID NOT NULL
);
ALTER TABLE
    "room_invitations" ADD PRIMARY KEY("id");