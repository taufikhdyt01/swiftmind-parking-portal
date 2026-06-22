import type { NextConfig } from "next";

// All browser calls go to /api/* on the Next.js origin and are proxied here to
// the API Gateway (the single backend entrypoint). Keeping it same-origin means
// the httpOnly auth cookie is first-party and no CORS handshake is needed.
const GATEWAY_URL = process.env.GATEWAY_URL ?? "http://localhost:8080";

const nextConfig: NextConfig = {
  // Pin the workspace root to this app (a stray lockfile exists above the repo).
  turbopack: { root: __dirname },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${GATEWAY_URL}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
