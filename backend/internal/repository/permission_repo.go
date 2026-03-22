package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PermissionRepo struct {
	db *pgxpool.Pool
}

func NewPermissionRepo(db *pgxpool.Pool) *PermissionRepo {
	return &PermissionRepo{db: db}
}

// GetAll returns the full permission matrix: map[role]map[permission_key]bool
func (r *PermissionRepo) GetAll(ctx context.Context) (map[string]map[string]bool, error) {
	rows, err := r.db.Query(ctx, `SELECT role, permission_key, allowed FROM role_permissions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string]bool)
	for rows.Next() {
		var role, key string
		var allowed bool
		if err := rows.Scan(&role, &key, &allowed); err != nil {
			return nil, err
		}
		if result[role] == nil {
			result[role] = make(map[string]bool)
		}
		result[role][key] = allowed
	}
	return result, rows.Err()
}

// GetByRole returns permissions for a specific role
func (r *PermissionRepo) GetByRole(ctx context.Context, role string) (map[string]bool, error) {
	rows, err := r.db.Query(ctx, `SELECT permission_key, allowed FROM role_permissions WHERE role = $1`, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]bool)
	for rows.Next() {
		var key string
		var allowed bool
		if err := rows.Scan(&key, &allowed); err != nil {
			return nil, err
		}
		result[key] = allowed
	}
	return result, rows.Err()
}

// SaveAll replaces all permissions for all roles (transaction)
func (r *PermissionRepo) SaveAll(ctx context.Context, perms map[string]map[string]bool) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions`); err != nil {
		return err
	}

	for role, keys := range perms {
		for key, allowed := range keys {
			if _, err := tx.Exec(ctx,
				`INSERT INTO role_permissions (role, permission_key, allowed) VALUES ($1, $2, $3)`,
				role, key, allowed,
			); err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
