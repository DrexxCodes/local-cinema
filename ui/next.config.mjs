/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "export", // static files embedded into the Go binary
  trailingSlash: true,
  images: { unoptimized: true },
  // In `npm run dev`, forward API calls to the Go server.
  async rewrites() {
    if (process.env.NODE_ENV !== "development") return [];
    const t = process.env.CINEMA_API || "http://localhost:8765";
    return [
      { source: "/api/:path*", destination: `${t}/api/:path*` },
      { source: "/stream/:path*", destination: `${t}/stream/:path*` },
    ];
  },
};
export default nextConfig;
