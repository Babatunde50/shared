package middlewares

import (
	"net/http"
	"strconv"
	"time"

	main_err "errors"

	"github.com/Babatunde50/shared/errors"
	"github.com/Babatunde50/shared/utils"
	"github.com/pascaldekloe/jwt"
)

func verifyJWT(token, jwtSecretKey string) (jwt.Claims, error) {

	claims, err := jwt.HMACCheck([]byte(token), []byte(jwtSecretKey))

	if err != nil {
		return *claims, err
	}

	if !claims.Valid(time.Now()) {
		return *claims, main_err.New("token is not valid at this time")
	}

	return *claims, nil

}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")

		if err != nil {
			errors.Unauthenticated(w, r)
			return
		}

		if err = cookie.Valid(); err != nil {
			errors.Unauthenticated(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CurrentUser(jwtSecretKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := verifyJWT(cookie.Value, jwtSecretKey)

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			email, ok := claims.Set["email"].(string)

			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			userId, _ := strconv.ParseInt(claims.Subject, 10, 64)

			next.ServeHTTP(w, r.WithContext(utils.ContextSetUserAuth(r, email, userId)))
		})
	}
}
