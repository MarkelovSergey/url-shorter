package middleware

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	cookieName = "user_id"
	secretKey  = "your-secret-key-change-in-production"
)

type contextKey string

const userIDKey contextKey = "userID"

type UserClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie(cookieName)
		if err == nil && cookie.Value != "" {
			if id, valid := validateJWT(cookie.Value); valid {
				userID = id
			}
		}

		if userID == "" {
			userID = uuid.New().String()
			tokenString, err := generateJWT(userID)
			if err != nil {
				http.Error(w, "Failed to generate token", http.StatusInternalServerError)

				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    tokenString,
				Path:     "/",
				HttpOnly: true,
			})
		}

		ctx := SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)

	return userID, ok
}

func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func validateJWT(tokenString string) (string, bool) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(secretKey), nil
	})

	if err != nil {
		return "", false
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims.UserID, true
	}

	return "", false
}

func generateJWT(userID string) (string, error) {
	claims := UserClaims{
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secretKey))
}
