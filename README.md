# Room Mate API

Go backend for **Room Mate**, a household cost-sharing app. The API supports social sign-in, user profiles, shared rooms, room invitations and membership, household bills, bill-share allocation, themes, and reusable image uploads.

This README describes the implementation currently registered in the API router. The source code is authoritative; functionality that has handlers or DTOs but no registered route is identified as not exposed.

## Contents

- [Technology stack](#technology-stack)
- [Architecture](#architecture)
- [Getting started](#getting-started)
- [Configuration](#configuration)
- [Database and migrations](#database-and-migrations)
- [API conventions](#api-conventions)
- [Endpoint reference](#endpoint-reference)
- [Application flows](#application-flows)
- [Image uploads and S3](#image-uploads-and-s3)
- [Room membership rules](#room-membership-rules)
- [Development commands](#development-commands)

## Technology stack

- Go 1.26
- Gin HTTP framework
- PostgreSQL with GORM
- Google and Apple identity-token verification
- JWT access tokens signed with HS256
- AWS S3 using AWS SDK for Go v2
- godotenv for local environment-file loading
- golang-migrate for SQL migrations
- MIME detection with github.com/gabriel-vasile/mimetype

## Architecture

The request flow for protected API operations is:

~~~text
HTTP request
    -> Gin router
    -> JWT authentication middleware
    -> Handler (parse and validate input)
    -> Service (application/business rules)
    -> Repository (database operations through GORM)
    -> PostgreSQL
~~~

Image upload flow:

~~~text
multipart/form-data request
    -> UploadHandler (size and MIME validation)
    -> UploadService (generate unique object key)
    -> S3 storage
    -> public imageUrl returned to client
~~~

Main source directories:

| Path | Responsibility |
| --- | --- |
| cmd/server | Application entry point |
| internal/config | Loads environment configuration |
| internal/database | Opens the PostgreSQL connection |
| internal/router | Builds dependencies and registers routes |
| internal/handler | HTTP handlers |
| internal/middleware | JWT authentication |
| internal/dto | Request and response structures |
| internal/service | Application rules and orchestration |
| internal/repository | Database access and transactions |
| internal/model | GORM models |
| internal/storage | S3 implementation |
| migrations | Versioned up/down SQL migrations |

## Getting started

### Prerequisites

- Go 1.26 or a compatible newer toolchain
- Docker Compose and Docker, or a local PostgreSQL server
- golang-migrate CLI
- AWS credentials and an S3 bucket if testing image upload functionality

### 1. Configure environment

Create a .env file in the repository root. See [Configuration](#configuration) for all environment keys read by the application.

### 2. Start PostgreSQL

~~~bash
make up
~~~

The Docker Compose file starts PostgreSQL 18 with these local development defaults:

| Setting | Value |
| --- | --- |
| Database | roommate |
| Username | user |
| Password | password |
| Port | 5432 |

The database connection currently uses localhost as its host and reads the database name, user, password, and port from environment variables.

### 3. Apply migrations

~~~bash
make migrate-up
~~~

### 4. Run the API

~~~bash
go mod download
make run
~~~

The server listens on the port configured by SERVER_PORT. API routes are grouped under /api/v1, except for the /hello smoke-test endpoint.

## Configuration

The configuration loader reads the following variables from .env:

| Variable | Purpose |
| --- | --- |
| SERVER_PORT | HTTP server port |
| SERVER_ENV | Environment label (currently loaded into config) |
| DATABASE_NAME | PostgreSQL database name |
| DATABASE_USER | PostgreSQL username |
| DATABASE_PASSWORD | PostgreSQL password |
| DATABASE_PORT | PostgreSQL port |
| JWT_SECRET | Secret used to sign and verify API JWTs |
| GOOGLE_CLIENT_ID | Expected Google token audience/client ID |
| APPLE_CLIENT_ID | Expected Apple token audience/client ID |
| AWS_REGION | AWS region used for S3 |
| AWS_S3_BUCKET | S3 bucket used for image uploads |

The application currently calls godotenv.Load(), so a readable .env file is expected when starting the server. The PostgreSQL host is hard-coded as localhost in the database connection code; it is not currently configurable through an environment variable.

When AWS_REGION and AWS_S3_BUCKET are both configured, the router initializes S3 storage. Without them, the API can start, but image upload and receipt URL validation operations that require storage will not work.

AWS credentials are loaded through the AWS SDK's default credential chain. For local development, use an AWS CLI profile or the standard AWS credential environment variables. In deployed environments, prefer an IAM role.

## Database and migrations

SQL migrations are stored in migrations/ and applied in version order. Use make migrate-up to apply pending migrations and make migrate-down to roll back one migration step.

Current schema concepts include:

- Users and provider identities for Google and Apple sign-in.
- Rooms and room membership history.
- Room invitations with pending/accepted/rejected states.
- Shared bill categories.
- Room bills and individually stored member shares.
- Receipt image object-key compatibility and the newer receipt image URL column.

Notable migrations include:

- 000011: room membership lifecycle constraints, including at most one active room per user and unique room codes/pending invitations.
- 000013: room bill receipt image object-key column for legacy records.
- 000014: room bill receipt image URL column.

Do not edit migrations that have already been applied to a database. Add a new paired up/down migration for future schema changes.

## API conventions

- Base path: /api/v1.
- All routes under /api/v1 require Authorization: Bearer <access-token>, except POST /auth/google and POST /auth/apple.
- GET /hello is also public and is outside the /api/v1 group.
- UUIDs are used for user, room, bill, category, and invitation identifiers unless a response DTO specifies otherwise.
- JSON request fields use camelCase.
- Image upload uses multipart/form-data with the field name image.
- Most validation errors return HTTP 400; unauthenticated requests return HTTP 401. Exact status codes vary by handler and are described below where useful.

Example request header:

~~~http
Authorization: Bearer <access-token>
Content-Type: application/json
~~~

## Endpoint reference

The following list reflects routes registered in internal/router/router.go.

### Public endpoints

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /hello | Basic server smoke test; returns a Hello, World! message |
| POST | /api/v1/auth/google | Verify a Google ID token and issue a Room Mate JWT |
| POST | /api/v1/auth/apple | Verify an Apple ID token and issue a Room Mate JWT |

Social login request body for either provider:

~~~json
{
  "idToken": "<provider-id-token>"
}
~~~

The API requires a verified email from the provider. For a new provider identity, it creates a user and auth identity in a database transaction. Existing provider identities resolve to their linked user. A successful login returns an access token, Bearer token type, and user data. The JWT is signed with HS256 and expires after 24 hours.

### Current user and profile

All endpoints in this subsection require authentication.

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /api/v1/me | Return the current user's profile, room if they have one, and selected theme if present |
| PUT | /api/v1/users | Update the current user's profile |
| GET | /api/v1/users/:id | Get the current user's profile; access to another user's ID is forbidden |
| DELETE | /api/v1/users | Delete the current user's account |
| POST | /api/v1/users/theme | Select a theme for the current user |
| GET | /api/v1/users/theme | Get the current user's selected theme |

User update accepts optional fields:

~~~json
{
  "name": "New display name",
  "email": "person@example.com",
  "profileUrl": "https://example.com/profile.jpg"
}
~~~

Theme selection request:

~~~json
{
  "themeId": "<theme-uuid>"
}
~~~

### Themes

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /api/v1/themes | List available themes |

Theme create, update, and delete handlers exist in the code, but they are **not currently registered as HTTP routes**. A theme detail handler also exists but is not exposed by the current router. The available theme routes are listing themes and reading/updating the authenticated user's selected theme through the user endpoints above.

### Rooms

| Method | Endpoint | Purpose |
| --- | --- | --- |
| POST | /api/v1/rooms | Create a room; the authenticated user is set as creator |
| GET | /api/v1/rooms/:id | Get a room by UUID |
| PUT | /api/v1/rooms/:id | Update room name and/or room number |
| DELETE | /api/v1/rooms/:id | Delete a room |

Create room request:

~~~json
{
  "roomName": "Samrong Apartment",
  "roomNumber": "A-1204"
}
~~~

Both roomName and roomNumber are required. The creator ID is taken from the authenticated user, not from the request body.

Room update accepts optional roomName and roomNumber fields.

### Room members and invitations

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /api/v1/rooms/:id/members | Get room detail and its member information |
| POST | /api/v1/rooms/:id/invitations | Invite a user to the room |
| GET | /api/v1/invitations | Get invitations for the authenticated user |
| POST | /api/v1/invitations/:invitationId/accept | Accept an invitation |
| POST | /api/v1/invitations/:invitationId/decline | Decline an invitation |
| POST | /api/v1/rooms/:id/transfer-ownership | Transfer room ownership |
| POST | /api/v1/rooms/:id/leave | Leave the room |

Invite request:

~~~json
{
  "inviteeId": "<user-uuid>"
}
~~~

Ownership transfer request:

~~~json
{
  "newOwnerId": "<active-member-user-uuid>"
}
~~~

The authenticated user is used as the inviter/requester; clients should not submit an inviter ID.

### Bill categories and room bills

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /api/v1/bill-categories | List bill categories |
| POST | /api/v1/rooms/:id/bills | Create a bill in a room |
| GET | /api/v1/rooms/:id/bills | List bills in a room |
| GET | /api/v1/rooms/:id/bills/:billId | Get one bill |
| POST | /api/v1/rooms/:id/bills/:billId/shares | Assign member share amounts for a bill |
| GET | /api/v1/rooms/:id/bills/:billId/shares | List member shares with payment summaries |
| POST | /api/v1/rooms/:id/bills/:billId/payments | Submit a partial or full payment with proof | 
| POST | /api/v1/rooms/:id/bills/:billId/payments/:paymentId/approve | Approve a pending payment (room owner only) |
| POST | /api/v1/rooms/:id/bills/:billId/payments/:paymentId/reject | Reject a pending payment (room owner only) |

Create bill request:

~~~json
{
  "categoryId": "<category-uuid>",
  "name": "Electricity bill",
  "description": "October electricity",
  "total": 1200,
  "dueDate": "2026-11-15T00:00:00Z",
  "receiptImageUrl": "https://your-bucket.s3.ap-southeast-1.amazonaws.com/uploads/images/<uuid>.jpg"
}
~~~

categoryId and name are required; total must be greater than zero. The room ID is taken from the URL and the creator is taken from the authenticated user. The service initializes the bill status as pending. The request DTO has a status field, but the current service does not use the submitted value when creating a bill.

Assign bill shares request:

~~~json
{
  "members": [
    {
      "userId": "<member-one-uuid>",
      "amount": 700
    },
    {
      "userId": "<member-two-uuid>",
      "amount": 500
    }
  ]
}
~~~

The members array must contain at least one item. Each item requires a userId and a positive amount. Share amounts may differ between members, but the sum must equal the bill total. Each user must be an active member of the room; duplicate users are rejected. Shares can only be assigned once per bill. The repository inserts all share rows in one transaction, so the operation succeeds or rolls back as a unit.

Get member shares:

~~~http
GET /api/v1/rooms/:id/bills/:billId/shares
~~~

The bill must belong to the room in :id, and the authenticated requester must be an active member of that room. The response contains a members array. Each member includes shareAmount, approvedAmount, pendingAmount, remainingAmount, and paymentStatus alongside the existing share and user fields. approvedAmount sums approved payments only; pendingAmount sums pending submissions separately. remainingAmount is shareAmount minus approvedAmount, so pending and rejected payments never reduce the outstanding balance. paymentStatus is unpaid when no payment has been approved, partially_paid when some but not all of the share is approved, and paid when the full share is approved.

## Application flows

### 1. Sign in

1. The client signs in with Google or Apple and obtains the provider ID token.
2. It sends the token to POST /api/v1/auth/google or POST /api/v1/auth/apple.
3. The API verifies the token and requires a verified email.
4. A new user and provider identity are created transactionally; returning users are loaded by provider identity.
5. The API returns a 24-hour JWT. The client includes it as a Bearer token on protected requests.

### 2. Create a room and invite housemates

1. An authenticated user creates a room with roomName and roomNumber.
2. The API uses the authenticated user as the creator and creates the room membership.
3. The owner invites an existing user using the room invitation endpoint.
4. The invitee checks their invitations.
5. The invitee accepts or declines the invitation.
6. On acceptance, the invitee joins the room if they do not already have an active room membership.

### 3. Create a bill and split it

1. If a receipt exists, the client uploads it through POST /api/v1/uploads/images.
2. The API validates the image and returns a permanent imageUrl.
3. The client creates the bill and passes that URL as receiptImageUrl.
4. The API verifies that the requester is an active member of the room and validates that the URL belongs to the configured S3 bucket and uploads/images/ prefix.
5. The API creates the bill with pending status.

### 5. Submit and review a bill payment

1. A member submits a payment against their own bill share through POST /api/v1/rooms/:id/bills/:billId/payments. The authenticated user is used as the payer; clients cannot choose the payer or initial status.
2. The request includes billShareId, a positive amount, proofUrl, and an optional note. Partial payments are supported.
3. The payment is created with pending status. Pending and approved amounts count toward the share's committed amount; rejected payments do not.
4. The active room owner reviews the payment using the approve or reject endpoint. Other members receive HTTP 403.
5. A review records reviewedBy and reviewedAt. Only pending payments can be reviewed; attempts to review an already approved or rejected payment return HTTP 409.
6. Review is transactional and locks the bill and payment rows so concurrent review requests cannot both change the same pending payment.

Approval and rejection do not accept a client-supplied status or reviewer ID. Both values are determined by the authenticated request and server-side action.

6. The client assigns each member's share through the bill shares endpoint. Amounts can be unequal, but must add up to the bill total.
7. Active room members can retrieve the room's bills.

### 4. Leave a room or transfer ownership

1. A member leaves through POST /api/v1/rooms/:id/leave.
2. Leaving closes the active membership by setting leave_at; membership history is retained.
3. The owner must transfer ownership to another active member before leaving.
4. Ownership transfer updates the room owner and member roles.
5. Room deletion is a separate creator-authorized operation.

## Image uploads and S3

The generic image endpoint can be reused by features that need an image URL. It requires authentication but does not require the user to belong to a room.

### Upload contract

| Method | Endpoint | Content type | Multipart field |
| --- | --- | --- | --- |
| POST | /api/v1/uploads/images | multipart/form-data | image |

Example:

~~~bash
curl -X POST "http://localhost:8080/api/v1/uploads/images" \
  -H "Authorization: Bearer <access-token>" \
  -F "image=@./receipt.jpg"
~~~

The endpoint:
- Accepts JPEG, PNG, and WebP image bytes.
- Enforces a maximum file size of 5 MB.
- Detects MIME type from the file contents; it does not rely only on the supplied filename or Content-Type header.
- Stores the file under uploads/images/ with a generated UUID filename.
- Returns HTTP 201 with an imageUrl field.

Example response:

~~~json
{
  "imageUrl": "https://your-bucket.s3.ap-southeast-1.amazonaws.com/uploads/images/<uuid>.jpg"
}
~~~

### Receipt URL storage

New room bills store the complete public URL in room_bills.receipt_image_url rather than storing only the S3 object key. The service validates the URL against the configured bucket/region host and requires the key to use the uploads/images/ prefix. Existing rows that contain receipt_image_key are still mapped to public URLs when returned, for backward compatibility.

The URL is a stable direct S3 URL, not a presigned URL. Anyone who can access the URL may view the object if the bucket policy permits public reads. Do not upload sensitive/private images under a publicly readable prefix.

### S3 setup

Set AWS_REGION and AWS_S3_BUCKET and configure AWS credentials through the AWS SDK default credential chain. The application uploads objects using PutObject. Configure S3 permissions for the application identity and public-read access only if permanent public image URLs are intended.

A bucket policy for public reads can be scoped to uploads/images/* (replace the bucket name):

~~~json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadUploadedImages",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::your-bucket/uploads/images/*"
    }
  ]
}
~~~

S3 Block Public Access settings must permit the policy if public reads are required. Keep object writes restricted to the application identity; do not grant public PutObject or DeleteObject.

## Room membership rules

- A user can have only one active room membership at a time.
- An active membership is represented by leave_at being NULL.
- Leaving closes the membership record rather than deleting its history.
- A user with an active room membership cannot accept another room invitation.
- Only the invitee can accept or decline their invitation.
- An invitation must be pending to be accepted or declined.
- Only the room owner can invite users or transfer ownership.
- The owner must transfer ownership to another active member before leaving.
- The new owner must already be an active member of the same room.
- Database partial unique indexes enforce one active room per user and prevent duplicate pending invitations for the same room and invitee.

## Development commands

| Command | Description |
| --- | --- |
| make up | Start PostgreSQL using docker/compose.yml |
| make down | Stop the Compose services |
| make run | Run the API server |
| make dev | Run the API server (same as make run) |
| make migrate-up | Apply all pending SQL migrations |
| make migrate-down | Roll back one migration step |
| make tidy | Run go mod tidy |

