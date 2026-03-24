package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/sangiagao/rice-marketplace/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newJWTManager() *jwtpkg.Manager {
	return jwtpkg.NewManager("test-secret-at-least-32-chars-long", 15*time.Minute, 720*time.Hour)
}

func setupRouter(jm *jwtpkg.Manager, roles ...string) *gin.Engine {
	r := gin.New()
	r.Use(JWTAuth(jm))
	if len(roles) > 0 {
		r.Use(RequireRole(roles...))
	}
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id":   c.GetString("user_id"),
			"user_role": c.GetString("user_role"),
		})
	})
	return r
}

func TestJWTAuth_ValidToken(t *testing.T) {
	jm := newJWTManager()
	pair, _ := jm.GenerateTokenPair("user-123", "member")

	r := setupRouter(jm)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), "user-123")
	assert.Contains(t, w.Body.String(), "member")
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	jm := newJWTManager()
	r := setupRouter(jm)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization")
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	jm := newJWTManager()
	r := setupRouter(jm)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "missing authorization")
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	jm := jwtpkg.NewManager("test-secret-at-least-32-chars-long", -1*time.Second, -1*time.Second)
	pair, _ := jm.GenerateTokenPair("user-123", "member")

	r := setupRouter(newJWTManager())
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
	assert.Contains(t, w.Body.String(), "invalid or expired token")
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	jm := newJWTManager()
	r := setupRouter(jm)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code)
}

func TestRequireRole_Allowed(t *testing.T) {
	jm := newJWTManager()
	pair, _ := jm.GenerateTokenPair("user-123", "admin")

	r := setupRouter(jm, "admin")
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRequireRole_Forbidden(t *testing.T) {
	jm := newJWTManager()
	pair, _ := jm.GenerateTokenPair("user-123", "member")

	r := setupRouter(jm, "admin")
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 403, w.Code)
	assert.Contains(t, w.Body.String(), "insufficient permissions")
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	jm := newJWTManager()
	pair, _ := jm.GenerateTokenPair("user-123", "member")

	r := setupRouter(jm, "member", "admin")
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}
