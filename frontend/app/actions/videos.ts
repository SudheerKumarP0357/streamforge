'use server'

import { revalidateTag } from 'next/cache'
import { cookies } from 'next/headers'
import { getUserIdFromToken } from '../../lib/token'
import { serverApiUrl } from '../../lib/config'
import { logger } from '../../lib/logger'

export async function deleteVideoAction(videoId: string): Promise<{ success: boolean; error?: string }> {
  const token = (await cookies()).get('sf_access_token')?.value
  if (!token) {
    logger.warn('[deleteVideoAction]', 'Not authenticated — no access token')
    return { success: false, error: 'Not authenticated' }
  }

  const userId = getUserIdFromToken(token)

  const deleteUrl = `${serverApiUrl}/api/v1/videos/${videoId}`
  logger.serverFetch('DELETE', deleteUrl, { videoId, userId, backend: serverApiUrl })

  try {
    const res = await fetch(
      deleteUrl,
      {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store'
      }
    )

    logger.info('[deleteVideoAction]', 'Delete response', { status: res.status, videoId })

    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      logger.error('[deleteVideoAction]', 'Delete failed', { status: res.status, videoId, body })
      return { success: false, error: body.error ?? `Delete failed: ${res.status}` }
    }

    // Surgically invalidate ONLY this user's video list cache
    if (userId) {
      logger.info('[deleteVideoAction]', 'Revalidating cache', { tags: [`user-videos-${userId}`, `video-${videoId}`] })
      revalidateTag(`user-videos-${userId}`, 'max')     // ← busts the dashboard fetch
    }
    revalidateTag(`video-${videoId}`, 'max')            // ← busts the single video fetch

    logger.info('[deleteVideoAction]', 'Video deleted successfully', { videoId })
    return { success: true }

  } catch (err) {
    logger.error('[deleteVideoAction]', 'Network error during delete', { videoId, err })
    return { success: false, error: 'Network error. Please try again.' }
  }
}

