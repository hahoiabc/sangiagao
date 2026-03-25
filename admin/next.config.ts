import type { NextConfig } from "next";

const isProd = process.env.NODE_ENV === "production";

const nextConfig: NextConfig = {
  output: "standalone",
  images: {
    remotePatterns: [
      {
        protocol: "https",
        hostname: "sangiagao.vn",
        pathname: "/images/**",
      },
      ...(isProd
        ? []
        : [
            {
              protocol: "http" as const,
              hostname: "localhost",
              port: "9000",
              pathname: "/rice-images/**",
            },
          ]),
    ],
  },
  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "X-XSS-Protection", value: "1; mode=block" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
          // CSP is set by nginx
        ],
      },
    ];
  },
};

export default nextConfig;
