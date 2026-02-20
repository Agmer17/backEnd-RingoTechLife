package middleware

import (
	"backEnd-RingoTechLife/pkg"
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
	RoleAdmin string     = "ADMIN"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ambil Authorization header
		authHeader := r.Header.Get("Authorization")

		// Extract token dari header
		token, err := pkg.GetAccessToken(authHeader)
		if err != nil {
			pkg.JSONError(w, 401, err.Error())
			return
		}

		// Verify token
		claims, err := pkg.VerifyToken(token)
		if err != nil {
			pkg.JSONError(w, 401, "token tidak valid atau kadaluarsa")
			return
		}

		// Simpan user info ke context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)

		// Lanjutkan ke handler berikutnya
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				pkg.JSONError(w, http.StatusForbidden, "Kamu harus login sebelum bisa mengakses fitur ini")
				return
			}

			// Cek apakah role user termasuk dalam allowed roles
			allowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				pkg.JSONError(w, http.StatusForbidden, "Kamu tidak punya akses ke fitur ini")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}
