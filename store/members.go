package store

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"strings"
)

type Member struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Name     string `json:"displayName"`
	Role     string `json:"role"`
}

func (s *Store) MemberByID(ctx context.Context, id int64) (Member, int, error) {
	var m Member
	var version int
	err := s.DB.QueryRow(ctx, `SELECT id,email,username,display_name,role,token_version FROM gm_members WHERE id=$1`, id).Scan(&m.ID, &m.Email, &m.Username, &m.Name, &m.Role, &version)
	return m, version, err
}

func (s *Store) MemberByIdentity(ctx context.Context, identity string) (Member, string, int, error) {
	var m Member
	var hash string
	var version int
	// Email-shaped identities always identify an email; other identities are usernames.
	// Registration and the database constraint keep these namespaces separate.
	query := `SELECT id,email,username,display_name,role,password_hash,token_version FROM gm_members WHERE username=$1`
	if strings.Contains(identity, "@") {
		query = `SELECT id,email,username,display_name,role,password_hash,token_version FROM gm_members WHERE email=$1`
		identity = strings.ToLower(identity)
	}
	err := s.DB.QueryRow(ctx, query, identity).Scan(&m.ID, &m.Email, &m.Username, &m.Name, &m.Role, &hash, &version)
	return m, hash, version, err
}

func (s *Store) CreateMember(ctx context.Context, email, username, hash, name string) (int64, error) {
	var id int64
	err := s.DB.QueryRow(ctx, `INSERT INTO gm_members(email,username,password_hash,display_name) VALUES($1,$2,$3,$4) RETURNING id`, email, username, hash, name).Scan(&id)
	return id, err
}

func (s *Store) RevokeMemberSessions(ctx context.Context, id int64) error {
	_, err := s.DB.Exec(ctx, `UPDATE gm_members SET token_version=token_version+1 WHERE id=$1`, id)
	return err
}

func (s *Store) UpdateMemberName(ctx context.Context, id int64, name string) error {
	_, err := s.DB.Exec(ctx, `UPDATE gm_members SET display_name=$1 WHERE id=$2`, name, id)
	return err
}

func (s *Store) MemberPassword(ctx context.Context, id int64) (string, error) {
	var hash string
	err := s.DB.QueryRow(ctx, `SELECT password_hash FROM gm_members WHERE id=$1`, id).Scan(&hash)
	return hash, err
}

func (s *Store) UpdateMemberPassword(ctx context.Context, id int64, old, hash string) (bool, error) {
	result, err := s.DB.Exec(ctx, `UPDATE gm_members SET password_hash=$1,token_version=token_version+1 WHERE id=$2 AND password_hash=$3`, hash, id, old)
	return result.RowsAffected() == 1, err
}

func (s *Store) ListMembers(ctx context.Context) ([]Member, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,email,username,display_name,role FROM gm_members ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Member{}
	for rows.Next() {
		var m Member
		if err = rows.Scan(&m.ID, &m.Email, &m.Username, &m.Name, &m.Role); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

var ErrNotAdmin = errors.New("actor is not an admin")

func (s *Store) SetMemberRole(ctx context.Context, actorID, id int64, role string) (Member, error) {
	var updated Member
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return updated, err
	}
	defer tx.Rollback(ctx)
	// Lock actor and target in stable order and check the actor again.
	rows, err := tx.Query(ctx, `SELECT id,role FROM gm_members WHERE id IN ($1,$2) ORDER BY id FOR UPDATE`, actorID, id)
	if err != nil {
		return updated, err
	}
	roles := map[int64]string{}
	for rows.Next() {
		var memberID int64
		var value string
		if err = rows.Scan(&memberID, &value); err != nil {
			rows.Close()
			return updated, err
		}
		roles[memberID] = value
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return updated, err
	}
	if roles[actorID] != "admin" {
		return updated, ErrNotAdmin
	}
	if _, ok := roles[id]; !ok {
		return updated, pgx.ErrNoRows
	}
	err = tx.QueryRow(ctx, `UPDATE gm_members SET role=$2,token_version=token_version+CASE WHEN role<>$2 THEN 1 ELSE 0 END WHERE id=$1 RETURNING id,email,username,display_name,role`, id, role).Scan(&updated.ID, &updated.Email, &updated.Username, &updated.Name, &updated.Role)
	if err != nil {
		return updated, err
	}
	return updated, tx.Commit(ctx)
}
