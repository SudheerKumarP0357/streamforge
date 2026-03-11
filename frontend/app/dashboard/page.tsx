import { cookies } from 'next/headers';
import Link from 'next/link';
import { Video } from '../../lib/types';
import VideoGrid from '../../components/VideoGrid';
import { getUserIdFromToken } from '../../lib/token';
import { serverApiUrl } from '../../lib/config';

const BASE_URL = serverApiUrl;

export default async function DashboardPage() {
  const cookieStore = cookies();
  const token = cookieStore.get('sf_access_token')?.value;

  let videos: Video[] = [];

  if (token) {
    const userId = getUserIdFromToken(token);

    try {
      const res = await fetch(`${BASE_URL}/api/v1/videos?limit=50`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store'
      });
      if (res.ok) {
        const data = await res.json();
        videos = data.videos || [];
      }
    } catch (e) {
      console.error('Failed to fetch videos:', e);
    }
  }

  return (
    <div style={{ paddingTop: 'var(--nav-h)' }}>
      <div className="container-app" style={{ paddingTop: 40, paddingBottom: 64 }}>

        {/* ── Page header ── */}
        <div className="animate-fade-up" style={{ marginBottom: 36 }}>
          <div style={{
            display: 'flex',
            flexDirection: 'column',
            gap: 20,
          }}>
            {/* Title row */}
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 16 }}>
              <h1 className="serif" style={{ fontSize: '2rem', color: '#fff', lineHeight: 1.1 }}>
                Dashboard
              </h1>
              <Link href="/upload" className="btn btn-primary">
                <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0 }}>
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                </svg>
                Upload video
              </Link>
            </div>


          </div>
        </div>

        {/* ── Divider ── */}
        <div className="divider-glow animate-fade-up delay-75" style={{ marginBottom: 28 }} />

        {/* ── Grid ── */}
        <div className="animate-fade-up delay-150">
          <VideoGrid initialVideos={videos} />
        </div>

      </div>
    </div>
  );
}