import type { Metadata } from 'next'
import NavBar from '../components/NavBar'
import './globals.css'

export const metadata: Metadata = {
  title: 'StreamForge',
  description: 'Upload, transcode, and stream high-quality video content seamlessly.',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  const publicApiUrl = process.env.PUBLIC_API_URL || process.env["NEXT_PUBLIC_API_URL"] || "http://localhost:8080";
  const envScript = `
    window.__ENV = {
      PUBLIC_API_URL: "${publicApiUrl}"
    };
  `;

  return (
    <html lang="en" className="scroll-smooth">
      <head>
        <script dangerouslySetInnerHTML={{ __html: envScript }} />
      </head>
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