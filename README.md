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
AWS_S3_BUCKET=your-room-mate-bucket
```

The AWS SDK uses its standard credential chain (for local development, configure the AWS CLI profile or `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`; in production, prefer an IAM role). The application needs `s3:PutObject` and `s3:DeleteObject` permissions for the `room-bills/*` object prefix.

### Public image access

Receipt URLs are permanent direct S3 URLs, not presigned URLs. Anyone who obtains a receipt URL can view that image without signing in. Do not upload receipts containing information that should remain private.

Add a bucket policy like the following, replacing the bucket name. It grants public read access only to objects under `room-bills/*`:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadRoomBillReceiptImages",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::your-room-mate-bucket/room-bills/*"
    }
  ]
}
```

For this bucket policy to work, the S3 Block Public Access setting that blocks public bucket policies must allow this policy (and any applicable account-level setting must also allow it). Keep ACLs disabled; use the bucket policy above rather than `public-read` object ACLs. Do not grant public `PutObject` or `DeleteObject`.

The API validates image bytes, limits uploads to 5 MB, stores the S3 object key in PostgreSQL, and returns a stable URL such as `https://your-room-mate-bucket.s3.ap-southeast-1.amazonaws.com/room-bills/<room-id>/<bill-id>/receipt.jpg`. Room membership is checked before upload. If saving the bill fails after the upload, the service attempts to delete the uploaded object.

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
