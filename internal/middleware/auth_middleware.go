package middleware

import (
	"backEnd-RingoTechLife/pkg"
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
	RoleAdmin string     = "ADMIN"
	RoleUser  string     = "USER"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ambil Authorization header
		authHeader := r.Header.Get("Authorization")

		// Extract token dari header
		token, err := pkg.GetAccessToken(authHeader)
		if err != nil {
			fmt.Print("\nkenapa bisa error disini anjing ", err, "\n")
			pkg.JSONError(w, 401, err.Error())
			return
		}

		// Verify token
		claims, err := pkg.VerifyToken(token)
		if err != nil {
			pkg.JSONError(w, 401, "token tidak valid atau kadaluarsa")
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, RoleKey, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := GetRole(r.Context())
			if !ok {
				fmt.Println(role)
				pkg.JSONError(w, http.StatusForbidden, "Kamu harus login sebelum bisa mengakses fitur ini")
				return
			}
			allowed := slices.Contains(allowedRoles, role)

			if !allowed {
				pkg.JSONError(w, http.StatusForbidden, "Kamu tidak punya akses ke fitur ini")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AuthMiddlewareFromCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("access_token")
		fmt.Println("cookie ada ga : ", cookie)
		if err != nil {
			if err == http.ErrNoCookie {
				pkg.JSONError(w, 401, "no cookie found!")
				return
			}
			pkg.JSONError(w, 400, "terjadi kesalahan saat membaca cookie")
			return
		}

		claims, err := pkg.VerifyRefreshToken(cookie.Value)
		if err != nil {
			pkg.JSONError(w, 401, "token tidak valid atau kadaluarsa")
			return
		}
		fmt.Println(claims)
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(RoleKey).(string)
	return role, ok
}
