package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
)

const (
	cookieName = "user_id"
	secretKey  = "your-secret-key-change-in-production"
)

type contextKey string

const userIDKey contextKey = "userID"

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie(cookieName)
		if err == nil && cookie.Value != "" {
			if id, valid := validateSignedCookie(cookie.Value); valid {
				userID = id
			}
		}

		if userID == "" {
			userID = uuid.New().String()
			signedValue := signCookie(userID)

			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    signedValue,
				Path:     "/",
				HttpOnly: true,
			})
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	
	return userID, ok
}

func signCookie(userID string) string {
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	signature := hex.EncodeToString(h.Sum(nil))

	return userID + "." + signature
}

func validateSignedCookie(signedValue string) (string, bool) {
	dotIndex := -1
	for i := len(signedValue) - 1; i >= 0; i-- {
		if signedValue[i] == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 || dotIndex == 0 || dotIndex == len(signedValue)-1 {
		return "", false
	}

	userID := signedValue[:dotIndex]
	providedSignature := signedValue[dotIndex+1:]

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	if hmac.Equal([]byte(expectedSignature), []byte(providedSignature)) {
		return userID, true
	}

	return "", false
}
