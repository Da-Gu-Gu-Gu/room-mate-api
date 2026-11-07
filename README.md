# Room Mate API

Go API for the Room Mate household cost-sharing app.

## Generic image uploads

Use one reusable endpoint to upload images for any feature (profile pictures, room themes, bill receipts, payment proofs, etc.). The endpoint requires authentication and does not depend on a room.

### Upload an image

    curl -X POST "http://localhost:8080/api/v1/uploads/images" \
      -H "Authorization: Bearer <access-token>" \
      -F "image=@./receipt.jpg"

The endpoint accepts JPEG, PNG, or WebP images up to 5 MB. It returns a permanent public URL:

    {
      "imageUrl": "https://your-room-mate-bucket.s3.ap-southeast-1.amazonaws.com/uploads/images/<uuid>.jpg"
    }

Use the returned URL in any later JSON payload. For example, room bill creation accepts it as receiptImageUrl:

    curl -X POST "http://localhost:8080/api/v1/rooms/<room-id>/bills" \
      -H "Authorization: Bearer <access-token>" \
      -H "Content-Type: application/json" \
      -d '{
        "categoryId": "<category-id>",
        "name": "Electricity bill",
        "total": 1200,
        "receiptImageUrl": "https://your-room-mate-bucket.s3.ap-southeast-1.amazonaws.com/uploads/images/<uuid>.jpg"
      }'

The API validates that a bill receipt URL belongs to the configured S3 bucket and uses the generic uploads/images/ prefix. Older room-specific receipt URLs are still accepted.

### AWS configuration

Set AWS_REGION and AWS_S3_BUCKET. The AWS SDK uses its standard credential chain (for local development, configure the AWS CLI profile or AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY; in production, prefer an IAM role). The application needs s3:PutObject for uploads/images/* and any feature-specific prefixes still in use.

### Public image access

Image URLs are permanent direct S3 URLs, not presigned URLs. Anyone who obtains a URL can view that image without signing in. Do not upload images containing information that should remain private.

Add a bucket policy like the following, replacing the bucket name:

    {
      "Version": "2012-10-17",
      "Statement": [{
        "Sid": "PublicReadUploadedImages",
        "Effect": "Allow",
        "Principal": "*",
        "Action": "s3:GetObject",
        "Resource": "arn:aws:s3:::your-room-mate-bucket/uploads/images/*"
      }]
    }

For this bucket policy to work, S3 Block Public Access settings that block public bucket policies must allow it (including applicable account-level settings). Keep ACLs disabled; use a bucket policy rather than public-read object ACLs. Do not grant public PutObject or DeleteObject.

Configure an S3 lifecycle rule to expire unused uploads after a suitable period if clients may upload an image but never attach its URL to a record.

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
| POST | /api/v1/rooms/:id/transfer-ownership | Transfer ownership to another active member using newOwnerId. Only the current owner can do this. |
| POST | /api/v1/rooms/:id/leave | Leave a room as a member. The owner must transfer ownership first. |
| DELETE | /api/v1/rooms/:id | Delete a room as its creator. Active memberships are closed and invitations are removed in the same transaction. |

### Membership rules

- A user can belong to only one active room at a time.
- Leaving sets leave_at; it does not delete the membership history.
- Only the invitee can accept or decline their invitation.
- Invitations must still be pending to be accepted or declined.
- A user already in an active room cannot accept another invitation.
- An owner must transfer ownership to another active member before leaving. Ownership transfer updates both the active membership roles and rooms.created_by in one transaction.
- A user can only become the new owner if they are already an active member; the one-active-room-per-user constraint remains unchanged.
- Database partial unique indexes enforce one active membership per user and prevent duplicate pending invitations for the same room/user.

Apply migration 000011_room_member_lifecycle before using these endpoints.
