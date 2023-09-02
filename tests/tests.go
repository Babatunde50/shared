package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/Babatunde50/shared/utils"
	"github.com/pascaldekloe/jwt"
)

type TestServer struct {
	*httptest.Server
}

func NewTestServer(t *testing.T, h http.Handler) *TestServer {
	ts := httptest.NewServer(h)
	return &TestServer{ts}
}

func (ts *TestServer) Send(t *testing.T, urlPath string, method string, headers *map[string]string, body *map[string]interface{}, authUser *utils.AuthenticatedUser, baseURL string, jwtSecret string) (int, http.Header, []byte) {

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	client := &http.Client{
		Jar: jar,
	}

	var bodyReq io.Reader

	if body != nil {
		bodyInBytes, err := json.Marshal(body)

		if err != nil {
			panic(err)
		}

		bodyReq = bytes.NewReader(bodyInBytes)
	}

	req, err := http.NewRequest(method, ts.URL+urlPath, bodyReq)
	if err != nil {
		t.Fatal(err)
	}

	if authUser != nil {
		w := httptest.NewRecorder()
		if err = LogUserIn(w, req, *authUser, baseURL, jwtSecret); err != nil {
			t.Fatal(err)
		}
		for _, cookie := range w.Result().Cookies() {
			req.AddCookie(cookie)
		}
	}

	if headers != nil {
		for key, value := range *headers {
			req.Header.Set(key, value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.StatusCode, resp.Header, respBody
}

func LogUserIn(w http.ResponseWriter, r *http.Request, user utils.AuthenticatedUser, baseURL string, secretKey string) error {

	jwt, err := generateJWT(user, baseURL, secretKey)

	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwt,
		Path:     "/",
		MaxAge:   86400 * 1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Duration(24) * time.Hour),
	})

	return nil
}

func generateJWT(user utils.AuthenticatedUser, baseURL string, secretKey string) (string, error) {

	var claims jwt.Claims

	claims.Subject = strconv.Itoa(int(user.Id))
	claims.Set = map[string]interface{}{
		"id":    user.Id,
		"email": user.Email,
	}

	expiry := time.Now().Add(time.Duration(24) * time.Hour)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(expiry)

	claims.Issuer = baseURL
	claims.Audiences = []string{baseURL}

	jwtBytes, err := claims.HMACSign(jwt.HS256, []byte(secretKey))
	if err != nil {
		return string(jwtBytes), err
	}

	return string(jwtBytes), nil

}
