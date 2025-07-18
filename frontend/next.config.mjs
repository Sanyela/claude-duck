/** @type {import('next').NextConfig} */
const nextConfig = {
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  // SSR 配置优化，防止状态泄露
  experimental: {
    // 禁用内存缓存，防止用户状态混乱
    isrMemoryCacheSize: 0,
  },
  // 禁用ETag生成，确保每次请求独立处理
  generateEtags: false,
  // 生产环境缓存配置
  headers: async () => {
    return [
      {
        source: '/api/:path*',
        headers: [
          {
            key: 'Cache-Control',
            value: 'no-store, no-cache, must-revalidate, proxy-revalidate',
          },
          {
            key: 'Pragma',
            value: 'no-cache',
          },
          {
            key: 'Expires',
            value: '0',
          },
        ],
      },
    ];
  },
}

export default nextConfig
