'use client'

import Link from 'next/link'
import { useRouter, usePathname } from 'next/navigation'
import { useState, useEffect } from 'react'
import { clearAuthCookies, getTokenCookie } from '../lib/api'
import { getUserRoleFromToken } from '../lib/token'

export default function NavBar() {
  const router = useRouter()
  const pathname = usePathname()
  const [scrolled, setScrolled] = useState(false)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [isAdmin, setIsAdmin] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8)
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    // Check admin role whenever pathname changes (e.g. after login/logout)
    const token = getTokenCookie('sf_access_token')
    if (token) {
      setIsAdmin(getUserRoleFromToken(token) === 'admin')
    } else {
      setIsAdmin(false)
    }
  }, [pathname])

  // Close mobile menu on route change
  useEffect(() => { setMobileOpen(false) }, [pathname])

  // Hide on auth pages
  if (pathname === '/login' || pathname === '/register') return null

  const handleLogout = () => {
    clearAuthCookies()
    router.push('/login')
  }

  return (
    <>
      <nav
        style={{
          position: 'fixed',
          top: 0, left: 0, right: 0,
          height: 'var(--nav-h)',
          zIndex: 50,
          transition: 'background 0.25s, border-color 0.25s, box-shadow 0.25s',
          background: scrolled ? 'rgba(8,8,8,0.88)' : 'rgba(8,8,8,0.0)',
          borderBottom: `1px solid ${scrolled ? 'rgba(255,255,255,0.07)' : 'transparent'}`,
          backdropFilter: scrolled ? 'blur(20px)' : 'none',
          WebkitBackdropFilter: scrolled ? 'blur(20px)' : 'none',
          boxShadow: scrolled ? '0 1px 24px rgba(0,0,0,0.45)' : 'none',
        }}
      >
        <div
          style={{
            maxWidth: 1280,
            margin: '0 auto',
            padding: '0 24px',
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: 16,
          }}
        >

          {/* ── Logo ── */}
          <button
            onClick={() => router.push('/dashboard')}
            style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0, display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}
          >
            <div style={{ position: 'relative', width: 28, height: 28, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
              <svg
                width="28" height="28"
                viewBox="0 0 32 32"
                style={{ position: 'absolute', inset: 0, animation: 'spin-ring 9s linear infinite' }}
              >
                <circle cx="16" cy="16" r="13" fill="none"
                  stroke="rgba(245,200,66,0.28)" strokeWidth="1.2" strokeDasharray="3.5 4" />
              </svg>
              <div className="accent-dot" />
            </div>
            <span style={{ fontWeight: 600, fontSize: '0.88rem', letterSpacing: '0.1em', color: 'rgba(255,255,255,0.88)' }}>
              STREAM<span style={{ color: 'var(--accent)' }}>FORGE</span>
            </span>
          </button>

          {/* ── Desktop nav ── */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 2, flex: 1, justifyContent: 'center' }} className="hidden-mobile">
            <NavItem href="/dashboard" label="Dashboard" pathname={pathname} />
            <NavItem href="/upload" label="Upload" pathname={pathname} />
            <NavItem href="/history" label="History" pathname={pathname} />
            {isAdmin && <NavItem href="/admin/stats" label="Admin" pathname={pathname} />}
          </div>

          {/* ── Right side ── */}
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }} className="hidden-mobile">
            {/* Sign out — subtle ghost style */}
            <button
              onClick={handleLogout}
              style={{
                display: 'inline-flex', alignItems: 'center', gap: 7,
                padding: '7px 14px',
                background: 'rgba(255,255,255,0.04)',
                border: '1px solid rgba(255,255,255,0.08)',
                borderRadius: 8,
                color: 'rgba(255,255,255,0.45)',
                fontSize: '0.82rem', fontWeight: 500,
                cursor: 'pointer',
                transition: 'background 0.15s, border-color 0.15s, color 0.15s',
              }}
              onMouseEnter={e => {
                const el = e.currentTarget;
                el.style.background = 'rgba(248,113,113,0.08)';
                el.style.borderColor = 'rgba(248,113,113,0.2)';
                el.style.color = '#f87171';
              }}
              onMouseLeave={e => {
                const el = e.currentTarget;
                el.style.background = 'rgba(255,255,255,0.04)';
                el.style.borderColor = 'rgba(255,255,255,0.08)';
                el.style.color = 'rgba(255,255,255,0.45)';
              }}
            >
              <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2" style={{ flexShrink: 0 }}>
                <path strokeLinecap="round" strokeLinejoin="round"
                  d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              Sign out
            </button>
          </div>

          {/* ── Mobile hamburger ── */}
          <button
            className="show-mobile"
            onClick={() => setMobileOpen(v => !v)}
            aria-label="Toggle menu"
            style={{
              background: 'none', border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: 8, padding: '7px 9px', cursor: 'pointer',
              color: 'rgba(255,255,255,0.6)',
              display: 'none',
            }}
          >
            {mobileOpen ? (
              <svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            ) : (
              <svg width="18" height="18" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round" d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            )}
          </button>

        </div>
      </nav>

      {/* ── Mobile dropdown ── */}
      {mobileOpen && (
        <div
          style={{
            position: 'fixed', top: 'var(--nav-h)', left: 0, right: 0,
            zIndex: 49,
            background: 'rgba(10,10,10,0.97)',
            backdropFilter: 'blur(20px)',
            borderBottom: '1px solid rgba(255,255,255,0.07)',
            padding: '12px 20px 20px',
            animation: 'slide-down 0.22s cubic-bezier(0.16,1,0.3,1) both',
          }}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
            <MobileNavItem href="/dashboard" label="Dashboard" pathname={pathname} />
            <MobileNavItem href="/upload" label="Upload" pathname={pathname} />
            <MobileNavItem href="/history" label="History" pathname={pathname} />
            {isAdmin && <MobileNavItem href="/admin/stats" label="Admin" pathname={pathname} />}
            <div style={{ height: 1, background: 'rgba(255,255,255,0.06)', margin: '8px 0' }} />
            <button
              onClick={handleLogout}
              style={{
                display: 'flex', alignItems: 'center', gap: 10,
                padding: '12px 14px', borderRadius: 10,
                background: 'rgba(248,113,113,0.06)',
                border: '1px solid rgba(248,113,113,0.12)',
                color: '#f87171', fontSize: '0.9rem', fontWeight: 500,
                cursor: 'pointer', width: '100%',
              }}
            >
              <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
                <path strokeLinecap="round" strokeLinejoin="round"
                  d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              Sign out
            </button>
          </div>
        </div>
      )}
    </>
  )
}

function NavItem({ href, label, pathname }: { href: string; label: string; pathname: string }) {
  const active = pathname === href || pathname.startsWith(href + '/')
  return (
    <Link
      href={href}
      style={{
        position: 'relative',
        padding: '6px 14px',
        borderRadius: 8,
        fontSize: '0.875rem',
        fontWeight: active ? 500 : 400,
        color: active ? '#fff' : 'rgba(255,255,255,0.48)',
        transition: 'color 0.15s, background 0.15s',
        textDecoration: 'none',
      }}
      onMouseEnter={e => {
        if (!active) {
          e.currentTarget.style.color = 'rgba(255,255,255,0.85)';
          e.currentTarget.style.background = 'rgba(255,255,255,0.05)';
        }
      }}
      onMouseLeave={e => {
        if (!active) {
          e.currentTarget.style.color = 'rgba(255,255,255,0.48)';
          e.currentTarget.style.background = 'transparent';
        }
      }}
    >
      {label}
      {active && (
        <span style={{
          position: 'absolute',
          bottom: -1, left: 14, right: 14,
          height: 2,
          background: 'var(--accent)',
          borderRadius: 99,
          boxShadow: '0 0 8px rgba(245,200,66,0.6)',
        }} />
      )}
    </Link>
  )
}

function MobileNavItem({ href, label, pathname }: { href: string; label: string; pathname: string }) {
  const active = pathname === href || pathname.startsWith(href + '/')
  return (
    <Link
      href={href}
      style={{
        display: 'flex', alignItems: 'center', gap: 12,
        padding: '12px 14px', borderRadius: 10,
        background: active ? 'rgba(245,200,66,0.06)' : 'transparent',
        border: `1px solid ${active ? 'rgba(245,200,66,0.14)' : 'transparent'}`,
        color: active ? 'var(--accent)' : 'rgba(255,255,255,0.6)',
        fontSize: '0.9rem', fontWeight: active ? 500 : 400,
        textDecoration: 'none',
        transition: 'background 0.15s',
      }}
    >
      {label}
    </Link>
  )
}