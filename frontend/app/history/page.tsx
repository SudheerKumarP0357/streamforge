import { cookies } from 'next/headers';
import Link from 'next/link';
import { WatchHistory } from '../../lib/types';
import { serverApiUrl } from '../../lib/config';

const BASE_URL = serverApiUrl

export default async function HistoryPage() {
  const cookieStore = cookies();
  const token = cookieStore.get('sf_access_token')?.value;

  let history: WatchHistory[] = [];

  if (token) {
    try {
      const res = await fetch(`${BASE_URL}/api/v1/watch/history`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      });
      if (res.ok) {
        const data = await res.json();
        history = data.history || [];
      }
    } catch (e) {
      console.error('Failed to fetch history:', e);
    }
  }

  return (
    <div style={{ paddingTop: 'var(--nav-h)' }}>
      <div className="container-app" style={{ maxWidth: 860, paddingTop: 40, paddingBottom: 64 }}>

        {/* ── Header ── */}
        <div className="animate-fade-up" style={{ marginBottom: 32 }}>
          <h1 className="serif" style={{ fontSize: '2rem', color: '#fff', lineHeight: 1.1 }}>
            Watch History
          </h1>
          <p style={{ color: 'var(--text-faint)', fontSize: '0.85rem', marginTop: 6 }}>
            {history.length > 0 ? `${history.length} item${history.length !== 1 ? 's' : ''}` : 'No activity yet'}
          </p>
        </div>

        <div className="divider-glow animate-fade-up delay-75" style={{ marginBottom: 28 }} />

        {history.length === 0 ? (

          /* ── Empty state — horizontal layout, fixed icon size ── */
          <div className="animate-fade-up delay-100" style={{
            display: 'flex', alignItems: 'center', gap: 24,
            padding: '28px 32px', borderRadius: 14,
            border: '1px dashed rgba(255,255,255,0.08)',
            background: 'rgba(255,255,255,0.01)',
            maxWidth: 480,
          }}>
            <div style={{
              width: 56, height: 56, borderRadius: 14, flexShrink: 0,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              background: 'rgba(245,200,66,0.06)', border: '1px solid rgba(245,200,66,0.13)',
            }}>
              <svg width="24" height="24" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                style={{ color: 'rgba(245,200,66,0.5)' }}>
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5"
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h2 className="serif" style={{ fontSize: '1.15rem', color: '#fff', marginBottom: 5 }}>No history yet</h2>
              <p style={{ color: 'var(--text-faint)', fontSize: '0.835rem', lineHeight: 1.55, marginBottom: 14 }}>
                Start watching videos to build your personal history.
              </p>
              <Link href="/dashboard" className="btn btn-primary btn-sm">Explore videos</Link>
            </div>
          </div>

        ) : (

          /* ── History list ── */
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
            {history.map((item, idx) => {
              const progressPercent = item.duration_seconds && item.duration_seconds > 0
                ? Math.min(100, Math.round((item.progress_seconds / item.duration_seconds) * 100))
                : 0;

              return (
                <Link
                  key={`${item.video_id}-${idx}`}
                  href={`/watch/${item.video_id}`}
                  className="animate-fade-up"
                  style={{ animationDelay: `${Math.min(idx * 0.05, 0.4)}s`, color: 'inherit', display: 'block' }}
                >
                  <div className="card card-hover" style={{
                    display: 'flex', flexDirection: 'row',
                    alignItems: 'center', gap: 16,
                    padding: '12px 16px', cursor: 'pointer',
                  }}>

                    {/* Thumbnail */}
                    <div style={{
                      width: 176, flexShrink: 0,
                      aspectRatio: '16/9',
                      borderRadius: 10,
                      overflow: 'hidden',
                      position: 'relative',
                      background: 'rgba(22,22,22,1)',
                      border: '1px solid rgba(255,255,255,0.06)',
                    }}>
                      {/* Gradient */}
                      <div style={{
                        position: 'absolute', inset: 0, zIndex: 1,
                        background: 'linear-gradient(to top, rgba(10,10,10,0.65) 0%, transparent 55%)',
                      }} />

                      {/* Play button */}
                      <div style={{
                        position: 'absolute', inset: 0, zIndex: 2,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                      }}>
                        <div style={{
                          width: 34, height: 34, borderRadius: '50%',
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                          background: 'rgba(0,0,0,0.55)',
                          border: '1.5px solid rgba(255,255,255,0.14)',
                          backdropFilter: 'blur(6px)',
                        }}>
                          <svg width="12" height="12" fill="white" viewBox="0 0 24 24"
                            style={{ marginLeft: 2, flexShrink: 0 }}>
                            <path d="M8 5v14l11-7z" />
                          </svg>
                        </div>
                      </div>

                      {/* Completed badge */}
                      {item.completed && (
                        <div style={{ position: 'absolute', top: 6, right: 6, zIndex: 3 }}>
                          <span className="badge badge-success" style={{ fontSize: '0.58rem', padding: '2px 6px' }}>
                            Watched
                          </span>
                        </div>
                      )}

                      {/* Progress bar */}
                      {!item.completed && progressPercent > 0 && (
                        <div style={{
                          position: 'absolute', bottom: 0, left: 0, right: 0, zIndex: 3,
                          height: 3, background: 'rgba(255,255,255,0.08)',
                        }}>
                          <div style={{
                            height: '100%', width: `${progressPercent}%`,
                            background: 'linear-gradient(90deg, var(--accent), var(--accent-hover))',
                            boxShadow: '0 0 6px rgba(245,200,66,0.5)',
                          }} />
                        </div>
                      )}
                    </div>

                    {/* Text info */}
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <h3 style={{
                        fontSize: '0.95rem', fontWeight: 500,
                        color: 'rgba(255,255,255,0.88)',
                        marginBottom: 6, lineHeight: 1.35,
                        overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                      }}>
                        {item.title}
                      </h3>

                      <div style={{ display: 'flex', flexWrap: 'wrap', alignItems: 'center', gap: 8, fontSize: '0.775rem', color: 'var(--text-faint)' }}>
                        {/* Date */}
                        <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}>
                          <svg width="13" height="13" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                            style={{ flexShrink: 0 }}>
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                              d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                          </svg>
                          {new Date(item.watched_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                        </span>

                        {progressPercent > 0 && !item.completed && (
                          <>
                            <span style={{ width: 3, height: 3, borderRadius: '50%', background: 'rgba(255,255,255,0.15)', flexShrink: 0, display: 'inline-block' }} />
                            <span style={{ color: 'var(--accent)', fontWeight: 500 }}>{progressPercent}% watched</span>
                          </>
                        )}

                        {item.completed && (
                          <>
                            <span style={{ width: 3, height: 3, borderRadius: '50%', background: 'rgba(255,255,255,0.15)', flexShrink: 0, display: 'inline-block' }} />
                            <span className="badge badge-success" style={{ fontSize: '0.6rem', padding: '2px 7px' }}>Completed</span>
                          </>
                        )}
                      </div>
                    </div>

                    {/* Chevron */}
                    <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                      style={{ color: 'rgba(255,255,255,0.18)', flexShrink: 0 }}>
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 5l7 7-7 7" />
                    </svg>

                  </div>
                </Link>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
}