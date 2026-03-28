'use client'; // Error components must be Client Components

import { useEffect } from 'react';
import Link from 'next/link';
import { logger } from '../lib/logger';

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    // Log the error — silenced in production via APP_ENV
    logger.error('[ErrorBoundary]', 'Global error caught', {
      message: error.message,
      digest: error.digest,
      stack: error.stack,
    });
  }, [error]);

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      minHeight: '60vh',
      textAlign: 'center',
      padding: '20px',
    }}>
      <div style={{
        backgroundColor: 'rgba(239, 68, 68, 0.1)',
        border: '1px solid rgba(239, 68, 68, 0.2)',
        borderRadius: '50%',
        width: '64px',
        height: '64px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        marginBottom: '24px',
      }}>
        <svg width="32" height="32" fill="none" stroke="#ef4444" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
        </svg>
      </div>
      <h2 style={{ fontSize: '2rem', marginBottom: '16px', color: '#fff' }}>Something went wrong!</h2>
      <p style={{ color: 'var(--text-muted)', marginBottom: '32px', maxWidth: '400px' }}>
        We encountered an unexpected error. Please try again or return to the dashboard.
      </p>
      <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap', justifyContent: 'center' }}>
        <button
          onClick={() => reset()}
          className="btn btn-primary"
        >
          Try again
        </button>
        <Link href="/dashboard" className="btn btn-ghost">
          Go to Dashboard
        </Link>
      </div>
    </div>
  );
}
