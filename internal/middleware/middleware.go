package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const UserIDContextKey = "userID"

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Read the Authorization header.
		authHeader := c.GetHeader("Authorization")
		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid authorization header",
			})
			return
		}

		// 2. Parse and verify the JWT.
		claims := &jwt.RegisteredClaims{}

		token, err := jwt.ParseWithClaims(
			parts[1],
			claims,
			func(token *jwt.Token) (any, error) {
				// Accept only HS256 tokens.
				if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, jwt.ErrSignatureInvalid
				}

				return []byte(jwtSecret), nil
			},
			jwt.WithValidMethods([]string{
				jwt.SigningMethodHS256.Alg(),
			}),
		)

		if err != nil || token == nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		// 3. Require the claims used by our API.
		if claims.Subject == "" || claims.ExpiresAt == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			return
		}

		// 4. Validate that the subject is a user UUID.
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid user identity",
			})
			return
		}

		// 5. Store the authenticated user's ID for handlers.
		c.Set(UserIDContextKey, userID.String())

		// 6. Continue to the next handler.
		c.Next()
	}
}
