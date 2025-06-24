"use client"

import { useEffect, useState } from "react"

export function Greeting({ userName }: { userName: string }) {
  const [greeting, setGreeting] = useState("")

  useEffect(() => {
    const getCurrentGreeting = () => {
      const hour = new Date().getHours()
      if (hour < 12) {
        return "早上好"
      } else if (hour < 18) {
        return "中午好"
      } else {
        return "晚上好"
      }
    }
    setGreeting(getCurrentGreeting())
  }, [])

  return (
    <div className="mb-6">
      <h1 className="text-3xl font-bold text-foreground">
        <span role="img" aria-label="waving hand">
          👋
        </span>{" "}
        {greeting}, {userName}
      </h1>
    </div>
  )
}
