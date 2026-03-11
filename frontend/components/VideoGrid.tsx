'use client';
import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import VideoCard from './VideoCard';
import { Video } from '../lib/types';
import { videos as videosApi } from '../lib/api';

interface VideoGridProps {
  initialVideos: Video[];
}

export default function VideoGrid({ initialVideos }: VideoGridProps) {
  const [videos, setVideos] = useState<Video[]>(initialVideos);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    // If not searching, keep the local state in sync with the fresh server props
    if (!searchQuery.trim()) {
      setVideos(initialVideos);
    }
  }, [initialVideos]);

  useEffect(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }

    if (!searchQuery.trim()) {
      setVideos(initialVideos);
      setIsSearching(false);
      return;
    }

    setIsSearching(true);
    debounceTimerRef.current = setTimeout(async () => {
      try {
        const response = await videosApi.search(searchQuery.trim(), 1, 50);
        setVideos(response.videos);
      } catch (error) {
        console.error('Failed to search videos:', error);
      } finally {
        setIsSearching(false);
      }
    }, 400);

    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [searchQuery, initialVideos]);

  const handleDelete = (id: string) => setVideos(prev => prev.filter(v => v.id !== id));

  const handleRestore = (video: Video) => {
    setVideos(prev => [...prev, video].sort(
      (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    ));
  };

  if (!videos || videos.length === 0) {
    if (searchQuery.trim()) {
      return (
        <div className="video-grid-container" style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
          {/* Search Input */}
          <div style={{ position: 'relative', maxWidth: '400px', marginBottom: '8px' }}>
            <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-faint)' }} width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
            <input
              type="text"
              placeholder="Search videos..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="input-field"
              style={{
                width: '100%',
                paddingLeft: '38px',
                paddingRight: '12px',
                paddingTop: '10px',
                paddingBottom: '10px',
                background: 'rgba(255,255,255,0.03)',
                border: '1px solid rgba(255,255,255,0.1)',
                borderRadius: '8px',
                color: '#fff',
                fontSize: '0.9rem'
              }}
            />
          </div>

          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: 28,
            padding: '32px 36px',
            borderRadius: 16,
            border: '1px dashed rgba(255,255,255,0.08)',
            background: 'rgba(255,255,255,0.01)',
            maxWidth: 560,
          }}>
            <div style={{
              width: 64, height: 64,
              borderRadius: 16,
              flexShrink: 0,
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              background: 'rgba(255,255,255,0.03)',
              border: '1px solid rgba(255,255,255,0.08)',
            }}>
              <svg width="28" height="28" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                style={{ color: 'var(--text-faint)', flexShrink: 0 }}>
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5"
                  d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <div>
              <h3 className="serif" style={{ fontSize: '1.2rem', color: '#fff', marginBottom: 5, lineHeight: 1.2 }}>
                No results for "{searchQuery}"
              </h3>
              <p style={{ color: 'var(--text-faint)', fontSize: '0.835rem', lineHeight: 1.55, marginBottom: 16 }}>
                Try adjusting your search query or clear the search to see all videos.
              </p>
              <button 
                onClick={() => setSearchQuery('')} 
                className="btn btn-secondary btn-sm"
                style={{ background: 'rgba(255,255,255,0.05)', border: '1px solid rgba(255,255,255,0.1)', color: '#fff', padding: '6px 14px', borderRadius: '6px', fontSize: '0.85rem', cursor: 'pointer' }}
              >
                Clear search
              </button>
            </div>
          </div>
        </div>
      );
    }

    return (
      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 28,
        padding: '32px 36px',
        borderRadius: 16,
        border: '1px dashed rgba(255,255,255,0.08)',
        background: 'rgba(255,255,255,0.01)',
        maxWidth: 560,
      }}>
        {/* Icon box — fixed size, never stretches */}
        <div style={{
          width: 64, height: 64,
          borderRadius: 16,
          flexShrink: 0,
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: 'rgba(245,200,66,0.06)',
          border: '1px solid rgba(245,200,66,0.14)',
        }}>
          <svg width="28" height="28" fill="none" stroke="currentColor" viewBox="0 0 24 24"
            style={{ color: 'rgba(245,200,66,0.5)', flexShrink: 0 }}>
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5"
              d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
          </svg>
        </div>

        {/* Text + CTA */}
        <div>
          <h3 className="serif" style={{ fontSize: '1.2rem', color: '#fff', marginBottom: 5, lineHeight: 1.2 }}>
            No videos yet
          </h3>
          <p style={{ color: 'var(--text-faint)', fontSize: '0.835rem', lineHeight: 1.55, marginBottom: 16 }}>
            Your workspace is empty. Upload your first video to get started.
          </p>
          <Link href="/upload" className="btn btn-primary btn-sm">
            <svg width="13" height="13" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0 }}>
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5"
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            Upload your first video
          </Link>
        </div>
      </div>
    );
  }

  const handleUpdate = (updatedVideo: Video) => {
    setVideos(prev => 
      prev.map(v => (v.id === updatedVideo.id ? updatedVideo : v))
    );
  };

  return (
    <div className="video-grid-container" style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
      
      {/* Stats strip */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        <span style={{ fontSize: '0.78rem', color: 'var(--text-faint)', fontWeight: 500 }}>
          {videos.length === 0
            ? 'No videos yet'
            : `${videos.length} video${videos.length !== 1 ? 's' : ''}`}
        </span>
        {videos.length > 0 && (
          <>
            <span style={{ width: 3, height: 3, borderRadius: '50%', background: 'rgba(255,255,255,0.15)', display: 'inline-block' }} />
            <span style={{ fontSize: '0.78rem', color: 'var(--text-faint)' }}>
              {videos.filter(v => v.status === 'ready').length} ready
            </span>
            {videos.some(v => v.status === 'processing') && (
              <>
                <span style={{ width: 3, height: 3, borderRadius: '50%', background: 'rgba(255,255,255,0.15)', display: 'inline-block' }} />
                <span style={{ fontSize: '0.78rem', color: '#fbbf24' }}>
                  {videos.filter(v => v.status === 'processing').length} processing
                </span>
              </>
            )}
          </>
        )}
      </div>

      {/* Search Input */}
      {initialVideos.length > 0 && (
        <div style={{ position: 'relative', maxWidth: '400px', marginBottom: '8px' }}>
          <svg style={{ position: 'absolute', left: '12px', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-faint)' }} width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <input
            type="text"
            placeholder="Search videos..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="input-field"
            style={{
              width: '100%',
              paddingLeft: '38px',
              paddingRight: '12px',
              paddingTop: '10px',
              paddingBottom: '10px',
              background: 'rgba(255,255,255,0.03)',
              border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: '8px',
              color: '#fff',
              fontSize: '0.9rem',
              transition: 'all 0.2s ease',
              opacity: isSearching ? 0.7 : 1
            }}
          />
          {searchQuery && (
            <button 
              onClick={() => setSearchQuery('')}
              style={{
                position: 'absolute',
                right: '12px',
                top: '50%',
                transform: 'translateY(-50%)',
                background: 'none',
                border: 'none',
                color: 'var(--text-faint)',
                cursor: 'pointer',
                padding: '4px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center'
              }}
            >
              <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          )}
        </div>
      )}

      {/* Grid */}
      <div className="video-grid" style={{ opacity: isSearching ? 0.5 : 1, transition: 'opacity 0.2s ease' }}>
        {videos.map((video, i) => (
          <div
            key={video.id}
            className="animate-fade-up"
            style={{ animationDelay: `${Math.min(i * 0.05, 0.35)}s`, minWidth: 0 }}
          >
            <VideoCard video={video} onDelete={handleDelete} onRestore={handleRestore} onUpdate={handleUpdate} />
          </div>
        ))}
      </div>
    </div>
  );
}

/* ── Skeleton ── */
export function VideoGridSkeleton() {
  return (
    <div className="video-grid">
      {Array.from({ length: 8 }).map((_, i) => (
        <div key={i} style={{ minWidth: 0 }}>
          {/* Thumbnail */}
          <div className="skeleton" style={{ aspectRatio: '16/9', borderRadius: 12, width: '100%' }} />
          {/* Meta row */}
          <div style={{ display: 'flex', gap: 12, padding: '10px 2px 4px' }}>
            <div className="skeleton" style={{ width: 36, height: 36, borderRadius: '50%', flexShrink: 0 }} />
            <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 8, paddingTop: 2 }}>
              <div className="skeleton" style={{ height: 14, width: '90%', borderRadius: 4 }} />
              <div className="skeleton" style={{ height: 12, width: '55%', borderRadius: 4 }} />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}