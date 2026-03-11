'use client';
import Link from 'next/link';
import UploadForm from '../../components/UploadForm';

export default function UploadPage() {
  return (
    <div style={{ paddingTop: 'var(--nav-h)' }}>
      <div className="container-app" style={{ maxWidth: 960, paddingTop: 36, paddingBottom: 60 }}>

        {/* Header */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 28 }} className="animate-fade-up">
          <Link href="/dashboard" className="btn btn-ghost btn-icon" title="Back to Dashboard">
            <svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0 }}>
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="1.5" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
            </svg>
          </Link>
          <div>
            <p style={{ color: 'var(--accent)', fontSize: '0.72rem', letterSpacing: '0.14em', textTransform: 'uppercase', fontWeight: 500, marginBottom: 4 }}>
              New video
            </p>
            <h1 className="serif" style={{ fontSize: '2rem', color: '#fff', lineHeight: 1 }}>Upload</h1>
            <p style={{ color: 'var(--text-muted)', fontSize: '0.875rem', marginTop: 5 }}>
              Select a video file to transcode and stream.
            </p>
          </div>
        </div>

        {/* Form card */}
        <div className="card animate-fade-up delay-100" style={{ padding: '28px 32px' }}>
          <UploadForm />
        </div>
      </div>
    </div>
  );
}