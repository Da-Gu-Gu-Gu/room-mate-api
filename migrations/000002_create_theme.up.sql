CREATE TABLE IF NOT EXISTS "themes"(
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "primary_color" TEXT NOT NULL,
    "secondary_color" TEXT NOT NULL,
    "background_color" TEXT NOT NULL,
    "background_img_url" VARCHAR(255) NULL
);
