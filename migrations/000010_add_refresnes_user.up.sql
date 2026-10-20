ALTER TABLE "users-themes"
ADD CONSTRAINT fk_users_themes_user
FOREIGN KEY (user_id) REFERENCES users(id);

ALTER TABLE "users-themes"
ADD CONSTRAINT fk_users_themes_theme
FOREIGN KEY (theme_id) REFERENCES themes(id);