"use client"

import { useEffect, useState } from 'react'
import { useConfig } from '@/hooks/useConfig'

interface InstallationGuideProps {
  markdownContent: string
}

export function InstallationGuide({ markdownContent }: InstallationGuideProps) {
  const [processedContent, setProcessedContent] = useState('')
  const { appName } = useConfig()

  useEffect(() => {
    const processed = markdownContent.replace(/\{\{ appName \}\}/g, appName)
    setProcessedContent(processed)
  }, [markdownContent, appName])

  return (
    <div className="nextra-content">
      <div dangerouslySetInnerHTML={{ __html: processedContent }} />
    </div>
  )
}