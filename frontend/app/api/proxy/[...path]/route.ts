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

  const token = (await cookies()).get('sf_access_token')?.value

  logger.serverFetch(method, url, { hasToken: !!token, backend: BACKEND })

  const headers: Record<string, string> = {}
  if (token) headers['Authorization'] = `Bearer ${token}`

  const startTime = Date.now()

  // In dev mode, expose the resolved backend URL to the client via a response header
  const devHeaders: Record<string, string> = {}
  if (logger.isEnabled) {
    devHeaders['X-SF-Backend-URL'] = url
  }

  // For multipart uploads: forward the body and content-type as-is
  // Do NOT set Content-Type manually — let the browser set it with the boundary
  const contentType = req.headers.get('content-type') || ''
  if (contentType.includes('multipart/form-data')) {
    logger.info('[proxy]', 'Multipart upload detected', { path, contentType: contentType.substring(0, 60) })
    const body = await req.arrayBuffer()
    headers['Content-Type'] = contentType   // preserve boundary parameter
    const res = await fetch(url, { method, headers, body, cache: 'no-store' })
    const elapsed = Date.now() - startTime
    const data = await res.text()
    logger.serverFetch(method, url, { status: res.status, elapsed, multipart: true })
    return new NextResponse(data, {
      status: res.status,
      headers: { ...devHeaders },
    })
  }

  // Not multipart
  headers['Content-Type'] = 'application/json'

  const body = method !== 'GET' && method !== 'DELETE'
    ? await req.text()
    : undefined

  try {
    const res = await fetch(url, {
      method,
      headers,
      body,
      cache: 'no-store',
    })

    const elapsed = Date.now() - startTime
    const data = await res.text()

    logger.serverFetch(method, url, { status: res.status, elapsed, responseLength: data.length })

    return new NextResponse(data, {
      status: res.status,
      headers: {
        'Content-Type': res.headers.get('Content-Type') || 'application/json',
        ...devHeaders,
      },
    })
  } catch (err) {
    const elapsed = Date.now() - startTime
    logger.error('[proxy]', `${method} ${url} failed after ${elapsed}ms`, err)
    return NextResponse.json(
      {
        error: 'Failed to reach backend service',
        ...(logger.isEnabled ? { backend_url: url, backend: BACKEND } : {}),
      },
      {
        status: 503,
        headers: { ...devHeaders },
      }
    )
  }
}
