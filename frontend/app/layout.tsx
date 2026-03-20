import type { Metadata } from 'next'
import NavBar from '../components/NavBar'
import './globals.css'

export const metadata: Metadata = {
  title: 'StreamForge',
  description: 'Upload, transcode, and stream high-quality video content seamlessly.',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className="scroll-smooth" suppressHydrationWarning>
      <body className="min-h-screen overflow-x-hidden">

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