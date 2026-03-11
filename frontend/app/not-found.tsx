import Link from 'next/link';

export default function NotFound() {
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
        fontSize: '4rem',
        fontWeight: 'bold',
        color: 'var(--accent)',
        marginBottom: '16px',
        fontFamily: 'monospace',
      }}>
        404
      </div>
      <h2 style={{ fontSize: '2rem', marginBottom: '16px', color: '#fff' }}>Page Not Found</h2>
      <p style={{ color: 'var(--text-muted)', marginBottom: '32px', maxWidth: '400px' }}>
        The page you are looking for doesn't exist or has been moved.
      </p>
      <Link href="/dashboard" className="btn btn-primary">
        Go to Dashboard
      </Link>
    </div>
  );
}
