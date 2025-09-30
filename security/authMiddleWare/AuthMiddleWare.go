package authmiddleware

import (
	"context"
	"ecommerce/components/log"
	"ecommerce/errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type contextKey string

const ClaimsContextKey contextKey = "jwtClaims"

func GetClaims(r *http.Request) (jwt.MapClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, errors.NewHTTPError("unauthorized", http.StatusUnauthorized)
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims := jwt.MapClaims{}
	jwtSecret := viper.GetString("JWT_SECRET")
	if jwtSecret == "" {
		log.GetLogger().Print("JWT_SECRET not set in environment variables")
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}

	return claims, nil
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetClaims(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminMiddleware(next http.Handler) http.Handler {
	return roleMiddleware(next, "admin")
}

func roleMiddleware(next http.Handler, requiredRole string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claimsData, ok := r.Context().Value(ClaimsContextKey).(jwt.MapClaims)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		role, ok := claimsData["role"].(string)
		if !ok || strings.ToLower(role) != strings.ToLower(requiredRole) {
			http.Error(w, "access denied", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, *errors.HTTPError) {
	claimsData, ok := ctx.Value(ClaimsContextKey).(jwt.MapClaims)
	if !ok {
		return uuid.Nil, errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}

	userIDStr, ok := claimsData["UserId"].(string)
	if !ok {
		return uuid.Nil, errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}

	userUUID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}

	return userUUID, nil
}

func GetUserRoleFromContext(ctx context.Context) (string, *errors.HTTPError) {
	claimsData, ok := ctx.Value(ClaimsContextKey).(jwt.MapClaims)
	if !ok {
		return "", errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}

	role, ok := claimsData["role"].(string)
	if !ok {
		return "", errors.NewHTTPError("invalid token", http.StatusUnauthorized)
	}

	return role, nil
}
