/**
 * 通用复制函数，支持新版 navigator.clipboard 和旧版 document.execCommand 兼容
 */
export const copyToClipboard = async (text: string): Promise<boolean> => {
  try {
    // 优先使用新版 API
    if (typeof navigator !== 'undefined' && navigator.clipboard && typeof navigator.clipboard.writeText === 'function' && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    }
    
    // 降级到旧版方案
    const textArea = document.createElement('textarea')
    textArea.value = text
    textArea.style.position = 'fixed'
    textArea.style.left = '-999999px'
    textArea.style.top = '-999999px'
    document.body.appendChild(textArea)
    textArea.focus()
    textArea.select()
    
    const success = document.execCommand('copy')
    document.body.removeChild(textArea)
    
    return success
  } catch (err) {
    console.error('复制失败:', err)
    return false
  }
}