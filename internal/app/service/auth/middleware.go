package auth

import (
	"context"
	"go-developer-course-diploma/internal/app/service/auth/secure"
	"net/http"
)

func MiddlewareGeneratorAuthorization(userAuthorizationStore secure.UserAuthorization) (mw func(http.Handler) http.Handler) {
	mw = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := userAuthorizationStore.GetUserID(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserIDCtx, userID)))
		})
	}
	return
}
