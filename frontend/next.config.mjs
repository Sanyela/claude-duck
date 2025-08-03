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
  // 静态导出配置
  output: 'export',
  distDir: 'dist', // 指定输出目录为 dist
  trailingSlash: true,
  generateEtags: false,
  
  // 删除SSR相关配置，不再需要headers设置
  // SPA模式下缓存由nginx或后端控制
};

export default nextConfig;