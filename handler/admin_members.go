package handler

import (
	"net/http"
	"strconv"
)

// Like MovieHub's AdminMembers: never expose password hashes or token versions.
func (a *API) AdminMembers(w http.ResponseWriter, r *http.Request) {
	rows, err := a.S.DB.Query(r.Context(), `SELECT id,email,username,display_name,role FROM gm_members ORDER BY id DESC`)
	if err != nil {
		dbError(w, err)
		return
	}
	defer rows.Close()
	items := []user{}
	for rows.Next() {
		var u user
		if err = rows.Scan(&u.ID, &u.Email, &u.Username, &u.Name, &u.Role); err != nil {
			dbError(w, err)
			return
		}
		items = append(items, u)
	}
	if err = rows.Err(); err != nil {
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
	ctx := r.Context()
	tx, err := a.S.DB.Begin(ctx)
	if err != nil {
		dbError(w, err)
		return
	}
	defer tx.Rollback(ctx)
	// Always lock in the same order. Re-read actor and target in this transaction.
	rows, err := tx.Query(ctx, `SELECT id,role FROM gm_members WHERE id IN ($1,$2) ORDER BY id FOR UPDATE`, me, id)
	if err != nil {
		dbError(w, err)
		return
	}
	roles := map[int64]string{}
	for rows.Next() {
		var memberID int64
		var role string
		if err = rows.Scan(&memberID, &role); err != nil {
			rows.Close()
			dbError(w, err)
			return
		}
		roles[memberID] = role
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		dbError(w, err)
		return
	}
	if roles[me] != "admin" {
		fail(w, 403, "สำหรับผู้ดูแลระบบเท่านั้น")
		return
	}
	if _, ok := roles[id]; !ok {
		fail(w, 404, "ไม่พบสมาชิก")
		return
	}
	var updated user
	// Revoke old sessions only when the role actually changes.
	err = tx.QueryRow(ctx, `UPDATE gm_members SET role=$2,token_version=token_version+CASE WHEN role<>$2 THEN 1 ELSE 0 END WHERE id=$1 RETURNING id,email,username,display_name,role`, id, req.Role).Scan(&updated.ID, &updated.Email, &updated.Username, &updated.Name, &updated.Role)
	if err != nil {
		dbError(w, err)
		return
	}
	if err = tx.Commit(ctx); err != nil {
		dbError(w, err)
		return
	}
	send(w, 200, updated)
}
