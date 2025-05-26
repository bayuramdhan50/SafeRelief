import Image from "next/image";
import Link from "next/link";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "SafeRelief - Platform Donasi Bantuan Bencana Terpercaya",
  description: "Platform donasi bantuan bencana yang aman dan terpercaya. Salurkan bantuan Anda untuk membantu korban bencana di seluruh Indonesia dengan transparansi penuh.",
  keywords: ["donasi", "bantuan bencana", "kemanusiaan", "Indonesia", "relief", "disaster", "charity", "donation"],
  openGraph: {
    title: "SafeRelief - Platform Donasi Bantuan Bencana Terpercaya",
    description: "Platform donasi bantuan bencana yang aman dan terpercaya di Indonesia",
    images: ["/og-image.jpg"],
  },
};

export default function Home() {
  return (
    <div className="overflow-x-hidden">
      {/* Hero Section */}
      <section className="relative min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-900 via-blue-800 to-purple-900">
        {/* Background Pattern */}
        <div className="absolute inset-0 opacity-10">
          <div className="absolute inset-0" style={{
            backgroundImage: `url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.1'%3E%3Ccircle cx='20' cy='20' r='2'/%3E%3Ccircle cx='40' cy='40' r='2'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E")`,
          }} />
        </div>
        
        {/* Floating Elements */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-20 left-20 w-72 h-72 bg-blue-400 rounded-full opacity-10 animate-pulse"></div>
          <div className="absolute bottom-20 right-20 w-96 h-96 bg-purple-400 rounded-full opacity-10 animate-pulse" style={{ animationDelay: '1s' }}></div>
          <div className="absolute top-1/2 left-10 w-48 h-48 bg-white rounded-full opacity-5 animate-pulse" style={{ animationDelay: '2s' }}></div>
        </div>

        <div className="container-custom relative z-10">
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 lg:gap-16 items-center">
            {/* Hero Text */}
            <div className="text-center lg:text-left animate-fade-in-up">
              <div className="inline-flex items-center space-x-2 bg-white/10 backdrop-blur-md border border-white/20 rounded-full px-4 py-2 mb-6">
                <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                <span className="text-white/90 text-sm font-medium">Platform Terpercaya #1 di Indonesia</span>
              </div>
              
              <h1 className="text-4xl sm:text-5xl lg:text-7xl font-bold font-poppins text-white leading-tight mb-6">
                Bersama Kita
                <span className="block bg-gradient-to-r from-blue-400 to-purple-400 bg-clip-text text-transparent">
                  Bantu Sesama
                </span>
              </h1>
              
              <p className="text-xl text-blue-100 mb-8 leading-relaxed">
                SafeRelief menghubungkan kebaikan hati Anda dengan mereka yang membutuhkan. 
                Platform donasi bantuan bencana yang aman, transparan, dan terpercaya.
              </p>
              
              <div className="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start">
                <Link
                  href="/disasters"
                  className="btn-primary text-lg px-8 py-4"
                >
                  Mulai Donasi
                </Link>
                <Link
                  href="/disasters/report"
                  className="btn-secondary text-lg px-8 py-4"
                >
                  Laporkan Bencana
                </Link>
              </div>
              
              {/* Stats */}
              <div className="grid grid-cols-3 gap-8 mt-12 pt-8 border-t border-white/20">
                <div className="text-center">
                  <div className="text-3xl font-bold text-white mb-2">1000+</div>
                  <div className="text-blue-200 text-sm">Bantuan Tersalurkan</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-white mb-2">50+</div>
                  <div className="text-blue-200 text-sm">Lokasi Bencana</div>
                </div>
                <div className="text-center">
                  <div className="text-3xl font-bold text-white mb-2">24/7</div>
                  <div className="text-blue-200 text-sm">Siaga Darurat</div>
                </div>
              </div>
            </div>

            {/* Hero Image */}
            <div className="relative animate-fade-in-up" style={{ animationDelay: '0.2s' }}>
              <div className="relative z-10">
                <div className="bg-white/10 backdrop-blur-md border border-white/20 rounded-3xl p-8 shadow-2xl">
                  <div className="aspect-square bg-gradient-to-br from-blue-500/20 to-purple-500/20 rounded-2xl flex items-center justify-center mb-6">
                    <svg className="w-32 h-32 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1} d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                    </svg>
                  </div>
                  <h3 className="text-2xl font-bold text-white mb-4 text-center">Bantuan Real-time</h3>
                  <p className="text-blue-100 text-center leading-relaxed">
                    Monitor perkembangan bantuan Anda secara real-time dengan sistem tracking yang canggih
                  </p>
                </div>
              </div>
              
              {/* Floating Cards */}
              <div className="absolute -top-4 -right-4 bg-green-500 text-white p-4 rounded-xl shadow-lg animate-bounce">
                <div className="text-sm font-semibold">✓ Verified</div>
              </div>
              <div className="absolute -bottom-4 -left-4 bg-yellow-500 text-white p-4 rounded-xl shadow-lg animate-bounce" style={{ animationDelay: '0.5s' }}>
                <div className="text-sm font-semibold">🚨 Emergency</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-gray-50">
        <div className="container-custom">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-gray-900 mb-6">Mengapa Memilih SafeRelief?</h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              Platform yang dirancang khusus untuk memastikan bantuan Anda sampai kepada yang membutuhkan 
              dengan sistem yang aman dan transparan
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            <div className="bg-white p-8 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-2">
              <div className="w-16 h-16 bg-blue-100 rounded-2xl flex items-center justify-center mb-6">
                <svg className="w-8 h-8 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-4">Aman & Terpercaya</h3>
              <p className="text-gray-600 leading-relaxed">
                Sistem keamanan berlapis dan transparansi penuh dalam penyaluran donasi untuk memastikan 
                bantuan sampai ke tangan yang tepat.
              </p>
            </div>

            <div className="bg-white p-8 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-2">
              <div className="w-16 h-16 bg-green-100 rounded-2xl flex items-center justify-center mb-6">
                <svg className="w-8 h-8 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-4">Respons Cepat</h3>
              <p className="text-gray-600 leading-relaxed">
                Bantuan langsung tersalurkan ke lokasi bencana dengan koordinasi tim di lapangan 
                yang siap 24/7 untuk tanggap darurat.
              </p>
            </div>

            <div className="bg-white p-8 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 hover:-translate-y-2">
              <div className="w-16 h-16 bg-purple-100 rounded-2xl flex items-center justify-center mb-6">
                <svg className="w-8 h-8 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <h3 className="text-2xl font-bold text-gray-900 mb-4">Terverifikasi</h3>
              <p className="text-gray-600 leading-relaxed">
                Setiap laporan bencana diverifikasi oleh tim khusus untuk memastikan keabsahan 
                dan kebutuhan bantuan yang sebenarnya.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* How it Works Section */}
      <section className="py-20 bg-white">
        <div className="container-custom">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold text-gray-900 mb-6">Cara Kerja SafeRelief</h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              Proses yang sederhana dan transparan untuk memastikan bantuan Anda tepat sasaran
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
            <div className="text-center group">
              <div className="w-20 h-20 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                <span className="text-2xl font-bold text-white">1</span>
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-4">Laporan Masuk</h3>
              <p className="text-gray-600">Laporan bencana diterima dari masyarakat atau tim lapangan</p>
            </div>

            <div className="text-center group">
              <div className="w-20 h-20 bg-gradient-to-br from-green-500 to-green-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                <span className="text-2xl font-bold text-white">2</span>
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-4">Verifikasi</h3>
              <p className="text-gray-600">Tim khusus memverifikasi keabsahan dan tingkat kebutuhan bantuan</p>
            </div>

            <div className="text-center group">
              <div className="w-20 h-20 bg-gradient-to-br from-purple-500 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                <span className="text-2xl font-bold text-white">3</span>
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-4">Penggalangan</h3>
              <p className="text-gray-600">Campaign donasi dibuka untuk masyarakat luas</p>
            </div>

            <div className="text-center group">
              <div className="w-20 h-20 bg-gradient-to-br from-orange-500 to-orange-600 rounded-full flex items-center justify-center mx-auto mb-6 group-hover:scale-110 transition-transform duration-300">
                <span className="text-2xl font-bold text-white">4</span>
              </div>
              <h3 className="text-xl font-bold text-gray-900 mb-4">Penyaluran</h3>
              <p className="text-gray-600">Bantuan disalurkan langsung ke lokasi dengan monitoring real-time</p>
            </div>
          </div>
        </div>
      </section>

      {/* Call to Action */}
      <section className="bg-gradient-to-r from-blue-600 to-purple-600 text-white py-20">
        <div className="container-custom text-center">
          <h2 className="text-4xl font-bold mb-6">Mulai Berbagi Sekarang</h2>
          <p className="text-xl mb-8 text-blue-100">Setiap bantuan yang Anda berikan sangat berarti bagi mereka yang membutuhkan</p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/disasters"
              className="btn-secondary"
            >
              Lihat Bencana Terkini
            </Link>
            <Link
              href="/register"
              className="bg-white text-blue-600 px-8 py-4 rounded-xl font-semibold hover:bg-blue-50 transition-all duration-300 shadow-lg hover:shadow-xl"
            >
              Bergabung Sekarang
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
