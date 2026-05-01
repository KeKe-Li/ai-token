import type { NextConfig } from "next";

const apiUrl = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const nextConfig: NextConfig = {
  output: "standalone",
  rewrites: async () => [
    {
      source: "/v1/:path*",
      destination: `${apiUrl}/v1/:path*`,
    },
    {
      source: "/api/:path*",
      destination: `${apiUrl}/api/:path*`,
    },
  ],
};

export default nextConfig;
