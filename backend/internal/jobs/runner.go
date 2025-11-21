package jobs

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"navidrome-helper/internal/config"
	"navidrome-helper/internal/store"
)

const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"

	PhaseQueued          = "queued"
	PhaseFetchingSource  = "fetching_source"
	PhaseDownloading     = "downloading"
	PhaseExtracting      = "extracting"
	PhasePlacing         = "placing"
	PhaseCleanup         = "cleanup"
	PhaseCompleted       = "completed"
	PhaseFailed          = "failed"
)

// Runner processes jobs asynchronously.
type Runner struct {
	store *store.Store
	cfg   config.Config
	queue chan *store.Job
}

func NewRunner(st *store.Store, cfg config.Config) *Runner {
	return &Runner{
		store: st,
		cfg:   cfg,
		queue: make(chan *store.Job, 16),
	}
}

// Start begins processing jobs until the context is done.
func (r *Runner) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case job := <-r.queue:
				if job == nil {
					continue
				}
				if err := r.handle(job); err != nil {
					log.Printf("job %s failed: %v", job.ID, err)
				}
			}
		}
	}()
}

func (r *Runner) Enqueue(job *store.Job) {
	r.queue <- job
}

func (r *Runner) handle(job *store.Job) error {
	if err := r.store.UpdateJobState(job.ID, StatusRunning, PhaseFetchingSource, "Fetching pixeldrain link via doubledouble.top (stubbed)", 0.05, false); err != nil {
		return err
	}
	_ = r.store.AddJobLog(job.ID, "Fetching pixeldrain link via doubledouble.top (stubbed)")
	time.Sleep(300 * time.Millisecond)

	if err := r.store.UpdateJobState(job.ID, StatusRunning, PhaseDownloading, "Downloading zip (stubbed)", 0.2, false); err != nil {
		return err
	}
	_ = r.store.AddJobLog(job.ID, "Downloading zip (stubbed)")
	time.Sleep(300 * time.Millisecond)

	if err := r.store.UpdateJobState(job.ID, StatusRunning, PhaseExtracting, "Extracting archive (stubbed)", 0.45, false); err != nil {
		return err
	}
	_ = r.store.AddJobLog(job.ID, "Extracting archive (stubbed)")
	time.Sleep(300 * time.Millisecond)

	if err := r.placeFiles(job); err != nil {
		_ = r.store.UpdateJobState(job.ID, StatusFailed, PhaseFailed, err.Error(), job.Progress, true)
		_ = r.store.AddJobLog(job.ID, fmt.Sprintf("Job failed: %v", err))
		return err
	}

	if err := r.store.UpdateJobState(job.ID, StatusRunning, PhaseCleanup, "Cleaning up temp files (stubbed)", 0.95, false); err != nil {
		return err
	}
	_ = r.store.AddJobLog(job.ID, "Cleaning up temp files (stubbed)")
	time.Sleep(150 * time.Millisecond)

	if err := r.store.UpdateJobState(job.ID, StatusCompleted, PhaseCompleted, "Completed", 1.0, true); err != nil {
		return err
	}
	_ = r.store.AddJobLog(job.ID, "Job completed")
	return nil
}

func (r *Runner) placeFiles(job *store.Job) error {
	artist := sanitizeName(job.Artist)
	album := sanitizeName(job.Album)
	if artist == "" {
		artist = "Unknown Artist"
	}
	if album == "" {
		album = "Unknown Album"
	}

	targetDir := filepath.Join(r.cfg.NavidromePath, artist, album)
	if _, err := os.Stat(targetDir); err == nil {
		msg := fmt.Sprintf("Album already exists at %s, skipping", targetDir)
		_ = r.store.AddJobLog(job.ID, msg)
		_ = r.store.UpdateJobState(job.ID, StatusCompleted, PhaseCompleted, msg, 1.0, true)
		return nil
	}
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("create target dir: %w", err)
	}

	if err := r.store.UpdateJobState(job.ID, StatusRunning, PhasePlacing, "Placing files into Navidrome path (stubbed)", 0.7, false); err != nil {
		return err
	}
	_ = r.store.AddJobLog(job.ID, fmt.Sprintf("Writing placeholder files to %s", targetDir))

	placeholder := filepath.Join(targetDir, "IMPORT_README.txt")
	content := fmt.Sprintf("Placeholder import for job %s\nArtist: %s\nAlbum: %s\nThis is a stub; wire actual download/extract logic.", job.ID, artist, album)
	if err := os.WriteFile(placeholder, []byte(content), 0644); err != nil {
		return fmt.Errorf("write placeholder: %w", err)
	}
	return nil
}

func sanitizeName(name string) string {
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.TrimSpace(name)
	return name
}
