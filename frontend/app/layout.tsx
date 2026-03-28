import type { Metadata } from 'next'
import NavBar from '../components/NavBar'
import './globals.css'

export const metadata: Metadata = {
  title: 'StreamForge',
  description: 'Upload, transcode, and stream high-quality video content seamlessly.',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  // Read APP_ENV at runtime (not baked in at build time)
  const appEnv = process.env.APP_ENV ?? 'development'

  return (
    <html lang="en" className="scroll-smooth" suppressHydrationWarning>
      <body className="min-h-screen overflow-x-hidden">

        {/* Inject APP_ENV for client-side logger — runtime configurable, no rebuild needed */}
        <script
          dangerouslySetInnerHTML={{
            __html: `window.__SF_APP_ENV=${JSON.stringify(appEnv)};`,
          }}
        />

        {/* Fixed ambient background — sits behind everything */}
        <div className="fixed inset-0 -z-10 bg-[#080808]" />
        <div className="fixed inset-0 -z-10 bg-grid opacity-100" />
        <div className="fixed inset-0 -z-10 bg-ambient" />

        <NavBar />

        <main className="min-h-screen" style={{ paddingTop: 'var(--nav-h)' }}>
          {children}
        </main>

      </body>
    </html>
  )
}