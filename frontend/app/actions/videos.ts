'use server'

import { revalidateTag } from 'next/cache'
import { cookies } from 'next/headers'
import { getUserIdFromToken } from '../../lib/token'
import { serverApiUrl } from '../../lib/config'
import { logger } from '../../lib/logger'

export async function deleteVideoAction(videoId: string): Promise<{ success: boolean; error?: string }> {
  const token = (await cookies()).get('sf_access_token')?.value
  if (!token) {
    logger.warn('Not authenticated — no access token', { component: 'deleteVideoAction' })
    return { success: false, error: 'Not authenticated' }
  }

  const userId = getUserIdFromToken(token)

  const deleteUrl = `${serverApiUrl}/api/v1/videos/${videoId}`
  logger.debug(`DELETE ${deleteUrl}`, { component: 'deleteVideoAction', action: 'server-fetch', video_id: videoId, user_id: userId, backend: serverApiUrl })

  try {
    const res = await fetch(
      deleteUrl,
      {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store'
      }
    )

    logger.info('Delete response', { component: 'deleteVideoAction', status: res.status, video_id: videoId })

    if (!res.ok) {
      const body = await res.json().catch(() => ({}))
      logger.error('Delete failed', { component: 'deleteVideoAction', status: res.status, video_id: videoId, body })
      return { success: false, error: body.error ?? `Delete failed: ${res.status}` }
    }

    // Surgically invalidate ONLY this user's video list cache
    if (userId) {
      logger.info('Revalidating cache', { component: 'deleteVideoAction', tags: [`user-videos-${userId}`, `video-${videoId}`] })
      revalidateTag(`user-videos-${userId}`, 'max')     // ← busts the dashboard fetch
    }
    revalidateTag(`video-${videoId}`, 'max')            // ← busts the single video fetch

    logger.info('Video deleted successfully', { component: 'deleteVideoAction', video_id: videoId })
    return { success: true }

  } catch (err) {
    logger.error('Network error during delete', { component: 'deleteVideoAction', video_id: videoId, err })
    return { success: false, error: 'Network error. Please try again.' }
  }
}

