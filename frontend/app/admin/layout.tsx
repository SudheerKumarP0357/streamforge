import Link from 'next/link';

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <div style={{ paddingTop: 'var(--nav-h)', display: 'flex' }}>
      {/* Admin Sidebar */}
      <div style={{
        width: 250,
        height: 'calc(100vh - var(--nav-h))',
        borderRight: '1px solid rgba(255,255,255,0.08)',
        background: 'rgba(0,0,0,0.2)',
        padding: '32px 24px',
        display: 'flex',
        flexDirection: 'column',
        gap: 12
      }}>
        <h2 className="serif" style={{ fontSize: '1.2rem', color: '#fff', marginBottom: 16 }}>Admin Panel</h2>

        <Link href="/admin/stats" style={{
          padding: '10px 14px',
          borderRadius: 8,
          background: 'rgba(255,255,255,0.05)',
          color: '#fff',
          fontSize: '0.9rem',
          textDecoration: 'none'
        }}>
          Statistics
        </Link>

        <Link href="/admin/videos" style={{
          padding: '10px 14px',
          borderRadius: 8,
          background: 'rgba(255,255,255,0.05)',
          color: '#fff',
          fontSize: '0.9rem',
          textDecoration: 'none'
        }}>
          All Videos
        </Link>
      </div>

      {/* Admin Content */}
      <div style={{ flex: 1, padding: '40px 64px', overflowY: 'auto', height: 'calc(100vh - var(--nav-h))' }}>
        {children}
      </div>
    </div>
  );
}
