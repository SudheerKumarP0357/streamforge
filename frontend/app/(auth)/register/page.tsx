'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { auth } from '../../../lib/api';
import { logger } from '../../../lib/logger';

export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPass, setShowPass] = useState(false);
  const [showConf, setShowConf] = useState(false);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const validateEmail = (e: string) => /\S+@\S+\.\S+/.test(e);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    if (!email || !validateEmail(email)) { setError('Please enter a valid email address.'); return; }
    if (password.length < 8) { setError('Password must be at least 8 characters long.'); return; }
    if (password !== confirmPassword) { setError('Passwords do not match.'); return; }

    setLoading(true);
    logger.info('[RegisterPage]', 'Submitting registration form', { email });
    try {
      logger.info('[RegisterPage]', 'Calling auth.register', { email });
      await auth.register(email, password);
      logger.info('[RegisterPage]', 'Registration successful, auto-logging in');
      await auth.login(email, password);
      logger.info('[RegisterPage]', 'Auto-login successful, redirecting to dashboard');
      router.push('/dashboard');
      router.refresh();
    } catch (err: any) {
      logger.error('[RegisterPage]', 'Registration failed', { email, error: err?.error });
      const msg = err?.error?.toLowerCase() || '';
      if (msg.includes('duplicate') || msg.includes('exists') || msg.includes('taken') || msg.includes('already')) {
        setError('Email address is already in use.');
      } else {
        setError(err?.error || 'Failed to create account. Please try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-page">

      {/* ── Background layers ── */}
      <div style={{ position: 'absolute', inset: 0, background: '#080808' }} />
      <div className="bg-grid" style={{ position: 'absolute', inset: 0 }} />
      <div style={{
        position: 'absolute', inset: 0,
        background: 'linear-gradient(135deg, #0d0a00 0%, #080808 55%, #00080d 100%)',
        opacity: 0.8,
      }} />

      {/* Ambient orbs */}
      <div className="animate-orb-a" style={{
        position: 'absolute', top: '15%', left: '-15%',
        width: 500, height: 500, borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(245,200,66,0.07) 0%, transparent 70%)',
        pointerEvents: 'none',
      }} />
      <div className="animate-orb-b" style={{
        position: 'absolute', bottom: '15%', right: '-15%',
        width: 500, height: 500, borderRadius: '50%',
        background: 'radial-gradient(circle, rgba(66,180,245,0.05) 0%, transparent 70%)',
        pointerEvents: 'none',
      }} />

      {/* ── Card ── */}
      <div className="auth-card">
        <div style={{
          borderRadius: 20,
          overflow: 'hidden',
          background: 'rgba(11,11,11,0.93)',
          backdropFilter: 'blur(28px)',
          WebkitBackdropFilter: 'blur(28px)',
          border: '1px solid rgba(255,255,255,0.07)',
          boxShadow: '0 0 0 1px rgba(255,255,255,0.03) inset, 0 48px 96px rgba(0,0,0,0.7)',
        }}>

          {/* Amber top accent */}
          <div style={{ height: 1, background: 'linear-gradient(90deg, transparent, rgba(245,200,66,0.55) 50%, transparent)' }} />

          <div style={{ padding: '36px 40px 32px' }}>

            {/* Logo */}
            <div className="auth-f1" style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 32 }}>
              <div style={{ position: 'relative', width: 28, height: 28, display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
                <svg width="28" height="28" viewBox="0 0 32 32"
                  style={{ position: 'absolute', inset: 0, animation: 'spin-ring 9s linear infinite' }}>
                  <circle cx="16" cy="16" r="13" fill="none"
                    stroke="rgba(245,200,66,0.25)" strokeWidth="1.2" strokeDasharray="3.5 4" />
                </svg>
                <div className="accent-dot" />
              </div>
              <span style={{ color: 'rgba(255,255,255,0.65)', fontSize: '0.75rem', letterSpacing: '0.18em', textTransform: 'uppercase', fontWeight: 500 }}>
                StreamForge
              </span>
            </div>

            {/* Headline */}
            <div className="auth-f1" style={{ marginBottom: 28 }}>
              <h1 className="serif" style={{ fontSize: '2.4rem', lineHeight: 1.1, color: '#fff', marginBottom: 6 }}>
                Create your<br />
                <span className="serif italic" style={{ color: 'var(--accent)' }}>account.</span>
              </h1>
              <p style={{ color: 'rgba(255,255,255,0.32)', fontSize: '0.875rem' }}>
                Join StreamForge to start uploading videos
              </p>
            </div>

            {/* Error */}
            {error && (
              <div className="auth-err" style={{
                marginBottom: 20, padding: '10px 14px', borderRadius: 10,
                background: 'rgba(239,68,68,0.08)', border: '1px solid rgba(239,68,68,0.18)',
                color: '#f87171', fontSize: '0.85rem',
                display: 'flex', alignItems: 'center', gap: 8,
              }}>
                <svg width="15" height="15" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0 }}>
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
                    d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
                </svg>
                {error}
              </div>
            )}

            {/* Form */}
            <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>

              {/* Email */}
              <div className="auth-f2">
                <label className="auth-label">Email address</label>
                <input
                  type="email" value={email}
                  onChange={e => setEmail(e.target.value)}
                  className="auth-input"
                  placeholder="you@example.com"
                  autoComplete="email"
                />
              </div>

              {/* Password */}
              <div className="auth-f3">
                <label className="auth-label">Password</label>
                <div style={{ position: 'relative' }}>
                  <input
                    type={showPass ? 'text' : 'password'} value={password}
                    onChange={e => setPassword(e.target.value)}
                    className="auth-input" style={{ paddingRight: 42 }}
                    placeholder="Min. 8 characters" autoComplete="new-password"
                  />
                  <EyeToggle open={showPass} onToggle={() => setShowPass(v => !v)} />
                </div>
              </div>

              {/* Confirm */}
              <div className="auth-f4">
                <label className="auth-label">Confirm password</label>
                <div style={{ position: 'relative' }}>
                  <input
                    type={showConf ? 'text' : 'password'} value={confirmPassword}
                    onChange={e => setConfirmPassword(e.target.value)}
                    className="auth-input" style={{ paddingRight: 42 }}
                    placeholder="Repeat password" autoComplete="new-password"
                  />
                  <EyeToggle open={showConf} onToggle={() => setShowConf(v => !v)} />
                </div>
              </div>

              {/* Submit */}
              <div className="auth-f5" style={{ paddingTop: 4 }}>
                <button type="submit" disabled={loading} className="auth-btn">
                  {loading ? (
                    <>
                      <svg className="animate-spin" width="16" height="16" fill="none" viewBox="0 0 24 24">
                        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
                        <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                      </svg>
                      Creating account…
                    </>
                  ) : (
                    <>
                      Create account
                      <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24" style={{ flexShrink: 0 }}>
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                      </svg>
                    </>
                  )}
                </button>
              </div>
            </form>

            {/* Footer */}
            <p style={{ marginTop: 24, textAlign: 'center', fontSize: '0.8rem', color: 'rgba(255,255,255,0.28)' }}>
              Already have an account?{' '}
              <Link href="/login" style={{ color: 'var(--accent)', fontWeight: 500 }}>
                Sign in
              </Link>
            </p>
          </div>

          <div style={{ height: 1, background: 'linear-gradient(90deg, transparent, rgba(255,255,255,0.04), transparent)' }} />
        </div>
      </div>
    </div>
  );
}

function EyeToggle({ open, onToggle }: { open: boolean; onToggle: () => void }) {
  return (
    <button type="button" className="auth-eye" onClick={onToggle} tabIndex={-1}>
      {open ? (
        <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
            d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
        </svg>
      ) : (
        <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2"
            d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
        </svg>
      )}
    </button>
  );
}