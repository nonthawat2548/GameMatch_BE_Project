package tests

import (
	"gamematch/store"
	"testing"
)

func TestRegistrationIdentityRulesAndLogin(t *testing.T) {
	_, server := databaseAPI(t)
	const password = "Password123!"
	register := func(email, username string, status int) int64 {
		t.Helper()
		body := map[string]string{"email": email, "password": password, "displayName": "Identity test member"}
		if username != "" {
			body["username"] = username
		}
		var created struct {
			ID int64 `json:"id"`
		}
		requestJSON(t, server, "POST", "/api/auth/register", "", body, status, &created)
		if status == 201 && created.ID <= 0 {
			t.Fatal("registration did not return a member ID")
		}
		return created.ID
	}
	login := func(field, identity string, memberID int64, username string) {
		t.Helper()
		var session struct {
			Token string       `json:"token"`
			User  store.Member `json:"user"`
		}
		requestJSON(t, server, "POST", "/api/auth/login", "", map[string]string{field: identity, "password": password}, 200, &session)
		if session.Token == "" || session.User.ID != memberID || session.User.Username != username {
			t.Fatal("login returned the wrong identity or no token")
		}
		var current store.Member
		requestJSON(t, server, "GET", "/api/me", session.Token, nil, 200, &current)
		if current.ID != memberID || current.Username != username {
			t.Fatal("authenticated member differs from the registered identity")
		}
	}

	normalID := register("normal@example.test", "", 201)
	login("email", "normal@example.test", normalID, "normal@example.test")

	plainID := register("plain@example.test", "LegacyUser", 201)
	login("email", "LegacyUser", plainID, "LegacyUser")
	login("username", "LegacyUser", plainID, "LegacyUser")
	requestJSON(t, server, "POST", "/api/auth/login", "", map[string]string{"username": "legacyuser", "password": password}, 401, nil)

	caseID := register("Mixed@EXAMPLE.TEST", "MIXED@example.test", 201)
	login("email", "MIXED@EXAMPLE.TEST", caseID, "mixed@example.test")

	register("attacker@example.test", "victim@example.test", 400)
	victimID := register("victim@example.test", "", 201)
	login("email", "victim@example.test", victimID, "victim@example.test")

	register("duplicate@example.test", "LegacyUser", 409)
}
