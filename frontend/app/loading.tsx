export default function Loading() {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-white">
      <div className="text-center">
        <div className="relative">
          <div className="w-16 h-16 border-4 border-blue-200 border-t-blue-600 rounded-full animate-spin mx-auto mb-4"></div>
          <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
            <div className="w-8 h-8 bg-blue-600 rounded-full opacity-20 animate-pulse"></div>
          </div>
        </div>
        <h2 className="text-2xl font-bold font-poppins text-gray-900 mb-2">SafeRelief</h2>
        <p className="text-gray-600">Memuat halaman...</p>
      </div>
    </div>
  )
}
