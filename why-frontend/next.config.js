/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone', // Required for Docker deployment
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'loki-minio.observability.svc.cluster.local',
      },
    ],
  },
}

module.exports = nextConfig
