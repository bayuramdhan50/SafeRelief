import Link from 'next/link';
import { useState } from 'react';
import { useAuth } from '@/hooks/useAuth';

export const Navbar = () => {
  const [isOpen, setIsOpen] = useState(false);
  const { user, logout } = useAuth();

  return (
    <nav className="bg-blue-600">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          <div className="flex items-center">
            <Link href="/" className="text-white text-xl font-bold">
              SafeRelief
            </Link>
          </div>

          <div className="hidden md:block">
            <div className="ml-10 flex items-baseline space-x-4">
              <Link
                href="/"
                className="text-white hover:bg-blue-700 px-3 py-2 rounded-md"
              >
                Home
              </Link>
              <Link
                href="/disasters"
                className="text-white hover:bg-blue-700 px-3 py-2 rounded-md"
              >
                Disasters
              </Link>
              <Link
                href="/donate"
                className="text-white hover:bg-blue-700 px-3 py-2 rounded-md"
              >
                Donate
              </Link>
              {user ? (
                <>
                  <Link
                    href="/dashboard"
                    className="text-white hover:bg-blue-700 px-3 py-2 rounded-md"
                  >
                    Dashboard
                  </Link>
                  <button
                    onClick={logout}
                    className="text-white hover:bg-blue-700 px-3 py-2 rounded-md"
                  >
                    Logout
                  </button>
                </>
              ) : (
                <>
                  <Link
                    href="/login"
                    className="text-white hover:bg-blue-700 px-3 py-2 rounded-md"
                  >
                    Login
                  </Link>
                  <Link
                    href="/register"
                    className="text-white hover:bg-blue-700 px-3 py-2 rounded-md"
                  >
                    Register
                  </Link>
                </>
              )}
            </div>
          </div>

          <div className="md:hidden">
            <button
              onClick={() => setIsOpen(!isOpen)}
              className="text-white hover:bg-blue-700 p-2 rounded-md"
            >
              <svg
                className="h-6 w-6"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                {isOpen ? (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                ) : (
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 6h16M4 12h16m-7 6h7"
                  />
                )}
              </svg>
            </button>
          </div>
        </div>
      </div>

      {isOpen && (
        <div className="md:hidden">
          <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3">
            <Link
              href="/"
              className="text-white hover:bg-blue-700 block px-3 py-2 rounded-md"
            >
              Home
            </Link>
            <Link
              href="/disasters"
              className="text-white hover:bg-blue-700 block px-3 py-2 rounded-md"
            >
              Disasters
            </Link>
            <Link
              href="/donate"
              className="text-white hover:bg-blue-700 block px-3 py-2 rounded-md"
            >
              Donate
            </Link>
            {user ? (
              <>
                <Link
                  href="/dashboard"
                  className="text-white hover:bg-blue-700 block px-3 py-2 rounded-md"
                >
                  Dashboard
                </Link>
                <button
                  onClick={logout}
                  className="text-white hover:bg-blue-700 block px-3 py-2 rounded-md w-full text-left"
                >
                  Logout
                </button>
              </>
            ) : (
              <>
                <Link
                  href="/login"
                  className="text-white hover:bg-blue-700 block px-3 py-2 rounded-md"
                >
                  Login
                </Link>
                <Link
                  href="/register"
                  className="text-white hover:bg-blue-700 block px-3 py-2 rounded-md"
                >
                  Register
                </Link>
              </>
            )}
          </div>
        </div>
      )}
    </nav>
  );
};
