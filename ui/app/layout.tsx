import type { Metadata, Viewport } from "next";
import "./globals.css";
import { CinemaProvider } from "@/hooks/useCinema";
import Footer from "@/components/layout/Footer";

export const metadata: Metadata = { title: "Local Cinema" };
export const viewport: Viewport = { themeColor: "#0b0a0c" };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="flex min-h-screen flex-col antialiased">
        <CinemaProvider>
          <div className="flex-1">{children}</div>
          <Footer />
        </CinemaProvider>
      </body>
    </html>
  );
}
