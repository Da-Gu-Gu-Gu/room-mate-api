CREATE TABLE IF NOT EXISTS "users"(
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "name" TEXT NOT NULL,
    "profile_url" CHAR(255),
    "email" VARCHAR(255) NOT NULL UNIQUE,
    "created_at" DATE NOT NULL
);
