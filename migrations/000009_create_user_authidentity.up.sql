CREATE TABLE IF NOT EXISTS "auth_identities" (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(20) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_provider_identity
        UNIQUE (provider, provider_user_id),

    CONSTRAINT valid_auth_provider
        CHECK (provider IN ('google', 'apple'))
);