"use client"

import { useEffect, useState } from "react"
import { TypingEffect } from "./TypingEffect"

export function Greeting({ userName }: { userName: string }) {
  const [greeting, setGreeting] = useState("")
  const [hitokoto, setHitokoto] = useState("天生我材必有用，千金散尽还复来。")
  const [displayText, setDisplayText] = useState("")
  const [source, setSource] = useState("—— 将进酒")
  const [showSource, setShowSource] = useState(false)

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
    
    const runTypingEffect = (text: string, from?: string) => {
      setDisplayText("") // 清空显示文本
      setShowSource(false)
      
      // 构造完整文本，包括引文和来源
      const fullText = from ? `「 ${text} 」 —— ${from}` : `「 ${text} 」`
      
      // 使用TypingEffect实现打字机效果
      new TypingEffect(
        fullText, 
        800, 
        true, // 启用光标
        setDisplayText
      ).run()
    }
    
    // 从每日一言API获取数据
    fetch("https://v1.hitokoto.cn/?c=i")
      .then(response => response.json())
      .then(data => {
        if (data && data.hitokoto) {
          setHitokoto(data.hitokoto)
          runTypingEffect(data.hitokoto, data.from)
        }
      })
      .catch(error => {
        console.error("Error fetching hitokoto:", error)
        // 如果API失败，使用默认文本
        runTypingEffect("天生我材必有用，千金散尽还复来。", "将进酒")
      })
  }, [])

  return (
    <div className="mb-6">
      <h1 className="text-3xl font-bold text-foreground">
        <span role="img" aria-label="waving hand">
          👋
        </span>{" "}
        {greeting}, {userName}
      </h1>
      <div className="mt-2">
        <p className="text-base text-muted-foreground">{displayText}</p>
      </div>
    </div>
  )
}
