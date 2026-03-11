import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  // Define public paths that don't require authentication
  const publicPaths = ['/login', '/register']
  
  const isPublicPath = publicPaths.some(path => request.nextUrl.pathname.startsWith(path))
  
  // If the path is public, let the request continue
  if (isPublicPath) {
    return NextResponse.next()
  }

  // Get the token from cookies
  const cookieName = process.env.NEXT_PUBLIC_JWT_COOKIE_NAME || 'sf_access_token'
  const token = request.cookies.get(cookieName)

  // If no token exists, redirect to login page
  if (!token) {
    const loginUrl = new URL('/login', request.url)
    return NextResponse.redirect(loginUrl)
  }

  // Allow the request to continue. Token validation happens backend side.
  return NextResponse.next()
}

// Configure the paths that middleware should run on
export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes)
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}
