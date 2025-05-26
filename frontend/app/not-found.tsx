export default function NotFound() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-white">
      <div className="text-center max-w-md mx-auto px-4">
        <div className="mb-8">
          <div className="w-24 h-24 bg-gradient-to-br from-blue-600 to-purple-600 rounded-full flex items-center justify-center mx-auto mb-6">
            <svg className="w-12 h-12 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <h1 className="text-6xl font-bold font-poppins text-gray-900 mb-4">404</h1>
          <h2 className="text-2xl font-bold font-poppins text-gray-700 mb-4">Halaman Tidak Ditemukan</h2>
          <p className="text-gray-600 mb-8">
            Maaf, halaman yang Anda cari tidak dapat ditemukan. Mungkin halaman telah dipindahkan atau URL salah.
          </p>
          <div className="space-y-4">
            <a
              href="/"
              className="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-all duration-300 transform hover:-translate-y-0.5"
            >
              Kembali ke Beranda
            </a>
            <div className="flex justify-center space-x-4 text-sm">
              <a href="/disasters" className="text-blue-600 hover:text-blue-800 transition-colors">
                Lihat Bencana
              </a>
              <span className="text-gray-400">•</span>
              <a href="/contact" className="text-blue-600 hover:text-blue-800 transition-colors">
                Hubungi Kami
              </a>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
