# Room Mate API

Go API for the Room Mate household cost-sharing app.

## Room bill receipt images

Room bill creation accepts either the existing JSON request (without an image) or `multipart/form-data` with an optional `receiptImage` file.

Multipart fields:
- `categoryId` (required UUID)
- `name` (required)
- `total` (required, greater than zero)
- `description` (optional)
- `dueDate` (optional, `YYYY-MM-DD` or RFC3339)
- `receiptImage` (optional JPEG, PNG, or WebP; maximum 5 MB)

Example:

```bash
curl -X POST "http://localhost:8080/api/v1/rooms/<room-id>/bills" \
  -H "Authorization: Bearer <access-token>" \
  -F "categoryId=<category-id>" \
  -F "name=Electricity bill" \
  -F "total=1200" \
  -F "dueDate=2026-11-15" \
  -F "receiptImage=@./electricity-receipt.jpg"
```

Set these environment variables to enable receipt uploads:

```env
AWS_REGION=ap-southeast-1
AWS_S3_BUCKET=your-private-room-mate-bucket
```

The AWS SDK uses its standard credential chain (for local development, configure the AWS CLI profile or `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`; in production, prefer an IAM role). Keep the bucket private and Block Public Access enabled. The app needs `s3:PutObject`, `s3:GetObject`, and `s3:DeleteObject` permissions for the `room-bills/*` object prefix.

The API validates the image bytes, limits uploads to 5 MB, stores the S3 object key in PostgreSQL, and returns a private presigned GET URL that expires after 10 minutes. Room membership is checked before upload. If saving the bill fails after the upload, the service attempts to delete the uploaded object.

Apply migration `000013_room_bill_receipt_image.up.sql` after the existing room-bill migrations. Run `go mod tidy` after pulling the AWS SDK dependencies to update `go.sum`.

## Room membership lifecycle

All endpoints below require a valid access token.

| Method | Endpoint | Purpose |
| --- | --- | --- |
| POST | /api/v1/rooms | Create a room. The authenticated user becomes its owner/member atomically. |
| GET | /api/v1/rooms/:id/members | List active members. The caller must be an active member of the room. |
| POST | /api/v1/rooms/:id/invitations | Invite a user by inviteeId UUID. Only the room owner can invite. |
| GET | /api/v1/invitations | List pending invitations for the authenticated user. |
| POST | /api/v1/invitations/:invitationId/accept | Accept an invitation and join its room. |
| POST | /api/v1/invitations/:invitationId/decline | Decline an invitation. |
| POST | /api/v1/rooms/:id/transfer-ownership | Transfer ownership to another active member using `newOwnerId`. Only the current owner can do this. |
| POST | /api/v1/rooms/:id/leave | Leave a room as a member. The owner must transfer ownership first. |
| DELETE | /api/v1/rooms/:id | Delete a room as its creator. Active memberships are closed and invitations are removed in the same transaction. |

### Membership rules

- A user can belong to only one active room at a time.
- Leaving sets leave_at; it does not delete the membership history.
- Only the invitee can accept or decline their invitation.
- Invitations must still be pending to be accepted or declined.
- A user already in an active room cannot accept another invitation.
- An owner must transfer ownership to another active member before leaving. Ownership transfer updates both the active membership roles and `rooms.created_by` in one transaction.
- A user can only become the new owner if they are already an active member; the one-active-room-per-user constraint remains unchanged.
- Database partial unique indexes enforce one active membership per user and prevent duplicate pending invitations for the same room/user.

Apply migration 000011_room_member_lifecycle before using these endpoints. It normalizes legacy duplicate active memberships before adding the constraint.
