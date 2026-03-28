import { cookies } from 'next/headers';
import { redirect } from 'next/navigation';
import { getUserRoleFromToken } from '../../../lib/token';
import { AdminVideo } from '../../../lib/types';
import { serverApiUrl } from '../../../lib/config';
import { logger } from '../../../lib/logger';

export default async function AdminVideosPage({ searchParams }: { searchParams: Promise<{ page?: string }> }) {
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

  const resolvedParams = await searchParams;
  const page = parseInt(resolvedParams.page || '1', 10);
  let videos: AdminVideo[] = [];
  let total = 0;

  const fetchUrl = `${serverApiUrl}/api/v1/admin/videos?page=${page}&limit=50`;
  logger.serverFetch('GET', fetchUrl, { page, backend: serverApiUrl });

  try {
    const res = await fetch(fetchUrl, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store'
    });
    logger.info('[AdminVideosPage]', 'Admin videos fetch response', { status: res.status, page });
    if (res.ok) {
      const data = await res.json();
      videos = data.videos || [];
      total = data.total || 0;
      logger.info('[AdminVideosPage]', `Loaded ${videos.length} videos (total: ${total})`);
    }
  } catch (e) {
    logger.error('[AdminVideosPage]', 'Failed to fetch admin videos', e);
  }

  return (
    <div className="animate-fade-up">
      <h1 className="serif" style={{ fontSize: '2rem', color: '#fff', marginBottom: 8 }}>System Videos</h1>
      <p style={{ color: 'var(--text-faint)', marginBottom: 32 }}>Managing all {total} uploaded videos across the platform.</p>

      <div style={{ overflowX: 'auto', background: 'rgba(255,255,255,0.02)', borderRadius: 12, border: '1px solid rgba(255,255,255,0.08)' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid rgba(255,255,255,0.08)' }}>
              <th style={{ padding: '16px 20px', color: 'var(--text-faint)', fontWeight: 500, fontSize: '0.85rem' }}>ID</th>
              <th style={{ padding: '16px 20px', color: 'var(--text-faint)', fontWeight: 500, fontSize: '0.85rem' }}>Title</th>
              <th style={{ padding: '16px 20px', color: 'var(--text-faint)', fontWeight: 500, fontSize: '0.85rem' }}>Owner Email</th>
              <th style={{ padding: '16px 20px', color: 'var(--text-faint)', fontWeight: 500, fontSize: '0.85rem' }}>Status</th>
              <th style={{ padding: '16px 20px', color: 'var(--text-faint)', fontWeight: 500, fontSize: '0.85rem' }}>Created</th>
              <th style={{ padding: '16px 20px', color: 'var(--text-faint)', fontWeight: 500, fontSize: '0.85rem' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {videos.map(video => (
              <tr key={video.id} style={{ borderBottom: '1px solid rgba(255,255,255,0.04)' }}>
                <td style={{ padding: '16px 20px', fontSize: '0.85rem', color: 'var(--text-faint)', fontFamily: 'monospace' }}>
                  {video.id.split('-')[0]}...
                </td>
                <td style={{ padding: '16px 20px', fontWeight: 500 }}>
                  {video.title} {video.duration_seconds && <span style={{ color: 'var(--text-faint)', fontWeight: 400, marginLeft: 6 }}>({Math.floor(video.duration_seconds / 60)}:{String(video.duration_seconds % 60).padStart(2, '0')})</span>}
                </td>
                <td style={{ padding: '16px 20px', fontSize: '0.9rem', color: '#a0aec0' }}>
                  {video.user_email}
                </td>
                <td style={{ padding: '16px 20px' }}>
                  <span style={{
                    padding: '4px 8px', borderRadius: 4, fontSize: '0.75rem', fontWeight: 600,
                    background: video.status === 'ready' ? 'rgba(34,197,94,0.1)' :
                      video.status === 'failed' ? 'rgba(239,68,68,0.1)' : 'rgba(251,191,36,0.1)',
                    color: video.status === 'ready' ? '#4ade80' :
                      video.status === 'failed' ? '#f87171' : '#fbbf24'
                  }}>
                    {video.status.toUpperCase()}
                  </span>
                </td>
                <td style={{ padding: '16px 20px', fontSize: '0.85rem', color: 'var(--text-faint)' }}>
                  {new Date(video.created_at).toLocaleDateString()}
                </td>
                <td style={{ padding: '16px 20px' }}>
                  {/* Delete button handled by a client component in real production, but here we can just render the UI */}
                  <button className="btn btn-secondary btn-sm" style={{ padding: '6px 12px', fontSize: '0.75rem' }}>
                    Delete
                  </button>
                </td>
              </tr>
            ))}
            {videos.length === 0 && (
              <tr>
                <td colSpan={6} style={{ padding: 40, textAlign: 'center', color: 'var(--text-faint)' }}>
                  No videos found in the system.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
