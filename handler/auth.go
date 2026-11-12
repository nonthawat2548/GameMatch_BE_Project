package handler

import (
	"github.com/jackc/pgx/v5"
	"errors"
	"gamematch/auth"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"
)

func (a *API) Register(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
		Name     string `json:"displayName"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Email = strings.ToLower(strings.TrimSpace(b.Email))
	b.Username = strings.TrimSpace(b.Username)
	b.Name = strings.TrimSpace(b.Name)
	if b.Username == "" && b.Email != "" {
		b.Username = b.Email
	}
	address, e := mail.ParseAddress(b.Email)
	if e != nil || address.Address != b.Email || !strings.Contains(b.Email, "@") || len(b.Email) > 255 || strings.HasSuffix(b.Email, "@local.invalid") {
		fail(w, 400, "กรุณากรอกอีเมลให้ถูกต้อง")
		return
	}
	if b.Username == "" || len(b.Username) > 255 {
		fail(w, 400, "ชื่อผู้ใช้ไม่ถูกต้อง")
		return
	}
	if b.Name == "" || utf8.RuneCountInString(b.Name) > 50 {
		fail(w, 400, "กรุณาระบุชื่อที่แสดง ไม่เกิน 50 ตัวอักษร")
		return
	}
	if !auth.ValidPassword(b.Password) {
		fail(w, 400, "รหัสผ่านต้องมีอย่างน้อย 8 ตัวอักษร และไม่เกิน 72 ไบต์")
		return
	}
	hash, e := auth.HashPassword(b.Password)
	if e != nil {
		fail(w, 500, "สร้างรหัสผ่านไม่สำเร็จ")
		return
	}
	id, e := a.S.CreateMember(r.Context(), b.Email, b.Username, hash, b.Name)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 201, map[string]any{"id": id})
}
func (a *API) Login(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decode(w, r, &b) {
		return
	}
	identity := strings.TrimSpace(b.Username)
	if identity == "" {
		identity = strings.TrimSpace(b.Email)
	}
	if len(identity) > 320 || len(b.Password) > 72 {
		fail(w, 401, "อีเมล ชื่อผู้ใช้ หรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	u, hash, version, e := a.S.MemberByIdentity(r.Context(), identity)
	if e != nil && !errors.Is(e, pgx.ErrNoRows) {
		dbError(w, e)
		return
	}
	if errors.Is(e, pgx.ErrNoRows) {
		auth.CheckPassword(auth.DummyHash, b.Password)
		fail(w, 401, "อีเมล ชื่อผู้ใช้ หรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	if !auth.CheckPassword(hash, b.Password) {
		fail(w, 401, "อีเมล ชื่อผู้ใช้ หรือรหัสผ่านไม่ถูกต้อง")
		return
	}
	token, e := auth.Issue(a.Secret, u.ID, version)
	if e != nil {
		fail(w, 500, "สร้าง session ไม่สำเร็จ")
		return
	}
	send(w, 200, map[string]any{"token": token, "user": u})
}
func (a *API) Logout(w http.ResponseWriter, r *http.Request) {
	e := a.S.RevokeMemberSessions(r.Context(), current(r).ID)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Profile(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Name string `json:"displayName"`
	}
	if !decode(w, r, &b) {
		return
	}
	b.Name = strings.TrimSpace(b.Name)
	if b.Name == "" || utf8.RuneCountInString(b.Name) > 50 {
		fail(w, 400, "กรุณาระบุชื่อที่แสดง ไม่เกิน 50 ตัวอักษร")
		return
	}
	e := a.S.UpdateMemberName(r.Context(), current(r).ID, b.Name)
	if e != nil {
		dbError(w, e)
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
func (a *API) Password(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Current string `json:"currentPassword"`
		New     string `json:"newPassword"`
	}
	if !decode(w, r, &b) {
		return
	}
	if !auth.ValidPassword(b.New) {
		fail(w, 400, "รหัสผ่านใหม่ต้องมีอย่างน้อย 8 ตัวอักษร และไม่เกิน 72 ไบต์")
		return
	}
	old, e := a.S.MemberPassword(r.Context(), current(r).ID)
	if e != nil {
		dbError(w, e)
		return
	}
	if !auth.CheckPassword(old, b.Current) {
		fail(w, 400, "รหัสผ่านปัจจุบันไม่ถูกต้อง")
		return
	}
	hash, e := auth.HashPassword(b.New)
	if e != nil {
		fail(w, 500, "สร้างรหัสผ่านไม่สำเร็จ")
		return
	}
	updated, e := a.S.UpdateMemberPassword(r.Context(), current(r).ID, old, hash)
	if e != nil {
		dbError(w, e)
		return
	}
	if !updated {
		fail(w, 409, "ข้อมูลเปลี่ยนแล้ว กรุณาลองใหม่")
		return
	}
	send(w, 200, map[string]bool{"ok": true})
}
