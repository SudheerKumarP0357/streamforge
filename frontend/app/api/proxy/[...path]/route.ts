import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'
import { logger } from '../../../../lib/logger'

const BACKEND = process.env.API_URL || 'http://localhost:8080'

export async function GET(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  const { path } = await params
  return proxy(req, path, 'GET')
}

export async function POST(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  const { path } = await params
  return proxy(req, path, 'POST')
}

export async function DELETE(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  const { path } = await params
  return proxy(req, path, 'DELETE')
}

export async function PUT(req: NextRequest, { params }: { params: Promise<{ path: string[] }> }) {
  const { path } = await params
  return proxy(req, path, 'PUT')
}

async function proxy(req: NextRequest, pathSegments: string[], method: string) {
  const path = pathSegments.join('/')
  const search = req.nextUrl.search
  const url = `${BACKEND}/api/v1/${path}${search}`
  const targetUrlBase = url.split('?')[0] // Scrub SAS tokens

  const token = (await cookies()).get('sf_access_token')?.value
  const requestId = crypto.randomUUID()

  logger.info('proxy request start', {
    component: 'proxy',
    method,
    path: `/${path}`,
    target_url: targetUrlBase,
    request_id: requestId
  })

  const headers: Record<string, string> = {
    'X-Request-ID': requestId
  }
  if (token) headers['Authorization'] = `Bearer ${token}`

  const startTime = Date.now()

  // In dev mode, expose the resolved backend URL to the client via a response header
  const devHeaders: Record<string, string> = {}
  if (logger.isEnabled) {
    devHeaders['X-SF-Backend-URL'] = targetUrlBase
  }

  // For multipart uploads: forward the body and content-type as-is
  const contentType = req.headers.get('content-type') || ''
  let body: BodyInit | undefined = undefined;

  if (contentType.includes('multipart/form-data')) {
    body = await req.arrayBuffer()
    headers['Content-Type'] = contentType   // preserve boundary parameter
  } else {
    headers['Content-Type'] = 'application/json'
    body = method !== 'GET' && method !== 'DELETE' ? await req.text() : undefined
  }

  try {
    const res = await fetch(url, {
      method,
      headers,
      body,
      cache: 'no-store',
    })

    const duration_ms = Date.now() - startTime
    const data = await res.text()
    const resContentType = res.headers.get('Content-Type') || 'application/json'

    const logCtx = {
      component: 'proxy',
      request_id: requestId,
      status_code: res.status,
      duration_ms,
      content_type: resContentType
    }

    if (res.status >= 500) {
      logger.error('proxy response error', undefined, logCtx)
    } else if (res.status >= 400) {
      logger.warn('proxy response warning', logCtx)
    } else {
      logger.info('proxy response success', logCtx)
    }

    return new NextResponse(data, {
      status: res.status,
      headers: {
        'Content-Type': resContentType,
        ...devHeaders,
      },
    })
  } catch (err) {
    const duration_ms = Date.now() - startTime
    const error_message = err instanceof Error ? err.message : String(err)
    
    logger.error('proxy request failed', err instanceof Error ? err : undefined, { 
      component: 'proxy',
      request_id: requestId,
      target_url: targetUrlBase,
      duration_ms,
      error_message
    })

    return NextResponse.json(
      {
        error: 'Failed to reach backend service',
        ...(logger.isEnabled ? { backend_url: targetUrlBase, backend: BACKEND } : {}),
      },
      {
        status: 503,
        headers: { ...devHeaders },
      }
    )
  }
}
