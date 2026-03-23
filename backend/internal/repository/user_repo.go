package repository

import (
	"context"
	"errors"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sangiagao/rice-marketplace/internal/model"
	"github.com/sangiagao/rice-marketplace/pkg/crypto"
)

var ErrUserNotFound = errors.New("user not found")

var phoneSearchRegex = regexp.MustCompile(`^0\d{9}$`)

type UserRepo struct {
	pool   *pgxpool.Pool
	crypto *crypto.PhoneCrypto
}

func NewUserRepo(pool *pgxpool.Pool, phoneCrypto *crypto.PhoneCrypto) *UserRepo {
	return &UserRepo{pool: pool, crypto: phoneCrypto}
}

func (r *UserRepo) scanUser(row pgx.Row) (*model.User, error) {
	var u model.User
	var discardDistrict *string // DB column exists but is no longer used
	var phoneEncrypt *string
	err := row.Scan(
		&u.ID, &u.Phone, &u.Role, &u.Name, &u.AvatarURL,
		&u.Address, &u.Province, &discardDistrict, &u.Ward, &u.Description, &u.OrgName,
		&u.IsBlocked, &u.BlockReason, &u.AcceptedTOSAt,
		&u.CreatedAt, &u.UpdatedAt,
		&phoneEncrypt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	// Decrypt phone from phone_encrypt if available
	if phoneEncrypt != nil && *phoneEncrypt != "" {
		decrypted, decErr := r.crypto.Decrypt(*phoneEncrypt)
		if decErr == nil {
			u.Phone = decrypted
		}
		// If decryption fails, keep the phone value from the phone column
	}
	return &u, nil
}

const userColumns = `id, phone, role, name, avatar_url, address, province, district, ward,
	description, org_name, is_blocked, block_reason,
	accepted_tos_at, created_at, updated_at, phone_encrypt`

func (r *UserRepo) Create(ctx context.Context, phone, role string) (*model.User, error) {
	phoneHash := r.crypto.Hash(phone)
	phoneEnc, err := r.crypto.Encrypt(phone)
	if err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (phone, role, phone_hash, phone_encrypt) VALUES ($1, $2, $3, $4)
		 RETURNING `+userColumns,
		phone, role, phoneHash, phoneEnc,
	)
	return r.scanUser(row)
}

func (r *UserRepo) CreateWithPassword(ctx context.Context, phone, name, passwordHash, province, ward, address string) (*model.User, error) {
	phoneHash := r.crypto.Hash(phone)
	phoneEnc, err := r.crypto.Encrypt(phone)
	if err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (phone, role, name, password_hash, province, ward, address, accepted_tos_at, phone_hash, phone_encrypt)
		 VALUES ($1, 'member', $2, $3, $4, $5, $6, NOW(), $7, $8)
		 RETURNING `+userColumns,
		phone, name, passwordHash, province, ward, address, phoneHash, phoneEnc,
	)
	return r.scanUser(row)
}

func (r *UserRepo) GetPasswordHash(ctx context.Context, phone string) (string, error) {
	phoneHash := r.crypto.Hash(phone)
	var hash *string
	err := r.pool.QueryRow(ctx,
		`SELECT password_hash FROM users WHERE phone_hash = $1`, phoneHash,
	).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		// Fallback: try plaintext phone for unmigrated data
		err = r.pool.QueryRow(ctx,
			`SELECT password_hash FROM users WHERE phone = $1 AND phone_hash IS NULL`, phone,
		).Scan(&hash)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	if hash == nil {
		return "", errors.New("no password set")
	}
	return *hash, nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	phoneHash := r.crypto.Hash(phone)
	row := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE phone_hash = $1`, phoneHash,
	)
	u, err := r.scanUser(row)
	if err == ErrUserNotFound {
		// Fallback: try plaintext phone for unmigrated data
		row = r.pool.QueryRow(ctx,
			`SELECT `+userColumns+` FROM users WHERE phone = $1 AND phone_hash IS NULL`, phone,
		)
		return r.scanUser(row)
	}
	return u, err
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	var discardDistrict *string
	var phoneEncrypt *string
	err := r.pool.QueryRow(ctx,
		`SELECT `+userColumns+`,
			(SELECT MAX(expires_at) FROM subscriptions WHERE user_id = users.id AND status = 'active' AND expires_at > NOW())
		 FROM users WHERE id = $1`, id,
	).Scan(
		&u.ID, &u.Phone, &u.Role, &u.Name, &u.AvatarURL,
		&u.Address, &u.Province, &discardDistrict, &u.Ward, &u.Description, &u.OrgName,
		&u.IsBlocked, &u.BlockReason, &u.AcceptedTOSAt,
		&u.CreatedAt, &u.UpdatedAt,
		&phoneEncrypt,
		&u.SubscriptionExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	if phoneEncrypt != nil && *phoneEncrypt != "" {
		decrypted, decErr := r.crypto.Decrypt(*phoneEncrypt)
		if decErr == nil {
			u.Phone = decrypted
		}
	}
	return &u, nil
}

func (r *UserRepo) UpdateProfile(ctx context.Context, id string, req *model.UpdateProfileRequest) (*model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET
			name = COALESCE($2, name),
			address = COALESCE($3, address),
			province = COALESCE($4, province),
			ward = COALESCE($5, ward),
			description = COALESCE($6, description),
			org_name = COALESCE($7, org_name)
		 WHERE id = $1
		 RETURNING `+userColumns,
		id, req.Name, req.Address, req.Province, req.Ward, req.Description, req.OrgName,
	)
	return r.scanUser(row)
}

func (r *UserRepo) SetRole(ctx context.Context, id, role string) (*model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET role = $2 WHERE id = $1 RETURNING `+userColumns,
		id, role,
	)
	return r.scanUser(row)
}

func (r *UserRepo) AcceptTOS(ctx context.Context, id string) (*model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET accepted_tos_at = NOW() WHERE id = $1 RETURNING `+userColumns,
		id,
	)
	return r.scanUser(row)
}

func (r *UserRepo) UpdateAvatar(ctx context.Context, id, avatarURL string) (*model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET avatar_url = $2 WHERE id = $1 RETURNING `+userColumns,
		id, avatarURL,
	)
	return r.scanUser(row)
}

func (r *UserRepo) UpdatePassword(ctx context.Context, phone, passwordHash string) error {
	phoneHash := r.crypto.Hash(phone)
	result, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = NOW() WHERE phone_hash = $1`,
		phoneHash, passwordHash,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		// Fallback: try plaintext phone for unmigrated data
		_, err = r.pool.Exec(ctx,
			`UPDATE users SET password_hash = $2, updated_at = NOW() WHERE phone = $1 AND phone_hash IS NULL`,
			phone, passwordHash,
		)
		return err
	}
	return nil
}

func (r *UserRepo) GetPasswordHashByID(ctx context.Context, userID string) (string, error) {
	var hash *string
	err := r.pool.QueryRow(ctx,
		`SELECT password_hash FROM users WHERE id = $1`, userID,
	).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	if hash == nil || *hash == "" {
		return "", nil
	}
	return *hash, nil
}

func (r *UserRepo) UpdatePasswordByID(ctx context.Context, userID, passwordHash string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`,
		userID, passwordHash,
	)
	return err
}

func (r *UserRepo) UpdatePhone(ctx context.Context, userID, newPhone string) (*model.User, error) {
	phoneHash := r.crypto.Hash(newPhone)
	phoneEnc, err := r.crypto.Encrypt(newPhone)
	if err != nil {
		return nil, err
	}
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET phone = $2, phone_hash = $3, phone_encrypt = $4, updated_at = NOW()
		 WHERE id = $1 RETURNING `+userColumns,
		userID, newPhone, phoneHash, phoneEnc,
	)
	return r.scanUser(row)
}

func (r *UserRepo) PhoneExists(ctx context.Context, phone string) (bool, error) {
	phoneHash := r.crypto.Hash(phone)
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM users WHERE phone_hash = $1)`, phoneHash,
	).Scan(&exists)
	if err != nil {
		// Fallback: try plaintext phone
		err = r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1 AND phone_hash IS NULL)`, phone,
		).Scan(&exists)
	}
	return exists, err
}

// --- Admin methods ---

func (r *UserRepo) BlockUser(ctx context.Context, id, reason string) (*model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET is_blocked = true, block_reason = $2
		 WHERE id = $1 RETURNING `+userColumns,
		id, reason,
	)
	return r.scanUser(row)
}

func (r *UserRepo) DeleteUser(ctx context.Context, id string) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *UserRepo) UnblockUser(ctx context.Context, id string) (*model.User, error) {
	row := r.pool.QueryRow(ctx,
		`UPDATE users SET is_blocked = false, block_reason = NULL
		 WHERE id = $1 RETURNING `+userColumns,
		id,
	)
	return r.scanUser(row)
}

func (r *UserRepo) ListUsers(ctx context.Context, search string, page, limit int) ([]*model.User, int, error) {
	offset := (page - 1) * limit

	var total int
	var countQuery string
	var countArgs []interface{}

	if search != "" {
		if phoneSearchRegex.MatchString(search) {
			// Exact phone search: hash and match
			phoneHash := r.crypto.Hash(search)
			countQuery = `SELECT COUNT(*) FROM users WHERE phone_hash = $1`
			countArgs = []interface{}{phoneHash}
		} else {
			// Name search only (can't LIKE on encrypted phone)
			countQuery = `SELECT COUNT(*) FROM users WHERE name ILIKE $1`
			countArgs = []interface{}{"%" + search + "%"}
		}
	} else {
		countQuery = `SELECT COUNT(*) FROM users`
	}

	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	var dataQuery string
	var dataArgs []interface{}

	subExpiryJoin := ` LEFT JOIN (
		SELECT user_id, MAX(expires_at) AS sub_expires_at
		FROM subscriptions WHERE status = 'active' AND expires_at > NOW()
		GROUP BY user_id
	) sub ON sub.user_id = users.id`

	if search != "" {
		if phoneSearchRegex.MatchString(search) {
			phoneHash := r.crypto.Hash(search)
			dataQuery = `SELECT ` + userColumns + `, sub.sub_expires_at FROM users` + subExpiryJoin + `
				WHERE phone_hash = $1
				ORDER BY users.created_at DESC LIMIT $2 OFFSET $3`
			dataArgs = []interface{}{phoneHash, limit, offset}
		} else {
			dataQuery = `SELECT ` + userColumns + `, sub.sub_expires_at FROM users` + subExpiryJoin + `
				WHERE name ILIKE $1
				ORDER BY users.created_at DESC LIMIT $2 OFFSET $3`
			dataArgs = []interface{}{"%" + search + "%", limit, offset}
		}
	} else {
		dataQuery = `SELECT ` + userColumns + `, sub.sub_expires_at FROM users` + subExpiryJoin + `
			ORDER BY users.created_at DESC LIMIT $1 OFFSET $2`
		dataArgs = []interface{}{limit, offset}
	}

	rows, err := r.pool.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		var u model.User
		var discardDistrict *string
		var phoneEncrypt *string
		if err := rows.Scan(
			&u.ID, &u.Phone, &u.Role, &u.Name, &u.AvatarURL,
			&u.Address, &u.Province, &discardDistrict, &u.Ward, &u.Description, &u.OrgName,
			&u.IsBlocked, &u.BlockReason, &u.AcceptedTOSAt,
			&u.CreatedAt, &u.UpdatedAt,
			&phoneEncrypt,
			&u.SubscriptionExpiresAt,
		); err != nil {
			return nil, 0, err
		}
		if phoneEncrypt != nil && *phoneEncrypt != "" {
			decrypted, decErr := r.crypto.Decrypt(*phoneEncrypt)
			if decErr == nil {
				u.Phone = decrypted
			}
		}
		users = append(users, &u)
	}
	if users == nil {
		users = []*model.User{}
	}
	return users, total, rows.Err()
}

func (r *UserRepo) GetDashboardStats(ctx context.Context) (map[string]int, error) {
	stats := make(map[string]int)

	queries := map[string]string{
		"total_users":          `SELECT COUNT(*) FROM users`,
		"total_listings":       `SELECT COUNT(*) FROM listings WHERE status != 'deleted'`,
		"active_listings":      `SELECT COUNT(*) FROM listings WHERE status = 'active'`,
		"active_subscriptions": `SELECT COUNT(*) FROM subscriptions WHERE status = 'active' AND expires_at > NOW()`,
		"pending_reports":      `SELECT COUNT(*) FROM reports WHERE status = 'pending'`,
		"total_ratings":        `SELECT COUNT(*) FROM ratings`,
	}

	for key, q := range queries {
		var count int
		if err := r.pool.QueryRow(ctx, q).Scan(&count); err != nil {
			return nil, err
		}
		stats[key] = count
	}
	return stats, nil
}

type MonthCount struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type LabelCount struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type DashboardCharts struct {
	UsersByMonth       []MonthCount `json:"users_by_month"`
	ListingsByMonth    []MonthCount `json:"listings_by_month"`
	SubsByMonth        []MonthCount `json:"subs_by_month"`
	UsersByRole        []LabelCount `json:"users_by_role"`
	ListingsByRiceType []LabelCount `json:"listings_by_rice_type"`
	ListingsByProvince []LabelCount `json:"listings_by_province"`
}

func (r *UserRepo) GetDashboardCharts(ctx context.Context) (*DashboardCharts, error) {
	charts := &DashboardCharts{}

	// Single CTE query combining all 6 chart queries into one DB round-trip
	const cteQuery = `
	WITH date_cutoff AS (SELECT date_trunc('month', NOW()) - INTERVAL '5 months' AS since),
	users_by_month AS (
		SELECT 'ubm' AS src, TO_CHAR(date_trunc('month', created_at), 'MM/YYYY') AS label, COUNT(*) AS cnt
		FROM users, date_cutoff WHERE created_at >= date_cutoff.since
		GROUP BY date_trunc('month', created_at) ORDER BY date_trunc('month', created_at)
	),
	listings_by_month AS (
		SELECT 'lbm' AS src, TO_CHAR(date_trunc('month', created_at), 'MM/YYYY') AS label, COUNT(*) AS cnt
		FROM listings, date_cutoff WHERE created_at >= date_cutoff.since
		GROUP BY date_trunc('month', created_at) ORDER BY date_trunc('month', created_at)
	),
	subs_by_month AS (
		SELECT 'sbm' AS src, TO_CHAR(date_trunc('month', created_at), 'MM/YYYY') AS label, COUNT(*) AS cnt
		FROM subscriptions, date_cutoff WHERE created_at >= date_cutoff.since
		GROUP BY date_trunc('month', created_at) ORDER BY date_trunc('month', created_at)
	),
	users_by_role AS (
		SELECT 'ubr' AS src, role AS label, COUNT(*) AS cnt FROM users GROUP BY role ORDER BY cnt DESC
	),
	listings_by_type AS (
		SELECT 'lbt' AS src, rice_type AS label, COUNT(*) AS cnt
		FROM listings WHERE status != 'deleted' GROUP BY rice_type ORDER BY cnt DESC LIMIT 8
	),
	listings_by_province AS (
		SELECT 'lbp' AS src, province AS label, COUNT(*) AS cnt
		FROM listings WHERE province != '' AND status != 'deleted' GROUP BY province ORDER BY cnt DESC LIMIT 8
	)
	SELECT * FROM users_by_month
	UNION ALL SELECT * FROM listings_by_month
	UNION ALL SELECT * FROM subs_by_month
	UNION ALL SELECT * FROM users_by_role
	UNION ALL SELECT * FROM listings_by_type
	UNION ALL SELECT * FROM listings_by_province`

	rows, err := r.pool.Query(ctx, cteQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var src, label string
		var cnt int
		if err := rows.Scan(&src, &label, &cnt); err != nil {
			return nil, err
		}
		switch src {
		case "ubm":
			charts.UsersByMonth = append(charts.UsersByMonth, MonthCount{Month: label, Count: cnt})
		case "lbm":
			charts.ListingsByMonth = append(charts.ListingsByMonth, MonthCount{Month: label, Count: cnt})
		case "sbm":
			charts.SubsByMonth = append(charts.SubsByMonth, MonthCount{Month: label, Count: cnt})
		case "ubr":
			charts.UsersByRole = append(charts.UsersByRole, LabelCount{Label: label, Count: cnt})
		case "lbt":
			charts.ListingsByRiceType = append(charts.ListingsByRiceType, LabelCount{Label: label, Count: cnt})
		case "lbp":
			charts.ListingsByProvince = append(charts.ListingsByProvince, LabelCount{Label: label, Count: cnt})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Ensure non-nil slices
	if charts.UsersByMonth == nil { charts.UsersByMonth = []MonthCount{} }
	if charts.ListingsByMonth == nil { charts.ListingsByMonth = []MonthCount{} }
	if charts.SubsByMonth == nil { charts.SubsByMonth = []MonthCount{} }
	if charts.UsersByRole == nil { charts.UsersByRole = []LabelCount{} }
	if charts.ListingsByRiceType == nil { charts.ListingsByRiceType = []LabelCount{} }
	if charts.ListingsByProvince == nil { charts.ListingsByProvince = []LabelCount{} }

	return charts, nil
}
