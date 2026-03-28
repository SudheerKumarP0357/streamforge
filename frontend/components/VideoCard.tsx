'use client';
import { useState, useEffect } from 'react';
import Link from 'next/link';
import { Video, VideoStatus } from '../lib/types';
import { videos as apiVideos } from '../lib/api';
import { deleteVideoAction } from '../app/actions/videos';
import { logger } from '../lib/logger';

interface VideoCardProps {
  video: Video;
  onDelete: (id: string) => void;
  onRestore: (video: Video) => void;
  onUpdate?: (video: Video) => void;
}

function getRelativeTime(dateString: string) {
  const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' });
  const diff = Math.floor((new Date(dateString).getTime() - Date.now()) / 1000);
  const units: { unit: Intl.RelativeTimeFormatUnit; seconds: number }[] = [
    { unit: 'year', seconds: 31536000 },
    { unit: 'month', seconds: 2592000 },
    { unit: 'week', seconds: 604800 },
    { unit: 'day', seconds: 86400 },
    { unit: 'hour', seconds: 3600 },
    { unit: 'minute', seconds: 60 },
    { unit: 'second', seconds: 1 },
  ];
  for (const { unit, seconds } of units) {
    if (Math.abs(diff) >= seconds || unit === 'second')
      return rtf.format(Math.round(diff / seconds), unit);
  }
  return 'just now';
}

function fmtDuration(s?: number | null) {
  if (!s) return null;
  const h = Math.floor(s / 3600);
  const m = Math.floor((s % 3600) / 60);
  const sec = s % 60;
  if (h > 0) return `${h}:${m < 10 ? '0' : ''}${m}:${sec < 10 ? '0' : ''}${sec}`;
  return `${m}:${sec < 10 ? '0' : ''}${sec}`;
}

const STATUS_CFG: Record<VideoStatus, { label: string; color: string; bg: string; border: string }> = {
  ready: { label: 'Ready', color: '#4ade80', bg: 'rgba(74,222,128,0.1)', border: 'rgba(74,222,128,0.25)' },
  processing: { label: 'Processing', color: '#fbbf24', bg: 'rgba(251,191,36,0.1)', border: 'rgba(251,191,36,0.25)' },
  pending: { label: 'Pending', color: '#94a3b8', bg: 'rgba(148,163,184,0.08)', border: 'rgba(148,163,184,0.18)' },
  failed: { label: 'Failed', color: '#f87171', bg: 'rgba(248,113,113,0.1)', border: 'rgba(248,113,113,0.25)' },
};

function initials(title: string) {
  return title.split(/\s+/).slice(0, 2).map(w => w[0] ?? '').join('').toUpperCase() || '??';
}

export default function VideoCard({ video: initialVideo, onDelete, onRestore, onUpdate }: VideoCardProps) {
  const [video, setVideo] = useState<Video>(initialVideo);
  const [isDeleting, setIsDeleting] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  // Sync state if props change from parent
  useEffect(() => {
    setVideo(initialVideo);
  }, [initialVideo]);

  useEffect(() => {
    if (video.status !== 'processing' && video.status !== 'pending') return;
    logger.info('[VideoCard]', 'Starting status polling', { videoId: video.id, currentStatus: video.status });
    const iv = setInterval(async () => {
      try {
        const updated = await apiVideos.getById(video.id);
        setVideo(updated);
        
        // Let parent know status changed so stats strip updates
        if (updated.status !== video.status) {
          logger.info('[VideoCard]', 'Video status changed', { videoId: video.id, from: video.status, to: updated.status });
          onUpdate?.(updated);
        }

        if (updated.status === 'ready' || updated.status === 'failed') {
          logger.info('[VideoCard]', 'Polling complete — terminal status', { videoId: video.id, status: updated.status });
          clearInterval(iv);
        }
      } catch (e) {
        logger.error('[VideoCard]', 'Failed to poll video status', { videoId: video.id, error: e });
      }
    }, 5000);
    return () => clearInterval(iv);
  }, [video.id, video.status]);

  const handleDeleteClick = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    if (!window.confirm('Delete this video? This cannot be undone.')) {
      setMenuOpen(false);
      return;
    }
    
    setIsDeleting(true);
    setDeleteError(null);
    logger.info('[VideoCard]', 'Deleting video (optimistic)', { videoId: video.id, title: video.title });
    
    // Optimistic removal - hide the card immediately
    onDelete(video.id);
    
    // Call Server Action
    const result = await deleteVideoAction(video.id);
    
    if (!result.success) {
      logger.error('[VideoCard]', 'Delete failed, rolling back', { videoId: video.id, error: result.error });
      // Rollback: restore the card if deletion failed
      onRestore(video);
      setDeleteError(result.error ?? 'Delete failed. Please try again.');
      setIsDeleting(false);
    } else {
      logger.info('[VideoCard]', 'Delete confirmed by server', { videoId: video.id });
    }
  };

  const sc = STATUS_CFG[video.status] ?? STATUS_CFG.pending;
  const dur = fmtDuration(video.duration_seconds);

  const inner = (
    <div
      className="video-card"
      style={{ opacity: isDeleting ? 0.5 : 1, pointerEvents: isDeleting ? 'none' : undefined }}
    >
      {/* ── Thumbnail ── */}
      <div className="video-thumb">
        <div
          className="absolute inset-0"
          style={{ background: 'linear-gradient(135deg,rgba(245,200,66,0.04) 0%,rgba(20,20,20,0.92) 100%)' }}
        />

        {/* Centre play icon */}
        <div
          className="absolute inset-0 flex items-center justify-center"
          style={{ zIndex: 1, opacity: video.status === 'ready' ? 1 : 0.3 }}
        >
          <div style={{
            width: 44, height: 44, borderRadius: '50%',
            background: 'rgba(0,0,0,0.6)',
            border: '1.5px solid rgba(255,255,255,0.15)',
            backdropFilter: 'blur(8px)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}>
            <svg width="16" height="16" fill="white" viewBox="0 0 24 24" style={{ marginLeft: 2 }}>
              <path d="M8 5v14l11-7z" />
            </svg>
          </div>
        </div>

        {dur && <span className="video-duration">{dur}</span>}
        <span className="video-status-dot" style={{ background: sc.color, boxShadow: `0 0 5px ${sc.color}` }} />

        {video.status === 'processing' && (
          <div className="absolute inset-0 flex items-center justify-center" style={{ zIndex: 2, background: 'rgba(0,0,0,0.5)' }}>
            <svg className="animate-spin" width="28" height="28" fill="none" viewBox="0 0 24 24" style={{ color: '#fbbf24' }}>
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="3" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
            </svg>
          </div>
        )}
      </div>

      {/* ── Meta row ── */}
      <div className="video-meta">
        <div className="video-avatar">{initials(video.title)}</div>

        <div className="video-info">
          <p className="video-title">{video.title}</p>
          <div className="video-sub">
            <span>{getRelativeTime(video.created_at)}</span>
            <span className="video-sub-dot" />
            <span style={{
              fontSize: '0.68rem', fontWeight: 600, letterSpacing: '0.05em', textTransform: 'uppercase',
              color: sc.color, background: sc.bg, border: `1px solid ${sc.border}`,
              borderRadius: 4, padding: '1px 5px', lineHeight: 1.5,
              display: 'inline-flex', alignItems: 'center', gap: 3,
            }}>
              {video.status === 'processing' && (
                <svg className="animate-spin" width="7" height="7" fill="none" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                </svg>
              )}
              {sc.label}
            </span>
          </div>
        </div>

        {/* Three-dot menu */}
        <div style={{ position: 'relative', flexShrink: 0 }}>
          <button
            className="video-action"
            onClick={e => { e.preventDefault(); e.stopPropagation(); setMenuOpen(v => !v); }}
            aria-label="Video options"
          >
            <svg width="18" height="18" fill="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="5" r="1.5" /><circle cx="12" cy="12" r="1.5" /><circle cx="12" cy="19" r="1.5" />
            </svg>
          </button>

          {menuOpen && (
            <>
              <div className="fixed inset-0" style={{ zIndex: 30 }}
                onClick={e => { e.preventDefault(); e.stopPropagation(); setMenuOpen(false); }} />
              <div style={{
                position: 'absolute', right: 0, top: 28, zIndex: 40,
                minWidth: 164, borderRadius: 12, overflow: 'hidden',
                background: 'rgba(18,18,18,0.97)',
                border: '1px solid rgba(255,255,255,0.1)',
                backdropFilter: 'blur(16px)',
                boxShadow: '0 12px 32px rgba(0,0,0,0.65)',
                padding: '4px 0',
              }}>
                <button
                  onClick={handleDeleteClick}
                  disabled={isDeleting}
                  style={{
                    width: '100%', display: 'flex', alignItems: 'center', gap: 10,
                    padding: '10px 16px', fontSize: '0.82rem', fontWeight: 500,
                    color: '#f87171', background: 'none', border: 'none', cursor: isDeleting ? 'not-allowed' : 'pointer',
                    textAlign: 'left', transition: 'background 0.12s',
                    opacity: isDeleting ? 0.7 : 1,
                  }}
                  onMouseEnter={e => { if (!isDeleting) e.currentTarget.style.background = 'rgba(248,113,113,0.08)' }}
                  onMouseLeave={e => (e.currentTarget.style.background = 'none')}
                >
                  {isDeleting ? (
                    <svg className="animate-spin" width="14" height="14" fill="none" viewBox="0 0 24 24">
                      <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                      <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                    </svg>
                  ) : (
                    <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
                      <path strokeLinecap="round" strokeLinejoin="round"
                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                    </svg>
                  )}
                  {isDeleting ? 'Deleting...' : 'Delete video'}
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );

  return (
    <div style={{ minWidth: 0 }}>
      {video.status === 'ready' ? (
        <Link href={`/watch/${video.id}`} style={{ display: 'block', color: 'inherit', minWidth: 0 }}>
          {inner}
        </Link>
      ) : (
        inner
      )}
      
      {deleteError && (
        <div style={{ 
          marginTop: 12, 
          fontSize: '0.85rem', 
          color: '#f87171', 
          padding: '8px 12px', 
          background: 'rgba(248,113,113,0.1)', 
          border: '1px solid rgba(248,113,113,0.2)',
          borderRadius: 8,
          display: 'flex',
          alignItems: 'center',
          gap: 6
        }}>
          <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {deleteError}
        </div>
      )}
    </div>
  );
}