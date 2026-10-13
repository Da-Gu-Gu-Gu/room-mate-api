CREATE TYPE "status" AS ENUM ('pending', 'accepted', 'rejected');

CREATE TABLE IF NOT EXISTS "room_invitations"(
    "id" UUID NOT NULL,
    "room_id" UUID NOT NULL,
    "invitee_id" UUID NOT NULL,
    "status" status NOT NULL,
    "inviter_id" UUID NOT NULL
);
ALTER TABLE
    "room_invitations" ADD PRIMARY KEY("id");