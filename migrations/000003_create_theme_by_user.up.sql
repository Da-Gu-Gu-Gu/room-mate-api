CREATE TABLE IF NOT EXISTS "users-themes"(
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "user_id" UUID NOT NULL,
    "theme_id" UUID NOT NULL
);


