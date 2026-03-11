'use client';
import { useState, useRef } from 'react';
import Link from 'next/link';
import { getTokenCookie } from '../lib/api';

const MAX_FILE_SIZE = 2 * 1024 * 1024 * 1024;

export default function UploadForm() {
  const [file, setFile] = useState<File | null>(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [uploadSpeed, setUploadSpeed] = useState(0);
  const [warning, setWarning] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);
  const xhrRef = useRef<XMLHttpRequest | null>(null);

  const fmtBytes = (b: number) => {
    if (!b) return '0 B';
    const k = 1024, sizes = ['B', 'KB', 'MB', 'GB'], i = Math.floor(Math.log(b) / Math.log(k));
    return parseFloat((b / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };
  const fmtSpeed = (bps: number) => bps ? (bps / (1024 * 1024)).toFixed(2) + ' MB/s' : '—';

  const handleFileSelection = (f: File) => {
    setError(null);
    setWarning(null);
    if (!f.type.startsWith('video/')) { setError('Please select a valid video file.'); return; }
    if (f.size > MAX_FILE_SIZE) { setError('File size exceeds the 2 GB limit.'); return; }
    if (f.size > 500 * 1024 * 1024) { setWarning('Large file (>500MB) selected. Upload may take a while.'); }
    setFile(f);
    if (!title) setTitle(f.name.split('.').slice(0, -1).join('.').substring(0, 100));
  };

  const onDrop = (e: React.DragEvent) => {
    e.preventDefault(); setIsDragging(false);
    if (e.dataTransfer.files?.[0]) handleFileSelection(e.dataTransfer.files[0]);
  };

  const handleUpload = () => {
    if (!file) { setError('Please select a file.'); return; }
    if (!title.trim()) { setError('Please enter a title.'); return; }
    setIsUploading(true); setError(null); setProgress(0); setUploadSpeed(0);

    const formData = new FormData();
    formData.append('title', title);
    if (description) formData.append('description', description);
    formData.append('file', file);

    const xhr = new XMLHttpRequest();
    xhrRef.current = xhr;
    const token = getTokenCookie('sf_access_token');

    xhr.open('POST', `/api/proxy/videos`, true);
    if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);

    let startTime = Date.now(), loadedStart = 0;
    xhr.upload.onprogress = (ev) => {
      if (!ev.lengthComputable) return;
      setProgress((ev.loaded / ev.total) * 100);
      const now = Date.now(), elapsed = (now - startTime) / 1000;
      if (elapsed >= 1) {
        setUploadSpeed((ev.loaded - loadedStart) / elapsed);
        startTime = now; loadedStart = ev.loaded;
      }
    };
    xhr.onload = () => {
      setIsUploading(false);
      if (xhr.status === 201) { setSuccess(true); return; }
      try { setError(JSON.parse(xhr.responseText).error || 'Upload failed.'); }
      catch { setError('Upload failed with status: ' + xhr.status); }
    };
    xhr.onerror = () => { setIsUploading(false); setError('Network error. Check your connection.'); };
    xhr.onabort = () => { setIsUploading(false); setError('Upload was cancelled.'); };
    xhr.send(formData);
  };

  const handleCancel = () => xhrRef.current?.abort();
  const resetForm = () => {
    setFile(null); setTitle(''); setDescription('');
    setProgress(0); setUploadSpeed(0); setError(null); setWarning(null); setSuccess(false);
  };

  /* ── Success ── */
  if (success) {
    return (
      <div className="animate-fade-up" style={{
        padding: '48px 40px', textAlign: 'center',
        border: '1px solid rgba(74,222,128,0.18)',
        borderRadius: 16,
        background: 'rgba(74,222,128,0.03)',
      }}>
        <div style={{
          width: 64, height: 64, borderRadius: '50%', margin: '0 auto 20px',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          background: 'rgba(74,222,128,0.1)', border: '1px solid rgba(74,222,128,0.25)',
          boxShadow: '0 0 28px rgba(74,222,128,0.12)',
        }}>
          <svg width="28" height="28" fill="none" stroke="#4ade80" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <h2 className="serif" style={{ fontSize: '1.6rem', color: '#fff', marginBottom: 8 }}>Uploaded successfully</h2>
        <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem', lineHeight: 1.65, marginBottom: 28, maxWidth: 340, margin: '0 auto 28px' }}>
          Your video is received. Adaptive bitrate transcoding is initializing — it'll be ready to stream shortly.
        </p>
        <div style={{ display: 'flex', gap: 10, justifyContent: 'center', flexWrap: 'wrap' }}>
          <Link href="/dashboard" className="btn btn-primary">Go to Dashboard</Link>
          <button onClick={resetForm} className="btn btn-ghost">Upload another</button>
        </div>
      </div>
    );
  }

  const overLimit = title.length > 100 || description.length > 500;
  const canSubmit = !!file && !!title.trim() && !overLimit && !isUploading;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>

      {/* ── Error banner ── */}
      {error && (
        <div className="animate-slide-down" style={{
          display: 'flex', alignItems: 'flex-start', gap: 10,
          padding: '11px 16px', borderRadius: 10, marginBottom: 20,
          background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.18)',
          color: '#f87171', fontSize: '0.85rem',
        }}>
          <svg width="15" height="15" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0, marginTop: 1 }}>
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span style={{ flex: 1 }}>{error}</span>
          {!isUploading && file && (
            <button onClick={handleUpload} style={{
              fontSize: '0.75rem', fontWeight: 600, color: '#f87171',
              background: 'rgba(239,68,68,0.12)', border: '1px solid rgba(239,68,68,0.22)',
              borderRadius: 6, padding: '3px 10px', cursor: 'pointer', flexShrink: 0,
            }}>Retry</button>
          )}
        </div>
      )}

      {/* ── Warning banner ── */}
      {warning && !error && (
        <div className="animate-slide-down" style={{
          display: 'flex', alignItems: 'flex-start', gap: 10,
          padding: '11px 16px', borderRadius: 10, marginBottom: 20,
          background: 'rgba(245,200,66,0.08)', border: '1px solid rgba(245,200,66,0.18)',
          color: '#fcd34d', fontSize: '0.85rem',
        }}>
          <svg width="15" height="15" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0, marginTop: 1 }}>
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
              d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span style={{ flex: 1 }}>{warning}</span>
        </div>
      )}

      {/* ── Two-column grid ── */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'minmax(0,1fr) minmax(0,1fr)',
        gap: 24,
      }}
        className="upload-grid"
      >

        {/* LEFT — drop zone */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 0 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
            <label style={{ fontSize: '0.68rem', fontWeight: 600, letterSpacing: '0.09em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.4)' }}>
              Video file
            </label>
            <span style={{ fontSize: '0.68rem', color: 'var(--text-faint)' }}>Supported formats: MP4, MOV, AVI, MKV</span>
          </div>

          <div
            onClick={() => !isUploading && fileInputRef.current?.click()}
            onDragOver={e => { e.preventDefault(); setIsDragging(true); }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={onDrop}
            role="button" tabIndex={0}
            onKeyDown={e => e.key === 'Enter' && fileInputRef.current?.click()}
            style={{
              flex: 1,
              minHeight: 220,
              borderRadius: 14,
              border: `2px dashed ${isDragging ? 'rgba(245,200,66,0.55)' : file ? 'rgba(74,222,128,0.35)' : 'rgba(255,255,255,0.1)'}`,
              background: isDragging
                ? 'rgba(245,200,66,0.03)'
                : file
                  ? 'rgba(74,222,128,0.02)'
                  : 'rgba(255,255,255,0.015)',
              cursor: isUploading ? 'not-allowed' : 'pointer',
              opacity: isUploading ? 0.55 : 1,
              transition: 'border-color 0.18s, background 0.18s',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              gap: 12,
              padding: 24,
              textAlign: 'center',
              outline: 'none',
            }}
          >
            <input
              ref={fileInputRef} type="file" accept="video/*"
              style={{ display: 'none' }} disabled={isUploading}
              onChange={e => e.target.files?.[0] && handleFileSelection(e.target.files[0])}
            />

            {file ? (
              /* ── File selected state ── */
              <>
                {/* File type badge */}
                <div style={{
                  width: 52, height: 52, borderRadius: 12, flexShrink: 0,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  background: 'rgba(74,222,128,0.08)', border: '1px solid rgba(74,222,128,0.2)',
                }}>
                  <svg width="22" height="22" fill="none" stroke="#4ade80" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5"
                      d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
                  </svg>
                </div>

                {/* Filename + size */}
                <div style={{ maxWidth: '100%' }}>
                  <p style={{
                    color: 'rgba(255,255,255,0.88)', fontSize: '0.875rem', fontWeight: 500,
                    overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                    maxWidth: 240,
                  }}>{file.name}</p>
                  <p style={{ color: 'var(--text-faint)', fontSize: '0.78rem', marginTop: 3 }}>
                    {fmtBytes(file.size)}
                  </p>
                </div>

                {/* Replace link */}
                {!isUploading && (
                  <button
                    type="button"
                    onClick={e => { e.stopPropagation(); setFile(null); setTitle(''); }}
                    style={{
                      fontSize: '0.75rem', color: 'var(--text-faint)',
                      background: 'none', border: '1px solid rgba(255,255,255,0.1)',
                      borderRadius: 6, padding: '4px 12px', cursor: 'pointer',
                      transition: 'color 0.15s, border-color 0.15s',
                    }}
                    onMouseEnter={e => { e.currentTarget.style.color = '#fff'; e.currentTarget.style.borderColor = 'rgba(255,255,255,0.22)'; }}
                    onMouseLeave={e => { e.currentTarget.style.color = 'var(--text-faint)'; e.currentTarget.style.borderColor = 'rgba(255,255,255,0.1)'; }}
                  >
                    Replace file
                  </button>
                )}
              </>
            ) : (
              /* ── Empty drop state ── */
              <>
                {/* Upload icon — FIXED size, no SVG overflow */}
                <div style={{
                  width: 52, height: 52, borderRadius: 14, flexShrink: 0,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  background: isDragging ? 'rgba(245,200,66,0.1)' : 'rgba(255,255,255,0.04)',
                  border: `1px solid ${isDragging ? 'rgba(245,200,66,0.3)' : 'rgba(255,255,255,0.08)'}`,
                  transition: 'background 0.18s, border-color 0.18s',
                }}>
                  <svg width="22" height="22" fill="none" stroke="currentColor" viewBox="0 0 24 24"
                    style={{ color: isDragging ? 'var(--accent)' : 'rgba(255,255,255,0.28)', flexShrink: 0 }}>
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5"
                      d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                  </svg>
                </div>

                <div>
                  <p style={{ color: isDragging ? 'var(--accent)' : 'rgba(255,255,255,0.6)', fontSize: '0.9rem', fontWeight: 500, transition: 'color 0.18s' }}>
                    {isDragging ? 'Drop to add file' : 'Drop video here'}
                  </p>
                  <p style={{ color: 'var(--text-faint)', fontSize: '0.78rem', marginTop: 4 }}>
                    or <span style={{ color: 'var(--accent)', fontWeight: 500 }}>browse to upload</span>
                  </p>
                </div>
              </>
            )}
          </div>
        </div>

        {/* RIGHT — metadata */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

          {/* Title */}
          <div>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 7 }}>
              <label className="field-label" style={{ marginBottom: 0 }}>
                Title <span style={{ color: 'var(--accent)' }}>*</span>
              </label>
              <span style={{ fontSize: '0.65rem', color: title.length > 90 ? (title.length > 100 ? 'var(--error)' : 'var(--warning)') : 'var(--text-faint)' }}>
                {title.length}/100
              </span>
            </div>
            <input
              type="text" value={title}
              onChange={e => setTitle(e.target.value)}
              disabled={isUploading} maxLength={100}
              className="field-input"
              placeholder="Enter a descriptive title…"
            />
          </div>

          {/* Description */}
          <div style={{ display: 'flex', flexDirection: 'column', flex: 1 }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 7 }}>
              <label className="field-label" style={{ marginBottom: 0 }}>Description</label>
              <span style={{ fontSize: '0.65rem', color: description.length > 450 ? (description.length > 500 ? 'var(--error)' : 'var(--warning)') : 'var(--text-faint)' }}>
                {description.length}/500
              </span>
            </div>
            <textarea
              value={description}
              onChange={e => setDescription(e.target.value)}
              disabled={isUploading} maxLength={500}
              className="field-input"
              style={{ flex: 1, minHeight: 120, resize: 'none' }}
              placeholder="Tell viewers what this video is about…"
            />
          </div>

          {/* Upload hint row */}
          <div style={{
            display: 'flex', alignItems: 'center', gap: 8,
            padding: '10px 12px', borderRadius: 9,
            background: 'rgba(255,255,255,0.025)',
            border: '1px solid rgba(255,255,255,0.06)',
          }}>
            <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24"
              style={{ color: 'var(--accent)', flexShrink: 0 }}>
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-faint)', lineHeight: 1.4 }}>
              Videos are transcoded to HLS for adaptive streaming after upload.
            </span>
          </div>
        </div>
      </div>

      {/* ── Progress bar (shown while uploading) ── */}
      {isUploading && (
        <div className="animate-fade-in" style={{ marginTop: 20 }}>
          {/* Stats row */}
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 8 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <svg className="animate-spin" width="13" height="13" fill="none" viewBox="0 0 24 24"
                style={{ color: 'var(--accent)', flexShrink: 0 }}>
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
              <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)', fontWeight: 500 }}>
                Uploading… {Math.round(progress)}%
              </span>
            </div>
            <span style={{ fontSize: '0.75rem', color: 'var(--accent)', fontFamily: 'monospace', fontWeight: 600 }}>
              {fmtSpeed(uploadSpeed)}
            </span>
          </div>

          {/* Track */}
          <div style={{
            height: 5, borderRadius: 99,
            background: 'rgba(255,255,255,0.06)',
            overflow: 'hidden',
          }}>
            <div style={{
              height: '100%', borderRadius: 99,
              width: `${progress}%`,
              background: 'linear-gradient(90deg, var(--accent), var(--accent-hover))',
              boxShadow: '0 0 10px rgba(245,200,66,0.4)',
              transition: 'width 0.3s ease',
            }} />
          </div>
        </div>
      )}

      {/* ── Actions ── */}
      <div style={{
        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
        marginTop: 20, paddingTop: 20,
        borderTop: '1px solid rgba(255,255,255,0.06)',
      }}>
        {/* Left: file status chip */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          {file && !isUploading ? (
            <span style={{
              display: 'inline-flex', alignItems: 'center', gap: 5,
              fontSize: '0.75rem', color: 'rgba(255,255,255,0.45)',
              background: 'rgba(255,255,255,0.04)',
              border: '1px solid rgba(255,255,255,0.08)',
              borderRadius: 6, padding: '3px 10px',
            }}>
              <svg width="11" height="11" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ color: '#4ade80' }}>
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M5 13l4 4L19 7" />
              </svg>
              {fmtBytes(file.size)}
            </span>
          ) : !file ? (
            <span style={{ fontSize: '0.75rem', color: 'var(--text-faint)' }}>No file selected</span>
          ) : null}
        </div>

        {/* Right: action button */}
        {isUploading ? (
          <button
            onClick={handleCancel}
            style={{
              display: 'inline-flex', alignItems: 'center', gap: 7,
              padding: '10px 18px', borderRadius: 10,
              background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.2)',
              color: '#f87171', fontSize: '0.875rem', fontWeight: 600, cursor: 'pointer',
              transition: 'background 0.15s',
            }}
            onMouseEnter={e => (e.currentTarget.style.background = 'rgba(239,68,68,0.18)')}
            onMouseLeave={e => (e.currentTarget.style.background = 'rgba(239,68,68,0.1)')}
          >
            <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
            Cancel
          </button>
        ) : (
          <button
            onClick={handleUpload}
            disabled={!canSubmit}
            style={{
              display: 'inline-flex', alignItems: 'center', gap: 8,
              padding: '11px 24px', borderRadius: 10,
              background: canSubmit ? 'var(--accent)' : 'rgba(255,255,255,0.07)',
              border: 'none',
              color: canSubmit ? '#0b0800' : 'rgba(255,255,255,0.25)',
              fontSize: '0.9rem', fontWeight: 700, cursor: canSubmit ? 'pointer' : 'not-allowed',
              boxShadow: canSubmit ? '0 4px 20px rgba(245,200,66,0.25)' : 'none',
              transition: 'background 0.18s, box-shadow 0.18s, transform 0.12s',
            }}
            onMouseEnter={e => { if (canSubmit) { e.currentTarget.style.background = 'var(--accent-hover)'; e.currentTarget.style.transform = 'translateY(-1px)'; } }}
            onMouseLeave={e => { if (canSubmit) { e.currentTarget.style.background = 'var(--accent)'; e.currentTarget.style.transform = 'translateY(0)'; } }}
          >
            <svg width="15" height="15" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0 }}>
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5"
                d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
            </svg>
            Upload video
          </button>
        )}
      </div>
    </div>
  );
}