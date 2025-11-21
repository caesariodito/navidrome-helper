package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Job represents a single import job persisted to storage.
type Job struct {
	ID         string       `json:"id"`
	Status     string       `json:"status"`
	Phase      string       `json:"phase"`
	Message    string       `json:"message"`
	Progress   float64      `json:"progress"`
	Artist     string       `json:"artist"`
	Album      string       `json:"album"`
	CreatedAt  time.Time    `json:"createdAt"`
	UpdatedAt  time.Time    `json:"updatedAt"`
	FinishedAt *time.Time   `json:"finishedAt,omitempty"`
	Items      []JobItem    `json:"items,omitempty"`
	Logs       []JobLogLine `json:"logs,omitempty"`
}

// JobItem records each source item that maps to the job.
type JobItem struct {
	JobID      string    `json:"jobId"`
	SourceID   string    `json:"sourceId"`
	SourceType string    `json:"sourceType"`
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	Album      string    `json:"album"`
	CoverURL   string    `json:"coverUrl"`
	Status     string    `json:"status"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// JobLogLine captures a message tied to a timestamp.
type JobLogLine struct {
	JobID     string    `json:"jobId"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// LibraryEntry represents an album indexed from NAVIDROME_MUSIC_PATH.
type LibraryEntry struct {
	Artist      string    `json:"artist"`
	Album       string    `json:"album"`
	Path        string    `json:"path"`
	TrackCount  int       `json:"trackCount"`
	UpdatedAt   time.Time `json:"updatedAt"`
	artistNorm  string
	albumNorm   string
}

// Store wraps the sqlite database.
type Store struct {
	db *sql.DB
}

// New opens (and creates if missing) the sqlite database file.
func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1) // single writer is enough for this workload
	s := &Store{db: db}
	if err := s.bootstrap(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) bootstrap() error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			status TEXT NOT NULL,
			phase TEXT NOT NULL,
			message TEXT,
			progress REAL NOT NULL DEFAULT 0,
			artist TEXT,
			album TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			finished_at TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS job_items (
			job_id TEXT NOT NULL,
			source_id TEXT NOT NULL,
			source_type TEXT NOT NULL,
			title TEXT,
			artist TEXT,
			album TEXT,
			cover_url TEXT,
			status TEXT NOT NULL,
			message TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY (job_id, source_id),
			FOREIGN KEY(job_id) REFERENCES jobs(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS job_logs (
			job_id TEXT NOT NULL,
			message TEXT NOT NULL,
			created_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS library_index (
			artist TEXT NOT NULL,
			album TEXT NOT NULL,
			path TEXT NOT NULL,
			track_count INTEGER NOT NULL,
			updated_at TEXT NOT NULL,
			artist_norm TEXT NOT NULL,
			album_norm TEXT NOT NULL,
			PRIMARY KEY (artist_norm, album_norm)
		);`,
	}
	for _, q := range schemas {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("bootstrap schema: %w", err)
		}
	}
	return nil
}

// InsertJob writes a job and its items in a single transaction.
func (s *Store) InsertJob(job *Job) error {
	now := time.Now().UTC()
	job.CreatedAt = now
	job.UpdatedAt = now
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO jobs (id, status, phase, message, progress, artist, album, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.Status, job.Phase, job.Message, job.Progress, job.Artist, job.Album, job.CreatedAt.Format(time.RFC3339Nano), job.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	for _, item := range job.Items {
		item.CreatedAt = now
		item.UpdatedAt = now
		if _, err := tx.Exec(`INSERT INTO job_items (job_id, source_id, source_type, title, artist, album, cover_url, status, message, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			job.ID, item.SourceID, item.SourceType, item.Title, item.Artist, item.Album, item.CoverURL, item.Status, item.Message, item.CreatedAt.Format(time.RFC3339Nano), item.UpdatedAt.Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("insert job item: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// UpdateJobState sets the job status/phase/message/progress and optionally finished_at.
func (s *Store) UpdateJobState(id, status, phase, message string, progress float64, finished bool) error {
	now := time.Now().UTC()
	var err error
	if finished {
		_, err = s.db.Exec(`UPDATE jobs SET status=?, phase=?, message=?, progress=?, finished_at=?, updated_at=? WHERE id=?`,
			status, phase, message, progress, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), id)
	} else {
		_, err = s.db.Exec(`UPDATE jobs SET status=?, phase=?, message=?, progress=?, updated_at=? WHERE id=?`,
			status, phase, message, progress, now.Format(time.RFC3339Nano), id)
	}
	if err != nil {
		return fmt.Errorf("update job state: %w", err)
	}
	return nil
}

// AddJobLog appends a log line for a job.
func (s *Store) AddJobLog(jobID, message string) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(`INSERT INTO job_logs (job_id, message, created_at) VALUES (?, ?, ?)`, jobID, message, now.Format(time.RFC3339Nano))
	return err
}

// UpdateJobItem updates a single job item status and message.
func (s *Store) UpdateJobItem(jobID, sourceID, status, message string) error {
	now := time.Now().UTC()
	if _, err := s.db.Exec(`UPDATE job_items SET status=?, message=?, updated_at=? WHERE job_id=? AND source_id=?`,
		status, message, now.Format(time.RFC3339Nano), jobID, sourceID); err != nil {
		return fmt.Errorf("update job item: %w", err)
	}
	return nil
}

// ListJobs returns latest jobs up to limit.
func (s *Store) ListJobs(limit int) ([]Job, error) {
	rows, err := s.db.Query(`SELECT id, status, phase, message, progress, artist, album, created_at, updated_at, finished_at FROM jobs ORDER BY datetime(created_at) DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []Job
	for rows.Next() {
		var job Job
		var createdAt, updatedAt, finishedAt sql.NullString
		if err := rows.Scan(&job.ID, &job.Status, &job.Phase, &job.Message, &job.Progress, &job.Artist, &job.Album, &createdAt, &updatedAt, &finishedAt); err != nil {
			return nil, err
		}
		job.CreatedAt = parseTime(createdAt)
		job.UpdatedAt = parseTime(updatedAt)
		if finishedAt.Valid {
			t := parseTime(finishedAt)
			job.FinishedAt = &t
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}

// GetJob fetches a job by id including items and logs.
func (s *Store) GetJob(id string) (*Job, error) {
	row := s.db.QueryRow(`SELECT id, status, phase, message, progress, artist, album, created_at, updated_at, finished_at FROM jobs WHERE id=?`, id)
	var job Job
	var createdAt, updatedAt, finishedAt sql.NullString
	if err := row.Scan(&job.ID, &job.Status, &job.Phase, &job.Message, &job.Progress, &job.Artist, &job.Album, &createdAt, &updatedAt, &finishedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	job.CreatedAt = parseTime(createdAt)
	job.UpdatedAt = parseTime(updatedAt)
	if finishedAt.Valid {
		t := parseTime(finishedAt)
		job.FinishedAt = &t
	}

	items, err := s.loadItems(id)
	if err != nil {
		return nil, err
	}
	logs, err := s.loadLogs(id)
	if err != nil {
		return nil, err
	}
	job.Items = items
	job.Logs = logs
	return &job, nil
}

func (s *Store) loadItems(jobID string) ([]JobItem, error) {
	rows, err := s.db.Query(`SELECT job_id, source_id, source_type, title, artist, album, cover_url, status, message, created_at, updated_at FROM job_items WHERE job_id=?`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []JobItem
	for rows.Next() {
		var it JobItem
		var createdAt, updatedAt string
		if err := rows.Scan(&it.JobID, &it.SourceID, &it.SourceType, &it.Title, &it.Artist, &it.Album, &it.CoverURL, &it.Status, &it.Message, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		it.CreatedAt = parseTimeString(createdAt)
		it.UpdatedAt = parseTimeString(updatedAt)
		items = append(items, it)
	}
	return items, nil
}

func (s *Store) loadLogs(jobID string) ([]JobLogLine, error) {
	rows, err := s.db.Query(`SELECT job_id, message, created_at FROM job_logs WHERE job_id=? ORDER BY datetime(created_at) ASC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []JobLogLine
	for rows.Next() {
		var line JobLogLine
		var createdAt string
		if err := rows.Scan(&line.JobID, &line.Message, &createdAt); err != nil {
			return nil, err
		}
		line.CreatedAt = parseTimeString(createdAt)
		logs = append(logs, line)
	}
	return logs, nil
}

// ReplaceLibraryIndex replaces the entire library_index table with the provided entries.
func (s *Store) ReplaceLibraryIndex(entries []LibraryEntry) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM library_index`); err != nil {
		return fmt.Errorf("clear library_index: %w", err)
	}
	stmt, err := tx.Prepare(`INSERT INTO library_index (artist, album, path, track_count, updated_at, artist_norm, album_norm) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert library_index: %w", err)
	}
	defer stmt.Close()
	for _, e := range entries {
		if _, err := stmt.Exec(e.Artist, e.Album, e.Path, e.TrackCount, e.UpdatedAt.Format(time.RFC3339Nano), e.artistNorm, e.albumNorm); err != nil {
			return fmt.Errorf("insert library_index: %w", err)
		}
	}
	return tx.Commit()
}

// ListLibrary returns all library entries.
func (s *Store) ListLibrary() ([]LibraryEntry, error) {
	rows, err := s.db.Query(`SELECT artist, album, path, track_count, updated_at, artist_norm, album_norm FROM library_index ORDER BY artist_norm, album_norm`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []LibraryEntry
	for rows.Next() {
		var e LibraryEntry
		var updatedAt string
		if err := rows.Scan(&e.Artist, &e.Album, &e.Path, &e.TrackCount, &updatedAt, &e.artistNorm, &e.albumNorm); err != nil {
			return nil, err
		}
		e.UpdatedAt = parseTimeString(updatedAt)
		out = append(out, e)
	}
	return out, nil
}

// LibraryExists reports whether a given normalized artist/album is in the index.
func (s *Store) LibraryExists(artistNorm, albumNorm string) (bool, error) {
	row := s.db.QueryRow(`SELECT 1 FROM library_index WHERE artist_norm=? AND album_norm=? LIMIT 1`, artistNorm, albumNorm)
	var dummy int
	if err := row.Scan(&dummy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func parseTime(ns sql.NullString) time.Time {
	if !ns.Valid {
		return time.Time{}
	}
	return parseTimeString(ns.String)
}

func parseTimeString(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
