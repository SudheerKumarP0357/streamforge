import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { getUserRoleFromToken } from '../../../lib/token';
import { AdminStats } from '../../../lib/types';
import { serverApiUrl } from '../../../lib/config';

export default async function AdminStatsPage() {
  const cookieStore = await cookies();
  const token = cookieStore.get('sf_access_token')?.value;

  if (!token) {
    redirect('/login');
  }

  const role = getUserRoleFromToken(token);
  if (role !== 'admin') {
    return (
      <div style={{ padding: 40 }}>
        <h2>403 Forbidden</h2>
        <p>You do not have permission to view the admin panel.</p>
      </div>
    );
  }

  let stats: AdminStats | null = null;
  
  try {
    const res = await fetch(`${serverApiUrl}/api/v1/admin/stats`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store'
    });
    if (res.ok) {
      stats = await res.json();
    }
  } catch (e) {
    console.error('Failed to fetch admin stats:', e);
  }

  if (!stats) {
    return <div>Loading stats...</div>
  }

  const StatCard = ({ label, value, color = '#fff' }: { label: string, value: string | number, color?: string }) => (
    <div style={{
      background: 'rgba(255,255,255,0.03)',
      border: '1px solid rgba(255,255,255,0.08)',
      borderRadius: 16,
      padding: '24px 32px',
      display: 'flex',
      flexDirection: 'column',
      gap: 8
    }}>
      <span style={{ fontSize: '0.9rem', color: 'var(--text-faint)', fontWeight: 500 }}>{label}</span>
      <span className="serif" style={{ fontSize: '2.5rem', color, lineHeight: 1 }}>{value}</span>
    </div>
  );

  return (
    <div className="animate-fade-up">
      <h1 className="serif" style={{ fontSize: '2rem', color: '#fff', marginBottom: 8 }}>System Statistics</h1>
      <p style={{ color: 'var(--text-faint)', marginBottom: 32 }}>Overview of StreamForge growth and transcode queue health.</p>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(240px, 1fr))', gap: 24, marginBottom: 40 }}>
        <StatCard label="Total Users" value={stats.total_users} color="#60a5fa" />
        <StatCard label="Total Videos" value={stats.total_videos} />
      </div>

      <h2 style={{ fontSize: '1.2rem', color: '#fff', marginBottom: 20 }}>Video Status Breakdown</h2>
      
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 16 }}>
        <StatCard label="Ready" value={stats.videos_by_status.ready} color="#4ade80" />
        <StatCard label="Processing" value={stats.videos_by_status.processing} color="#fbbf24" />
        <StatCard label="Pending" value={stats.videos_by_status.pending} color="#a1a1aa" />
        <StatCard label="Failed" value={stats.videos_by_status.failed} color="#f87171" />
      </div>
    </div>
  );
}
