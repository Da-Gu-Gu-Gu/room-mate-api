# Room Mate API

Backend API for **Room Mate**, a household cost-sharing app. It helps people living together manage a shared room, invite housemates, track household bills, split costs (including unequal shares), and keep a record of bill receipts. The API is written in Go and uses PostgreSQL.

## Contents
- [Project overview](#project-overview)
- [Technology stack](#technology-stack)
- [Architecture](#architecture)
- [Getting started](#getting-started)
- [Environment variables](#environment-variables)
- [Database migrations](#database-migrations)
- [API conventions](#api-conventions)
- [Endpoints](#endpoints)
- [Main application flows](#main-application-flows)
- [Image uploads](#image-uploads)
- [Room membership rules](#room-membership-rules)
- [Development commands](#development-commands)

## Project overview

Room Mate is designed for people sharing a household. A user belongs to at most one active room at a time. Room members can invite other users, create household bills, assign each member a specific share amount, and attach a receipt image.

Current backend capabilities:
- Google and Apple ID-token authentication.
- User profile and current-user ("About me") endpoints.
- Room creation, update, retrieval, deletion, and membership management.
- Room invitations, acceptance/decline, leaving a room, and ownership transfer.
- Shared bill categories and room-specific bills.
- Bill-share assignment with different amounts per member.
- Generic image uploads to S3, returning permanent public image URLs.
- User-selected themes and theme listing.

The mobile client is planned separately; this repository contains the Go API.

## Technology stack
- **Language / HTTP:** Go, Gin
- **Database / ORM:** PostgreSQL, GORM
- **Authentication:** Google and Apple identity tokens; API access tokens are JWTs
- **Object storage:** AWS S3 (AWS SDK for Go v2)
- **Validation:** Gin binding/validator
- **Migrations:** golang-migrate CLI
- **Configuration:** godotenv

## Architecture

The code is organized by responsibility:

| Package | Responsibility |
| --- | --- |
| cmd/server | Application entry point; loads configuration, connects to PostgreSQL, and starts the HTTP server |
| internal/config | Reads environment configuration |
| internal/router | Builds dependencies and registers HTTP routes/middleware |
| internal/handler | Parses HTTP input and writes HTTP responses |
| internal/middleware | JWT authentication and request identity |
| internal/dto | Request and response payload types |
| internal/service | Application rules and use-case orchestration |
| internal/repository | Database access |
| internal/model | GORM/database models |
| internal/storage | S3 object-storage implementation |
| migrations | Versioned SQL schema changes |

Typical request path:

~~~text
HTTP request
    -> Gin router
    -> JWT middleware (protected routes)
    -> Handler (parse / validate request)
    -> Service (business rules)
    -> Repository (PostgreSQL via GORM)
    -> Handler response
~~~

Image upload requests use the upload service and S3 storage instead of the database repository.

## Getting started

### Prerequisites
- Go version declared in go.mod (currently Go 1.26.0).
- Docker and Docker Compose, or a local PostgreSQL server.
- golang-migrate CLI for applying migrations.
- AWS credentials and an S3 bucket if you want to test image uploads or bill receipts.

### 1. Configure environment

Create a .env file in the repository root. See [Environment variables](#environment-variables) for the keys read by the application.

### 2. Start PostgreSQL

~~~bash
make up
~~~

The Compose configuration starts PostgreSQL on localhost:5432 with the local development database roommate, user user, and password password.

### 3. Apply migrations

~~~bash
make migrate-up
~~~

The Makefile migration command currently uses the local development connection string. If your database credentials differ, update the command or run migrate with your own database URL.

### 4. Start the API

~~~bash
go mod download
make run
~~~

The server listens on SERVER_PORT. The API base path is /api/v1.

## Environment variables

The current configuration loader reads these exact keys:

| Variable | Purpose |
| --- | --- |
| SERVER_PORT | HTTP server port |
| SERVER_ENV | Environment label |
| DATABASE_NAME | PostgreSQL database name |
| DATABASE_USER | PostgreSQL username |
| DATABASE_PASSWORD | PostgreSQL password |
| DATABASE_PORT | PostgreSQL port |
| JWT_SECRET | Secret used to sign and verify API JWTs |
| GOOGLE_CLIENT_ID | Expected Google OAuth/OIDC client ID |
| APPLE_CLIENT_ID | Expected Apple client ID / audience |
| AWS_REGION | AWS region for S3 |
| AWS_S3_BUCKET | S3 bucket used for uploaded images |

The application currently requires a .env file to load successfully. The database connection is assembled by the database package; keep its host/connection settings aligned with your local or deployed PostgreSQL environment.

AWS SDK credentials are resolved through the standard AWS credential chain. For local development, configure an AWS CLI profile or the standard AWS credential environment variables. In production, prefer an IAM role.

## Database migrations

Migrations are stored in migrations/ and are applied in version order. Run make migrate-up after configuring the database.

Important schema concepts:
- Users are identified by UUIDs.
- auth_identities links a user to a provider identity (Google or Apple).
- rooms represent households.
- room_members stores membership history; an active membership has leave_at IS NULL.
- room_invitations stores pending and completed invitations.
- Room bills belong to a room and a bill category.
- Bill shares store a specific amount for each assigned member.
- New bill receipt URLs are stored in room_bills.receipt_image_url. Older rows with an object key remain supported by the API response mapper.

Do not edit an already-applied migration. Add a new up/down migration for schema changes.

## API conventions

- Base URL: http://localhost:<SERVER_PORT>/api/v1.
- Protected endpoints require Authorization: Bearer <access-token>.
- Authentication endpoints are public.
- IDs in request paths and payloads are UUIDs unless otherwise noted.
- JSON uses camelCase field names.
- Uploads use multipart/form-data; normal API payloads use JSON.
- The API uses standard HTTP status codes such as 200, 201, 400, 401, 403, 404, 409, 413, and 500. Some errors are returned as an error JSON field.

Example authenticated request header:

~~~http
Authorization: Bearer <access-token>
Content-Type: application/json
~~~

## Endpoints

All endpoints in this section are registered in internal/router/router.go. Unless marked **Public**, they require a valid JWT.

### Health / smoke test

| Method | Endpoint | Auth | Purpose |
| --- | --- | --- | --- |
| GET | /hello | Public | Simple server smoke test |

### Authentication

| Method | Endpoint | Auth | Purpose |
| --- | --- | --- | --- |
| POST | /auth/google | Public | Verify a Google ID token; create or find the user and issue an API JWT |
| POST | /auth/apple | Public | Verify an Apple ID token; create or find the user and issue an API JWT |

Google and Apple request body:

~~~json
{
  "idToken": "<provider-id-token>"
}
~~~

Successful authentication returns an access token, token type (Bearer), and user profile. The API JWT currently expires after 24 hours.

### Current user and profile

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /me | Get the authenticated user's profile, current room (if any), and selected theme (if any) |
| PUT | /users | Update the authenticated user's profile |
| GET | /users/:id | Get the authenticated user's own profile; requesting another user's ID is forbidden |
| DELETE | /users | Delete the authenticated user's account |
| POST | /users/theme | Set the authenticated user's selected theme |
| GET | /users/theme | Get the authenticated user's selected theme |

### Themes

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /themes | List available themes |

Theme selection request:

~~~json
{
  "themeId": "<theme-uuid>"
}
~~~

Theme creation/update handlers exist in the code, but create, update, and delete theme routes are not currently registered in the router. The public API currently exposes only the theme list and the authenticated user's theme selection/read endpoints.

### Rooms and membership

| Method | Endpoint | Purpose |
| --- | --- | --- |
| POST | /rooms | Create a room; the authenticated user becomes its owner and active member |
| GET | /rooms/:id | Get room details |
| PUT | /rooms/:id | Update room details |
| DELETE | /rooms/:id | Delete a room as its creator |
| GET | /rooms/:id/members | Get room details and active members |
| POST | /rooms/:id/invitations | Invite a user by inviteeId |
| POST | /invitations/:invitationId/accept | Accept an invitation |
| POST | /invitations/:invitationId/decline | Decline an invitation |
| GET | /invitations | List the authenticated user's invitations |
| POST | /rooms/:id/transfer-ownership | Transfer room ownership to another active member |
| POST | /rooms/:id/leave | Leave the current room |

Create room example:

~~~json
{
  "roomName": "Samrong Apartment",
  "roomNumber": "A-1204"
}
~~~

Invite member example:

~~~json
{
  "inviteeId": "<user-uuid>"
}
~~~

Ownership transfer example:

~~~json
{
  "newOwnerId": "<active-member-user-uuid>"
}
~~~

### Bills

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | /bill-categories | List available bill categories |
| POST | /rooms/:id/bills | Create a bill in a room |
| GET | /rooms/:id/bills | List bills for a room |
| GET | /rooms/:id/bills/:billId | Get one bill |
| POST | /rooms/:id/bills/:billId/shares | Assign the bill's share amounts to room members |

Create bill example:

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

The server sets the room ID and creator from the URL and authenticated user. The current create flow initializes bill status as pending.

Assign shares example:

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

Each member's amount is explicit, so shares do not need to be equal. The service/repository validates the selected members and share total against the bill total. Share assignment is a separate request after bill creation.

### Image uploads

| Method | Endpoint | Purpose |
| --- | --- | --- |
| POST | /uploads/images | Upload a reusable image and return its public URL |

See [Image uploads](#image-uploads) for the upload contract and storage configuration.

## Main application flows

### 1. Sign in
1. The mobile client signs in with Google or Apple and obtains the provider ID token.
2. The client sends idToken to POST /api/v1/auth/google or POST /api/v1/auth/apple.
3. The API verifies the provider token and requires a verified email.
4. If the provider identity already exists, the API loads its linked user. Otherwise, it creates the user and provider identity in a database transaction.
5. The API returns a 24-hour JWT. The client sends it as a Bearer token on protected endpoints.

### 2. Create a room and invite housemates
1. An authenticated user calls POST /rooms.
2. The API creates the room and the creator's active owner membership together.
3. The owner invites another existing user through POST /rooms/:id/invitations.
4. The invitee checks GET /invitations.
5. The invitee accepts or declines through the matching invitation action.
6. On acceptance, the API creates an active membership and marks the invitation accepted. The invitee must not already belong to another active room.

### 3. Create a bill and assign shares
1. A room member uploads a receipt first, if there is one, using POST /uploads/images.
2. The API validates the file and returns imageUrl.
3. The client sends the bill details and returned receiptImageUrl to POST /rooms/:id/bills.
4. The API checks that the requester can create bills in the room and validates that the receipt URL belongs to the configured S3 bucket and uploads/images/ prefix.
5. The bill is saved with status pending.
6. The client calls POST /rooms/:id/bills/:billId/shares with one amount per member. Amounts can differ between members.
7. Members can retrieve the bill and its shares through the bill read endpoints.

### 4. Leave a room or transfer ownership
1. A regular member can call POST /rooms/:id/leave; the membership is closed by setting leave_at.
2. The current owner must first transfer ownership to another active member.
3. Ownership transfer updates the room creator and the active membership roles in one transaction.
4. The previous owner can then leave.
5. Room deletion is a separate creator-only operation; it closes active memberships and removes invitations as part of the delete flow.

## Image uploads

Use one reusable endpoint for profile pictures, room themes, bill receipts, payment proofs, and other image features. The endpoint requires authentication and is independent of room membership.

### Upload an image

~~~bash
curl -X POST "http://localhost:8080/api/v1/uploads/images" \
  -H "Authorization: Bearer <access-token>" \
  -F "image=@./receipt.jpg"
~~~

The endpoint accepts JPEG, PNG, or WebP images up to 5 MB. The server detects the actual MIME type from the file bytes rather than trusting the filename or client-provided content type.

Successful response (HTTP 201):

~~~json
{
  "imageUrl": "https://your-room-mate-bucket.s3.ap-southeast-1.amazonaws.com/uploads/images/<uuid>.jpg"
}
~~~

Use the returned URL in later JSON requests. For example, room bill creation accepts it as receiptImageUrl.

New room bills store the complete public URL in receipt_image_url, not the S3 object key. The API validates the URL against the configured S3 bucket and the generic uploads/images/ prefix. Existing records that contain an S3 object key remain readable for backward compatibility.

### AWS configuration and access

Set AWS_REGION and AWS_S3_BUCKET. The AWS SDK uses its standard credential chain (for local development, configure the AWS CLI profile or AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY; in production, prefer an IAM role). The application needs s3:PutObject for uploads/images/*.

Image URLs are permanent direct S3 URLs, not presigned URLs. Anyone who obtains a URL can view that image without signing in. Do not upload images containing information that should remain private.

A bucket policy can grant public read access to uploaded images (replace the bucket name):

~~~json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadUploadedImages",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::your-room-mate-bucket/uploads/images/*"
    }
  ]
}
~~~

For this policy to work, S3 Block Public Access settings that block public bucket policies must allow it, including applicable account-level settings. Keep ACLs disabled; use a bucket policy rather than public-read object ACLs. Do not grant public PutObject or DeleteObject.

Configure an S3 lifecycle rule to expire unused uploads after a suitable period if clients may upload an image but never attach its URL to a record.

## Room membership rules

- A user can belong to only one active room at a time.
- Leaving sets leave_at; it does not delete membership history.
- Only the invitee can accept or decline their invitation.
- Invitations must still be pending to be accepted or declined.
- A user already in an active room cannot accept another invitation.
- Only the room owner can invite users and transfer ownership.
- An owner must transfer ownership to another active member before leaving.
- Ownership transfer updates the active membership roles and rooms.created_by in one transaction.
- A user can only become the new owner if they are already an active member.
- Database partial unique indexes enforce one active membership per user and prevent duplicate pending invitations for the same room/user.

Apply the membership lifecycle migration before using the related endpoints.

## Development commands

| Command | Description |
| --- | --- |
| make up | Start local PostgreSQL with Docker Compose |
| make down | Stop the local PostgreSQL container |
| make run | Run the API |
| make dev | Run the API (same command as make run) |
| make migrate-up | Apply all pending migrations |
| make migrate-down | Roll back one migration step (golang-migrate default) |
| make tidy | Run go mod tidy |
