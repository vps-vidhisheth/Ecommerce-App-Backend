package authmiddleware

import (
	"context"
	"ecommerce/components/log"
	Userservice "ecommerce/components/user/service"
	"ecommerce/errors"
	"ecommerce/security/token"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type contextKey string

const (
	ContextUserIDKey contextKey = "userID"
	ContextRoleKey   contextKey = "role"
)

func AuthMiddleware(userService *Userservice.UserService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &token.Claims{}
		jwtSecret := viper.GetString("JWT_SECRET")
		if jwtSecret == "" {
			log.GetLogger().Print("JWT_SECRET not set in environment variables")
			http.Error(w, "server misconfiguration", http.StatusInternalServerError)
			return
		}

		parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})
		if err != nil || !parsedToken.Valid {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		userID, err := uuid.Parse(claims.UserId)
		if err != nil {
			http.Error(w, "invalid user ID in token", http.StatusUnauthorized)
			return
		}

		user, err := userService.EnsureActiveUser(userID, claims.IssuedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ContextUserIDKey, user.ID)
		ctx = context.WithValue(ctx, ContextRoleKey, user.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return RoleMiddleware(next, "ADMIN")
}

func RoleMiddleware(next http.Handler, requiredRole string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(ContextRoleKey).(string)
		if !ok || strings.ToUpper(role) != strings.ToUpper(requiredRole) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, *errors.HTTPError) {
	userID, ok := ctx.Value(ContextUserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}
	return userID, nil
}

func GetUserRoleFromContext(ctx context.Context) (string, *errors.HTTPError) {
	role, ok := ctx.Value(ContextRoleKey).(string)
	if !ok {
		return "", errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}
	return role, nil
}

func EnsureOwnership(ctx context.Context, resourceOwnerID uuid.UUID) *errors.HTTPError {
	loggedInUserID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}
	if loggedInUserID != resourceOwnerID {
		return errors.NewHTTPError("forbidden: not your resource", http.StatusForbidden)
	}
	return nil
}
