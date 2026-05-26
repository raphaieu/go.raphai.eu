package store

import (
	"database/sql"
	"time"

	"go.raphai.eu/models"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS redirects (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			slug         TEXT UNIQUE NOT NULL,
			target_url   TEXT NOT NULL,
			is_active    INTEGER DEFAULT 1,
			redirect_type TEXT DEFAULT 'meta_refresh',
			utm_source   TEXT DEFAULT '',
			utm_medium   TEXT DEFAULT '',
			utm_campaign TEXT DEFAULT '',
			utm_term     TEXT DEFAULT '',
			utm_content  TEXT DEFAULT '',
			visit_count  INTEGER DEFAULT 0,
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func scanRedirect(scanner interface {
	Scan(dest ...interface{}) error
}) (*models.Redirect, error) {
	r := &models.Redirect{}
	var isActive int
	var visitCount int64
	var createdAt, updatedAt string
	err := scanner.Scan(
		&r.ID, &r.Slug, &r.TargetURL, &isActive, &r.RedirectType,
		&r.UTMSource, &r.UTMMedium, &r.UTMCampaign, &r.UTMTerm, &r.UTMContent,
		&visitCount, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, err
	}
	r.IsActive = isActive == 1
	r.VisitCount = visitCount
	r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	r.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt)
	return r, nil
}

func (s *Store) CreateRedirect(r *models.Redirect) (int64, error) {
	result, err := s.db.Exec(`
		INSERT INTO redirects (slug, target_url, is_active, redirect_type,
			utm_source, utm_medium, utm_campaign, utm_term, utm_content)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, r.Slug, r.TargetURL, boolInt(r.IsActive), r.RedirectType,
		r.UTMSource, r.UTMMedium, r.UTMCampaign, r.UTMTerm, r.UTMContent)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *Store) GetRedirectBySlug(slug string) (*models.Redirect, error) {
	row := s.db.QueryRow(`
		SELECT id, slug, target_url, is_active, redirect_type,
			utm_source, utm_medium, utm_campaign, utm_term, utm_content,
			visit_count, created_at, updated_at
		FROM redirects WHERE slug = ?
	`, slug)
	return scanRedirect(row)
}

func (s *Store) GetRedirectByID(id int64) (*models.Redirect, error) {
	row := s.db.QueryRow(`
		SELECT id, slug, target_url, is_active, redirect_type,
			utm_source, utm_medium, utm_campaign, utm_term, utm_content,
			visit_count, created_at, updated_at
		FROM redirects WHERE id = ?
	`, id)
	return scanRedirect(row)
}

func (s *Store) ListRedirects() ([]*models.Redirect, error) {
	rows, err := s.db.Query(`
		SELECT id, slug, target_url, is_active, redirect_type,
			utm_source, utm_medium, utm_campaign, utm_term, utm_content,
			visit_count, created_at, updated_at
		FROM redirects ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var redirects []*models.Redirect
	for rows.Next() {
		r, err := scanRedirect(rows)
		if err != nil {
			return nil, err
		}
		redirects = append(redirects, r)
	}
	return redirects, rows.Err()
}

func (s *Store) UpdateRedirect(r *models.Redirect) error {
	_, err := s.db.Exec(`
		UPDATE redirects SET slug = ?, target_url = ?, is_active = ?,
			redirect_type = ?, utm_source = ?, utm_medium = ?,
			utm_campaign = ?, utm_term = ?, utm_content = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, r.Slug, r.TargetURL, boolInt(r.IsActive), r.RedirectType,
		r.UTMSource, r.UTMMedium, r.UTMCampaign, r.UTMTerm, r.UTMContent,
		r.ID)
	return err
}

func (s *Store) DeleteRedirect(id int64) error {
	_, err := s.db.Exec("DELETE FROM redirects WHERE id = ?", id)
	return err
}

func (s *Store) ToggleRedirect(id int64) error {
	_, err := s.db.Exec(`UPDATE redirects SET is_active = CASE WHEN is_active = 1 THEN 0 ELSE 1 END, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}

func (s *Store) IncrementVisitCount(id int64) error {
	_, err := s.db.Exec("UPDATE redirects SET visit_count = visit_count + 1 WHERE id = ?", id)
	return err
}

func (s *Store) SlugExists(slug string, excludeID int64) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM redirects WHERE slug = ? AND id != ?", slug, excludeID).Scan(&count)
	return count > 0, err
}
