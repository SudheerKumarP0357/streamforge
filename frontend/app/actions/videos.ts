'use server'

import { revalidateTag } from 'next/cache'
import { cookies } from 'next/headers'
import { getUserIdFromToken } from '../../lib/token'
import { serverApiUrl } from '../../lib/config'

export async function deleteVideoAction(videoId: string): Promise<{ success: boolean; error?: string }> {
  const token = cookies().get('sf_access_token')?.value
  if (!token) return { success: false, error: 'Not authenticated' }

  const userId = getUserIdFromToken(token)

  try {
    const res = await fetch(
      `${serverApiUrl}/api/v1/videos/${videoId}`,
      {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store'
      }
    )

    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      return { success: false, error: body.error ?? `Delete failed: ${res.status}` }
    }

    // Surgically invalidate ONLY this user's video list cache
    if (userId) {
      revalidateTag(`user-videos-${userId}`)     // ← busts the dashboard fetch
    }
    revalidateTag(`video-${videoId}`)            // ← busts the single video fetch

    return { success: true }

  } catch (err) {
    return { success: false, error: 'Network error. Please try again.' }
  }
}
