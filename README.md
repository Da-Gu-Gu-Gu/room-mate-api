# Room Mate API

Go API for the Room Mate household cost-sharing app.

## Room membership lifecycle

All endpoints below require a valid access token.

| Method | Endpoint | Purpose |
| --- | --- | --- |
| POST | /api/v1/rooms | Create a room. The authenticated user becomes its owner/member atomically. |
| GET | /api/v1/rooms/:id/members | List active members. The caller must be an active member of the room. |
| POST | /api/v1/rooms/:id/invitations | Invite a user by inviteeId UUID. Only the room owner can invite. |
| POST | /api/v1/invitations/:invitationId/accept | Accept an invitation and join its room. |
| POST | /api/v1/invitations/:invitationId/decline | Decline an invitation. |
| POST | /api/v1/rooms/:id/leave | Leave a room as a non-owner member. |
| DELETE | /api/v1/rooms/:id | Delete a room as its creator. Active memberships are closed and invitations are removed in the same transaction. |

### Membership rules

- A user can belong to only one active room at a time.
- Leaving sets leave_at; it does not delete the membership history.
- Only the invitee can accept or decline their invitation.
- Invitations must still be pending to be accepted or declined.
- A user already in an active room cannot accept another invitation.
- The owner cannot leave while ownership transfer is not implemented; the owner can delete the room instead.
- Database partial unique indexes enforce one active membership per user and prevent duplicate pending invitations for the same room/user.

Apply migration 000011_room_member_lifecycle before using these endpoints. It normalizes legacy duplicate active memberships before adding the constraint.
