export const dynamic = 'force-dynamic'

import type { Metadata } from 'next'
import './globals.css'
import { ThemeProvider } from '@/components/theme-provider'
import { AuthProvider } from '@/contexts/AuthContext'
import { AuthGuard } from '@/components/auth/auth-guard'
import { Toaster } from '@/components/ui/toaster'
import { getServerAppName } from '@/lib/env'

const appName = getServerAppName()

export const metadata: Metadata = {
  title: appName,
  description: `${appName} 是一款基于 Claude 4 Sonnet 的 AI 助手，旨在为用户提供高效、便捷的 AI 服务。`,
  generator: `${appName} Team`,
  icons: {
    icon: '/icon.png',
  },
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          <AuthProvider>
            <AuthGuard>
              {children}
            </AuthGuard>
            <Toaster />
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}
