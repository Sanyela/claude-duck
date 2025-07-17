"use client"

import { useEffect, useState } from 'react'
import { getAppName } from '@/lib/env'

interface InstallationGuideProps {
  markdownContent: string
}

export function InstallationGuide({ markdownContent }: InstallationGuideProps) {
  const [processedContent, setProcessedContent] = useState('')

  useEffect(() => {
    const appName = getAppName()
    const processed = markdownContent.replace(/\{\{ appName \}\}/g, appName)
    setProcessedContent(processed)
  }, [markdownContent])

  return (
    <div className="nextra-content">
      <div dangerouslySetInnerHTML={{ __html: processedContent }} />
    </div>
  )
}