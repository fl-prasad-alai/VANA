package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-that-is-definitely-longer-than-32-chars"

func newTestServer() *Server {
	return &Server{jwtSecret: []byte(testSecret)}
}

func call(t *testing.T, h http.Handler, method, path, body, token string) (int, map[string]interface{}) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var out map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return rec.Code, out
}

func signed(t *testing.T, secret string, method jwt.SigningMethod, claims jwt.RegisteredClaims) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(method, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestRequireAuth(t *testing.T) {
	s := newTestServer()
	h := s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"user": userIDFrom(r)})
	})
	uid := "508794a7-91d8-4ad1-8289-fb733a171e72"
	valid, _ := s.issueToken(uid)
	future := jwt.NewNumericDate(time.Now().Add(time.Hour))

	cases := []struct {
		name  string
		token string
		want  int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"garbage", "not-a-jwt", http.StatusUnauthorized},
		{"wrong secret", signed(t, "another-secret-another-secret-123456", jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: uid, ExpiresAt: future}), http.StatusUnauthorized},
		{"expired", signed(t, testSecret, jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: uid, ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}), http.StatusUnauthorized},
		{"no expiry", signed(t, testSecret, jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: uid}), http.StatusUnauthorized},
		{"wrong alg", signed(t, testSecret, jwt.SigningMethodHS512, jwt.RegisteredClaims{Subject: uid, ExpiresAt: future}), http.StatusUnauthorized},
		{"non-uuid subject", signed(t, testSecret, jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: "1", ExpiresAt: future}), http.StatusUnauthorized},
		{"valid", valid, http.StatusOK},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			code, body := call(t, h, http.MethodGet, "/", "", c.token)
			if code != c.want {
				t.Fatalf("got %d, want %d (%v)", code, c.want, body)
			}
			if c.want == http.StatusOK && body["user"] != uid {
				t.Fatalf("user id from token = %v, want %s", body["user"], uid)
			}
		})
	}
}

func TestProtectedRoutesRejectAnonymous(t *testing.T) {
	h := newTestServer().Handler()
	for _, r := range []struct{ method, path string }{
		{http.MethodPost, "/api/chat"},
		{http.MethodGet, "/api/conversations"},
		{http.MethodGet, "/api/conversations/508794a7-91d8-4ad1-8289-fb733a171e72/messages"},
		{http.MethodGet, "/api/auth/me"},
	} {
		code, body := call(t, h, r.method, r.path, `{"message":"hi","userId":"508794a7-91d8-4ad1-8289-fb733a171e72"}`, "")
		if code != http.StatusUnauthorized || body["success"] != false || body["message"] == "" {
			t.Errorf("%s %s: got %d %v, want 401 JSON error", r.method, r.path, code, body)
		}
	}
}

func TestAuthDisabledWithoutSecret(t *testing.T) {
	h := (&Server{}).Handler()
	code, _ := call(t, h, http.MethodPost, "/api/auth/login", `{"email":"a@b.co","password":"whatever1"}`, "")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("login without JWT secret: got %d, want 503", code)
	}
}

func TestRegisterValidation(t *testing.T) {
	h := newTestServer().Handler()
	cases := map[string]string{
		"missing name":   `{"email":"a@b.co","password":"longenough","fullName":" "}`,
		"bad email":      `{"email":"not-an-email","password":"longenough","fullName":"A"}`,
		"no tld":         `{"email":"a@localhost","password":"longenough","fullName":"A"}`,
		"short password": `{"email":"a@b.co","password":"short","fullName":"A"}`,
		"long password":  `{"email":"a@b.co","password":"` + strings.Repeat("x", 73) + `","fullName":"A"}`,
		"bad json":       `{`,
	}
	for name, body := range cases {
		code, out := call(t, h, http.MethodPost, "/api/auth/register", body, "")
		if code != http.StatusBadRequest || out["message"] == "" {
			t.Errorf("%s: got %d %v, want 400 with message", name, code, out)
		}
	}
	if code, _ := call(t, h, http.MethodGet, "/api/auth/register", "", ""); code != http.StatusMethodNotAllowed {
		t.Errorf("GET register: got %d, want 405", code)
	}
}

func TestIsWeakSecret(t *testing.T) {
	for _, s := range []string{"", "short", "your_super_secret_jwt_key_here_change_me_please", "your-super-secret-key-change-in-production", "your-super-secret-jwt-key-change-in-production"} {
		if !isWeakSecret(s) {
			t.Errorf("isWeakSecret(%q) = false, want true", s)
		}
	}
	if isWeakSecret("k3Jx9vQm2Lr8Tz5Wp1Ny7Hb4Gd6Fs0Ae9Uc") {
		t.Error("random 36-char secret flagged as weak")
	}
}

func TestMakeTitle(t *testing.T) {
	if got := makeTitle("  I feel   stressed  "); got != "I feel stressed" {
		t.Errorf("got %q", got)
	}
	long := makeTitle(strings.Repeat("worried about work ", 10))
	if !strings.HasSuffix(long, "…") || len([]rune(long)) > titleRunes+1 {
		t.Errorf("long title not trimmed: %q", long)
	}
}
