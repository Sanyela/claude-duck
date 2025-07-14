"use client"

import { useEffect, useState, useRef } from "react"

interface AnimatedNumberProps {
  value: number
  duration?: number
  className?: string
  refreshTrigger?: number | string // 用于触发重新动画的任意值
}

export function AnimatedNumber({ value, duration = 600, className = "", refreshTrigger }: AnimatedNumberProps) {
  const [displayValue, setDisplayValue] = useState(0)
  const [isAnimating, setIsAnimating] = useState(false)
  const animationRef = useRef<number>()

  useEffect(() => {
    // 每次 value 或 refreshTrigger 变化时都重新开始动画
    setIsAnimating(true)
    setDisplayValue(0) // 总是从0开始
    
    const startTime = Date.now()
    
    const animate = () => {
      const elapsed = Date.now() - startTime
      const progress = Math.min(elapsed / duration, 1)
      
      // 使用缓动函数 (easeOutQuart)
      const easeProgress = 1 - Math.pow(1 - progress, 4)
      const currentValue = Math.floor(value * easeProgress)
      
      setDisplayValue(currentValue)
      
      if (progress < 1) {
        animationRef.current = requestAnimationFrame(animate)
      } else {
        setDisplayValue(value)
        setIsAnimating(false)
      }
    }

    // 清除之前的动画
    if (animationRef.current) {
      cancelAnimationFrame(animationRef.current)
    }
    
    // 稍微延迟开始动画，确保从0开始显示
    const timeoutId = setTimeout(() => {
      animationRef.current = requestAnimationFrame(animate)
    }, 50)

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current)
      }
      clearTimeout(timeoutId)
    }
  }, [value, duration, refreshTrigger])

  return (
    <span 
      className={`${className} ${isAnimating ? 'transition-all duration-200' : ''}`}
      style={{
        transform: isAnimating ? 'scale(1.05)' : 'scale(1)',
        transition: 'transform 0.2s ease-out'
      }}
    >
      {displayValue.toLocaleString()}
    </span>
  )
}

// 简化的滚动数字组件
export function RollingNumber({ value, className = "" }: { value: number; className?: string }) {
  return (
    <span className={className}>
      {value.toLocaleString()}
    </span>
  )
}