import React from 'react'
import { DocsThemeConfig } from 'nextra-theme-docs'
import { getAppName } from './lib/env'

const config: DocsThemeConfig = {
  logo: (
    <span style={{ 
      fontWeight: 600, 
      fontSize: '1.2rem',
      background: 'linear-gradient(90deg, #007AFF 0%, #5856D6 100%)',
      WebkitBackgroundClip: 'text',
      WebkitTextFillColor: 'transparent',
      backgroundClip: 'text'
    }}>
      {getAppName()} 安装教程
    </span>
  ),
  project: {
    link: 'https://github.com/anthropics/claude-code',
  },
  chat: {
    link: 'https://discord.gg/claude',
  },
  docsRepositoryBase: 'https://github.com/anthropics/claude-code',
  footer: {
    text: (
      <span>
        © 2024{' '}
        <a href="https://anthropic.com" target="_blank">
          {getAppName()}
        </a>
        . 保留所有权利。
      </span>
    ),
  },
  useNextSeoProps() {
    return {
      titleTemplate: `%s – ${getAppName()} 安装教程`
    }
  },
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta property="og:title" content={`${getAppName()} 安装教程`} />
      <meta property="og:description" content="详细的安装和配置指南" />
      <style jsx global>{`
        /* macOS 风格自定义样式 */
        :root {
          --nextra-primary-hue: 212deg;
          --nextra-navbar-height: 4rem;
          --nextra-sidebar-width: 20rem;
        }
        
        .nextra-nav-container {
          backdrop-filter: blur(20px) saturate(190%) contrast(70%) brightness(80%);
          background: rgba(255, 255, 255, 0.8);
          border-bottom: 1px solid rgba(0, 0, 0, 0.1);
          box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
        
        .dark .nextra-nav-container {
          background: rgba(30, 30, 30, 0.8);
          border-bottom: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .nextra-sidebar {
          background: rgba(249, 249, 249, 0.8);
          backdrop-filter: blur(20px);
          border-right: 1px solid rgba(0, 0, 0, 0.1);
        }
        
        .dark .nextra-sidebar {
          background: rgba(28, 28, 28, 0.8);
          border-right: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        .nextra-content {
          background: rgba(255, 255, 255, 0.6);
          backdrop-filter: blur(20px);
        }
        
        .dark .nextra-content {
          background: rgba(17, 17, 17, 0.6);
        }
        
        /* 按钮样式 */
        .nextra-button {
          background: linear-gradient(135deg, #007AFF 0%, #5856D6 100%);
          border: none;
          border-radius: 8px;
          color: white;
          padding: 8px 16px;
          font-weight: 500;
          transition: all 0.2s ease;
          box-shadow: 0 1px 3px rgba(0, 122, 255, 0.3);
        }
        
        .nextra-button:hover {
          transform: translateY(-1px);
          box-shadow: 0 4px 12px rgba(0, 122, 255, 0.4);
        }
        
        /* 代码块样式 */
        .nextra-code-block {
          background: rgba(28, 28, 28, 0.95);
          border-radius: 12px;
          border: 1px solid rgba(255, 255, 255, 0.1);
          backdrop-filter: blur(20px);
        }
        
        /* 圆角和阴影 */
        .nextra-card {
          border-radius: 12px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
          backdrop-filter: blur(20px);
        }
        
        .dark .nextra-card {
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
        }
        
        /* 搜索框 */
        .nextra-search input {
          background: rgba(255, 255, 255, 0.8);
          border: 1px solid rgba(0, 0, 0, 0.1);
          border-radius: 8px;
          backdrop-filter: blur(20px);
        }
        
        .dark .nextra-search input {
          background: rgba(60, 60, 60, 0.8);
          border: 1px solid rgba(255, 255, 255, 0.1);
        }
        
        /* 滚动条样式 */
        ::-webkit-scrollbar {
          width: 8px;
          height: 8px;
        }
        
        ::-webkit-scrollbar-track {
          background: rgba(0, 0, 0, 0.1);
          border-radius: 4px;
        }
        
        ::-webkit-scrollbar-thumb {
          background: rgba(0, 0, 0, 0.3);
          border-radius: 4px;
        }
        
        ::-webkit-scrollbar-thumb:hover {
          background: rgba(0, 0, 0, 0.5);
        }
        
        .dark ::-webkit-scrollbar-track {
          background: rgba(255, 255, 255, 0.1);
        }
        
        .dark ::-webkit-scrollbar-thumb {
          background: rgba(255, 255, 255, 0.3);
        }
        
        .dark ::-webkit-scrollbar-thumb:hover {
          background: rgba(255, 255, 255, 0.5);
        }
      `}</style>
    </>
  ),
  primaryHue: 212,
  primarySaturation: 100,
  sidebar: {
    defaultMenuCollapseLevel: 1,
    autoCollapse: true,
  },
  toc: {
    backToTop: true,
  },
  editLink: {
    text: '在 GitHub 上编辑此页面',
  },
  feedback: {
    content: '有问题？给我们反馈 →',
    labels: 'feedback',
  },
  gitTimestamp: '最后更新于',
  navigation: {
    prev: true,
    next: true,
  },
  darkMode: true,
  nextThemes: {
    defaultTheme: 'system',
  },
}

export default config