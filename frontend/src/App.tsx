import { useEffect, useMemo, useState } from 'react'
import './App.css'
import { createImport, getJob, listJobs, search } from './api'
import type { ImportRequestItem, Job, SearchItem } from './types'

const MIN_QUERY = 2

function App() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<SearchItem[]>([])
  const [loadingSearch, setLoadingSearch] = useState(false)
  const [selected, setSelected] = useState<Record<string, SearchItem>>({})
  const [showConfirm, setShowConfirm] = useState(false)
  const [error, setError] = useState('')

  const [jobId, setJobId] = useState('')
  const [activeJob, setActiveJob] = useState<Job | null>(null)
  const [recentJobs, setRecentJobs] = useState<Job[]>([])

  useEffect(() => {
    const handler = setTimeout(() => {
      const trimmed = query.trim()
      if (trimmed.length < MIN_QUERY) {
        setResults([])
        return
      }
      setLoadingSearch(true)
      setError('')
      search(trimmed)
        .then(setResults)
        .catch((err) => setError(err.message || 'Search failed'))
        .finally(() => setLoadingSearch(false))
    }, 250)
    return () => clearTimeout(handler)
  }, [query])

  useEffect(() => {
    listJobs()
      .then((res) => setRecentJobs(res.jobs ?? []))
      .catch(() => {
        // ignore for now; likely backend not running yet
      })
  }, [])

  useEffect(() => {
    if (!jobId) return
    let stop = false
    const poll = async () => {
      try {
        const job = await getJob(jobId)
        setActiveJob(job)
        if (job.status === 'completed' || job.status === 'failed') {
          stop = true
          listJobs().then((res) => setRecentJobs(res.jobs ?? []))
          return
        }
      } catch (err) {
        console.error(err)
      }
    }
    poll()
    const interval = setInterval(() => {
      if (stop) return
      poll()
    }, 1500)
    return () => clearInterval(interval)
  }, [jobId])

  const normalizedSelection = useMemo(() => {
    const next: Record<string, SearchItem> = {}
    Object.values(selected).forEach((item) => {
      const normalized = normalizeSelection(item)
      next[normalized.id] = normalized
    })
    return next
  }, [selected])

  const toggleSelection = (item: SearchItem) => {
    const key = item.type === 'song' && item.albumId ? item.albumId : item.id
    setSelected((prev) => {
      const next = { ...prev }
      if (next[key]) {
        delete next[key]
      } else {
        next[key] = item
      }
      return next
    })
  }

  const startImport = async () => {
    const items: ImportRequestItem[] = Object.values(normalizedSelection).map((i) => ({
      id: i.id,
      type: i.type,
      title: i.title,
      artist: i.artist,
      albumId: i.albumId,
      albumTitle: i.albumTitle,
      coverUrl: i.coverUrl,
    }))
    setError('')
    try {
      const res = await createImport(items)
      setJobId(res.jobId)
      setShowConfirm(false)
      setSelected({})
    } catch (err: any) {
      setError(err?.message ?? 'Failed to start import')
    }
  }

  return (
    <div className="page">
      <header className="hero">
        <div>
          <p className="eyebrow">Navidrome Import Helper</p>
          <h1>Search Amazon Music, import albums, and track progress.</h1>
          <p className="lede">
            Pick albums or singles (songs resolve to their parent albums). Confirm to enqueue an import;
            the backend will download, extract, and place files into your Navidrome path.
          </p>
        </div>
        <div className="summary-badge">
          <div className="pill">SQLite jobs</div>
          <div className="pill alt">Album-first imports</div>
        </div>
      </header>

      <section className="panel">
        <div className="panel-header">
          <div>
            <h2>Search Amazon Music</h2>
            <p>Albums and singles are shown; selecting a song will import its album.</p>
          </div>
          <div className="input-wrap">
            <input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search albums or songs..."
            />
            <span className="hint">{loadingSearch ? 'Searching...' : 'Debounced search'}</span>
          </div>
        </div>
        {error && <div className="error">{error}</div>}

        <div className="results-grid">
          {results.map((item) => {
            const normalized = normalizeSelection(item)
            const selectedState = normalizedSelection[normalized.id]
            return (
              <article
                key={item.id}
                className={`card ${selectedState ? 'selected' : ''}`}
                onClick={() => toggleSelection(item)}
              >
                <img className="cover" src={item.coverUrl} alt={`${item.title} cover`} />
                <div className="card-body">
                  <div className="type-pill">{item.type === 'song' ? 'Song → Album' : 'Album'}</div>
                  <h3>{item.title}</h3>
                  <p className="muted">{item.artist}</p>
                  {item.albumTitle && item.type === 'song' && (
                    <p className="muted small">Album: {item.albumTitle}</p>
                  )}
                  <p className="muted small">
                    {item.tracks ? `${item.tracks} tracks` : 'single'} ·{' '}
                    {item.duration ? formatDuration(item.duration) : 'duration n/a'}
                  </p>
                </div>
              </article>
            )
          })}
          {query.trim().length >= MIN_QUERY && !loadingSearch && results.length === 0 && (
            <div className="empty">No results from backend yet.</div>
          )}
          {query.trim().length < MIN_QUERY && (
            <div className="empty">Type at least {MIN_QUERY} characters to search.</div>
          )}
        </div>
      </section>

      <section className="panel selection">
        <div className="panel-header">
          <div>
            <h2>Selection</h2>
            <p>{Object.keys(normalizedSelection).length} album(s) will be imported.</p>
          </div>
          <div className="actions">
            <button
              className="ghost"
              onClick={() => setSelected({})}
              disabled={Object.keys(normalizedSelection).length === 0}
            >
              Clear
            </button>
            <button
              className="primary"
              disabled={Object.keys(normalizedSelection).length === 0}
              onClick={() => setShowConfirm(true)}
            >
              Confirm import
            </button>
          </div>
        </div>
        <div className="selection-list">
          {Object.values(normalizedSelection).map((item) => (
            <div key={item.id} className="selection-row">
              <div>
                <div className="label">{item.title}</div>
                <div className="muted small">{item.artist}</div>
                {item.type === 'song' && (
                  <div className="muted tiny">Song selected → album will be imported</div>
                )}
              </div>
              <button className="link" onClick={() => toggleSelection(item)}>
                Remove
              </button>
            </div>
          ))}
          {Object.keys(normalizedSelection).length === 0 && (
            <div className="empty">Select albums or songs above to import.</div>
          )}
        </div>
      </section>

      <section className="panel jobs">
        <div className="panel-header">
          <div>
            <h2>Active job</h2>
            <p>Progress updates from the backend worker.</p>
          </div>
          <button className="ghost" onClick={() => (jobId ? setJobId(jobId) : null)}>
            Refresh
          </button>
        </div>
        {activeJob ? (
          <div className="job-card">
            <div className="job-meta">
              <div className="badge">{activeJob.status}</div>
              <div>
                <div className="label">{activeJob.album || 'Unknown album'}</div>
                <div className="muted small">{activeJob.artist}</div>
              </div>
            </div>
            <div className="progress">
              <div className="progress-bar" style={{ width: `${(activeJob.progress || 0) * 100}%` }} />
            </div>
            <div className="muted small">
              Phase: {activeJob.phase} · {activeJob.message}
            </div>
            {activeJob.logs && activeJob.logs.length > 0 && (
              <div className="logs">
                {activeJob.logs.slice(-4).map((log) => (
                  <div key={log.createdAt} className="tiny muted">
                    {new Date(log.createdAt).toLocaleTimeString()} — {log.message}
                  </div>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div className="empty">No job selected yet.</div>
        )}
      </section>

      <section className="panel">
        <div className="panel-header">
          <div>
            <h2>Recent jobs</h2>
            <p>Latest 50 jobs from SQLite.</p>
          </div>
          <button className="ghost" onClick={() => listJobs().then((res) => setRecentJobs(res.jobs ?? []))}>
            Refresh
          </button>
        </div>
        {recentJobs.length === 0 && <div className="empty">No jobs yet.</div>}
        <div className="recent-grid">
          {recentJobs.map((job) => (
            <div key={job.id} className="recent-card">
              <div className="badge">{job.status}</div>
              <div className="label">{job.album || 'Unknown album'}</div>
              <div className="muted small">{job.artist}</div>
              <div className="muted tiny">{new Date(job.createdAt).toLocaleString()}</div>
            </div>
          ))}
        </div>
      </section>

      {showConfirm && (
        <div className="dialog-backdrop">
          <div className="dialog">
            <h3>Import {Object.keys(normalizedSelection).length} album(s)?</h3>
            <p>
              Selected items will be imported at album level. Backend will download via doubledouble.top → pixeldrain,
              extract, and place under NAVIDROME_MUSIC_PATH.
            </p>
            <div className="dialog-actions">
              <button className="ghost" onClick={() => setShowConfirm(false)}>
                Cancel
              </button>
              <button className="primary" onClick={startImport}>
                Yes, start import
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function normalizeSelection(item: SearchItem): SearchItem {
  if (item.type === 'song' && item.albumId) {
    return {
      ...item,
      id: item.albumId,
      type: 'album',
      title: item.albumTitle || item.title,
    }
  }
  return item
}

function formatDuration(seconds?: number) {
  if (!seconds) return '–'
  const mins = Math.round(seconds / 60)
  return `${mins} min`
}

export default App
