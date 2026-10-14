CREATE TABLE IF NOT EXISTS "rooms"(
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "room_name" VARCHAR(255) NOT NULL,
    "room_code" VARCHAR(255) NOT NULL, 
    "room_number" VARCHAR(255) NOT NULL,
    "created_by" UUID NOT NULL,
    "created_at" DATE NOT NULL
);

