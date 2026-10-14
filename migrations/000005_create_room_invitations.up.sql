
DO $$ 
BEGIN
    CREATE TYPE status AS ENUM ('pending', 'accepted', 'rejected');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS "room_invitations"(
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "room_id" UUID NOT NULL,
    "invitee_id" UUID NOT NULL,
    "status" status NOT NULL,
    "inviter_id" UUID NOT NULL
);
