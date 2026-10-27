-- Keep each user in at most one active room. Preserve historical memberships.
-- Normalize any legacy duplicate active memberships before adding the partial index.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY user_id
               ORDER BY joined_at DESC NULLS LAST, created_at DESC, id DESC
           ) AS rn
    FROM room_members
    WHERE leave_at IS NULL
)
UPDATE room_members AS rm
SET leave_at = NOW(), updated_at = NOW()
FROM ranked
WHERE rm.id = ranked.id AND ranked.rn > 1;

-- Older builds used a fixed room code. Give duplicate codes stable unique values.
WITH duplicates AS (
    SELECT id, room_code,
           COUNT(*) OVER (PARTITION BY room_code) AS code_count
    FROM rooms
)
UPDATE rooms AS r
SET room_code = UPPER(SUBSTRING(REPLACE(r.id::text, '-', '') FROM 1 FOR 8)),
    updated_at = NOW()
FROM duplicates AS d
WHERE r.id = d.id AND d.code_count > 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_room_members_one_active_room_per_user
    ON room_members (user_id)
    WHERE leave_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_rooms_room_code_unique
    ON rooms (room_code);

CREATE UNIQUE INDEX IF NOT EXISTS idx_room_invitations_one_pending_per_user
    ON room_invitations (room_id, invitee_id)
    WHERE status = 'pending';
