CREATE TABLE IF NOT EXISTS "theme"(
    "id" UUID NOT NULL,
    "primary_color" TEXT NOT NULL,
    "secondary_color" TEXT NOT NULL,
    "background_color" TEXT NOT NULL,
    "background_img_url" VARCHAR(255) NULL
);
ALTER TABLE
    "theme" ADD PRIMARY KEY("id");
