package auth

import (
	"go-developer-course-diploma/internal/app/service/auth/secure"
	"net/http"
)

func MiddlewareGeneratorAuthorization(userAuthorizationStore secure.UserAuthorization) (mw func(http.Handler) http.Handler) {
	mw = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userAuthorizationStore.IsValidAuthorization(r) {
				next.ServeHTTP(w, r)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
		})
	}
	return
}
