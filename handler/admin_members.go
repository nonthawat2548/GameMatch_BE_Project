package handler

import (
	"errors"
	"gamematch/store"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strconv"
)

// Like MovieHub's AdminMembers: never expose password hashes or token versions.
func (a *API) AdminMembers(w http.ResponseWriter, r *http.Request) {
	items, err := a.S.ListMembers(r.Context())
	if err != nil {
		dbError(w, err)
		return
	}
	send(w, 200, items)
}

// AdminSetRole is protected by Guard(true). It also checks the current DB
// role under row locks, preventing two admins from demoting each other at once.
func (a *API) AdminSetRole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(w, 400, "เลขสมาชิกไม่ถูกต้อง")
		return
	}
	var req struct {
		Role string `json:"role"`
	}
	if !decode(w, r, &req) {
		return
	}
	if req.Role != "member" && req.Role != "admin" {
		fail(w, 400, "บทบาทต้องเป็น member หรือ admin")
		return
	}
	me := current(r).ID
	if me == id {
		fail(w, 400, "ไม่สามารถเปลี่ยนบทบาทของตัวเองได้")
		return
	}
	updated, err := a.S.SetMemberRole(r.Context(), me, id, req.Role)
	switch {
	case errors.Is(err, store.ErrNotAdmin):
		fail(w, 403, "สำหรับผู้ดูแลระบบเท่านั้น")
	case errors.Is(err, pgx.ErrNoRows):
		fail(w, 404, "ไม่พบสมาชิก")
	case err != nil:
		dbError(w, err)
	default:
		send(w, 200, updated)
	}
}
