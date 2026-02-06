import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Why - Messaging App',
  description: 'A messaging application with full observability',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
