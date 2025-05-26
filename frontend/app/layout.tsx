import type { Metadata } from "next";
import { Inter, Poppins } from "next/font/google";
import "./globals.css";
import { Navbar } from "./components/Navbar";
import { Footer } from "./components/Footer";
import { Providers } from "./components/Providers";

const inter = Inter({
  variable: "--font-inter",
  subsets: ["latin"],
  display: "swap",
});

const poppins = Poppins({
  variable: "--font-poppins",
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700", "800"],
  display: "swap",
});

export const metadata: Metadata = {
  title: {
    default: "SafeRelief - Platform Donasi Bantuan Bencana Terpercaya",
    template: "%s | SafeRelief"
  },
  description: "Platform donasi bantuan bencana yang aman dan terpercaya. Salurkan bantuan Anda untuk membantu korban bencana di seluruh Indonesia dengan transparansi penuh.",
  keywords: ["donasi", "bantuan bencana", "kemanusiaan", "Indonesia", "relief", "disaster", "charity", "donation"],
  authors: [{ name: "SafeRelief Team" }],
  creator: "SafeRelief",
  publisher: "SafeRelief",
  formatDetection: {
    email: false,
    address: false,
    telephone: false,
  },
  metadataBase: new URL("https://saferelief.com"),
  alternates: {
    canonical: "/",
  },
  openGraph: {
    title: "SafeRelief - Platform Donasi Bantuan Bencana Terpercaya",
    description: "Platform donasi bantuan bencana yang aman dan terpercaya. Salurkan bantuan Anda untuk membantu korban bencana di seluruh Indonesia.",
    url: "https://saferelief.com",
    siteName: "SafeRelief",
    locale: "id_ID",
    type: "website",
    images: [
      {
        url: "/og-image.jpg",
        width: 1200,
        height: 630,
        alt: "SafeRelief - Platform Donasi Bantuan Bencana",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "SafeRelief - Platform Donasi Bantuan Bencana Terpercaya",
    description: "Platform donasi bantuan bencana yang aman dan terpercaya di Indonesia",
    images: ["/og-image.jpg"],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      "max-video-preview": -1,
      "max-image-preview": "large",
      "max-snippet": -1,
    },
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="id" className="scroll-smooth">
      <head>
        <link rel="icon" href="/favicon.ico" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/manifest.json" />
        <meta name="theme-color" content="#2563eb" />
        <meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover" />
      </head>
      <body
        className={`${inter.variable} ${poppins.variable} font-inter antialiased bg-white text-gray-900 selection:bg-blue-100 selection:text-blue-900`}
      >
        <Providers>
          <div className="flex flex-col min-h-screen">
            <Navbar />
            <main className="flex-grow">
              {children}
            </main>
            <Footer />
          </div>
        </Providers>
      </body>
    </html>
  );
}
