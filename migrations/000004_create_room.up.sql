CREATE TABLE IF NOT EXISTS "room"(
    "id" UUID NOT NULL,
    "room_name" VARCHAR(255) NOT NULL,
    "room_code" VARCHAR(255) NOT NULL, 
    "created_by" UUID NOT NULL,
    "created_at" DATE NOT NULL
);
ALTER TABLE
    "room" ADD PRIMARY KEY("id");
