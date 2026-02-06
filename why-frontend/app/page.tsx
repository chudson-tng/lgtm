import Link from 'next/link'

export default function HomePage() {
  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="text-center space-y-8">
        <h1 className="text-6xl font-bold text-gray-900">Why</h1>
        <p className="text-xl text-gray-600">A messaging app with full observability</p>
        <div className="flex gap-4 justify-center">
          <Link
            href="/signup"
            className="px-6 py-3 bg-blue-600 text-white rounded-md hover:bg-blue-700 font-medium"
          >
            Sign up
          </Link>
          <Link
            href="/login"
            className="px-6 py-3 bg-gray-200 text-gray-900 rounded-md hover:bg-gray-300 font-medium"
          >
            Log in
          </Link>
        </div>
      </div>
    </div>
  )
}
