import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  rewrites: async () => [
    {
      source: "/v1/:path*",
      destination: `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/v1/:path*`,
    },
    {
      source: "/api/:path*",
      destination: `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/:path*`,
    },
  ],
};

export default nextConfig;
