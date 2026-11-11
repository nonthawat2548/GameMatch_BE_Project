package tests

import (
	"gamematch/handler"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterRejectsRoleAndMalformedInputBeforeDatabase(t *testing.T) {
	a := &handler.API{}
	for _, body := range []string{
		`{"email":"demo@example.test","password":"Password123!","displayName":"Demo","role":"admin"}`,
		`{"email":"bad","password":"Password123!","displayName":"Demo"}`,
		`{"email":"demo@example.test","password":"short","displayName":"Demo"}`,
		`{"email":"demo@example.test","password":"Password123!","displayName":"` + strings.Repeat("ก", 34) + `"}`,
		`{"email":"demo@example.test","password":"` + strings.Repeat("a", 73) + `","displayName":"Demo"}`,
		`{} {}`, `null`, `[]`, strings.Repeat("x", (1<<20)+1),
	} {
		w := httptest.NewRecorder()
		a.Register(w, httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatalf("wanted 400, got %d", w.Code)
		}
	}
}

func TestLoginRejectsOversizedPasswordBeforeDatabase(t *testing.T) {
	a := &handler.API{}
	w := httptest.NewRecorder()
	a.Login(w, httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"email":"demo@example.test","password":"`+strings.Repeat("a", 73)+`"}`)))
	if w.Code != 401 {
		t.Fatalf("wanted 401, got %d", w.Code)
	}
}
