'use client'
 
export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <html>
      <body>
        <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-red-50 to-white">
          <div className="text-center max-w-md mx-auto px-4">
            <div className="mb-8">
              <div className="w-24 h-24 bg-gradient-to-br from-red-600 to-red-700 rounded-full flex items-center justify-center mx-auto mb-6">
                <svg className="w-12 h-12 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.314 15.5c-.77.833.192 2.5 1.732 2.5z" />
                </svg>
              </div>
              <h1 className="text-4xl font-bold font-poppins text-gray-900 mb-4">Oops!</h1>
              <h2 className="text-xl font-bold font-poppins text-gray-700 mb-4">Terjadi Kesalahan</h2>
              <p className="text-gray-600 mb-8">
                Maaf, terjadi kesalahan yang tidak terduga. Tim kami telah diberitahu dan sedang memperbaikinya.
              </p>
              <div className="space-y-4">
                <button
                  onClick={reset}
                  className="inline-block bg-blue-600 text-white px-8 py-3 rounded-lg font-semibold hover:bg-blue-700 transition-all duration-300 transform hover:-translate-y-0.5"
                >
                  Coba Lagi
                </button>
                <div className="flex justify-center space-x-4 text-sm">
                  <a href="/" className="text-blue-600 hover:text-blue-800 transition-colors">
                    Kembali ke Beranda
                  </a>
                  <span className="text-gray-400">•</span>
                  <a href="/contact" className="text-blue-600 hover:text-blue-800 transition-colors">
                    Laporkan Masalah
                  </a>
                </div>
              </div>
            </div>
          </div>
        </div>
      </body>
    </html>
  )
}
