'use client';

import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Link from 'next/link';
import { toast } from 'react-toastify';
import QRCode from 'qrcode.react';

interface UserProfile {
  id: string;
  username: string;
  email: string;
  mfaEnabled: boolean;
}

interface DisasterReport {
  id: string;
  title: string;
  status: string;
  severity: string;
  createdAt: string;
}

interface Donation {
  id: string;
  amount: number;
  currency: string;
  status: string;
  createdAt: string;
  disasterReportId: string;
}

export default function DashboardPage() {
  const { user } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [reports, setReports] = useState<DisasterReport[]>([]);
  const [donations, setDonations] = useState<Donation[]>([]);
  const [mfaSetupData, setMfaSetupData] = useState<{ secret: string; qrCode: string } | null>(null);
  const [showMfaSetup, setShowMfaSetup] = useState(false);
  const [mfaCode, setMfaCode] = useState('');

  useEffect(() => {
    if (!user) return;

    const fetchDashboardData = async () => {
      try {
        // Fetch user profile
        const profileResponse = await fetch('http://localhost:8080/api/users/me', {
          credentials: 'include',
        });
        if (!profileResponse.ok) throw new Error('Failed to fetch profile');
        const profileData = await profileResponse.json();
        setProfile(profileData);

        // Fetch user's disaster reports
        const reportsResponse = await fetch('http://localhost:8080/api/reports?reporter=me', {
          credentials: 'include',
        });
        if (!reportsResponse.ok) throw new Error('Failed to fetch reports');
        const reportsData = await reportsResponse.json();
        setReports(reportsData);

        // Fetch user's donations
        const donationsResponse = await fetch('http://localhost:8080/api/donations', {
          credentials: 'include',
        });
        if (!donationsResponse.ok) throw new Error('Failed to fetch donations');
        const donationsData = await donationsResponse.json();
        setDonations(donationsData);
      } catch (error) {
        toast.error('Error fetching dashboard data');
      }
    };

    fetchDashboardData();
  }, [user]);

  const setupMFA = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/users/me/mfa', {
        method: 'POST',
        credentials: 'include',
      });
      if (!response.ok) throw new Error('Failed to setup MFA');
      const data = await response.json();
      setMfaSetupData(data);
      setShowMfaSetup(true);
    } catch (error) {
      toast.error('Failed to setup MFA');
    }
  };

  const verifyMFA = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/users/me/mfa/verify', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ code: mfaCode }),
      });

      if (!response.ok) throw new Error('Invalid MFA code');

      toast.success('MFA enabled successfully');
      setShowMfaSetup(false);
      setProfile((prev) => prev ? { ...prev, mfaEnabled: true } : null);
    } catch (error) {
      toast.error('Failed to verify MFA code');
    }
  };

  if (!profile) {
    return <div>Loading...</div>;
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Profile Section */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Profile</h2>
            <div className="space-y-4">
              <div>
                <p className="text-gray-600">Username</p>
                <p className="font-semibold">{profile.username}</p>
              </div>
              <div>
                <p className="text-gray-600">Email</p>
                <p className="font-semibold">{profile.email}</p>
              </div>
              <div>
                <p className="text-gray-600">Security</p>
                <div className="mt-2">
                  {profile.mfaEnabled ? (
                    <span className="inline-flex items-center px-3 py-1 rounded-full text-sm font-semibold bg-green-100 text-green-800">
                      MFA Enabled
                    </span>
                  ) : (
                    <button
                      onClick={setupMFA}
                      className="text-blue-600 hover:text-blue-800"
                    >
                      Enable 2FA
                    </button>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* MFA Setup Modal */}
          {showMfaSetup && mfaSetupData && (
            <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4">
              <div className="bg-white rounded-lg p-6 max-w-md w-full">
                <h3 className="text-xl font-semibold mb-4">Setup Two-Factor Authentication</h3>
                <p className="text-gray-600 mb-4">
                  Scan this QR code with your authenticator app:
                </p>
                <div className="flex justify-center mb-4">
                  <QRCode value={mfaSetupData.qrCode} size={200} />
                </div>
                <p className="text-gray-600 mb-2">Or enter this code manually:</p>
                <p className="font-mono bg-gray-100 p-2 rounded mb-4">
                  {mfaSetupData.secret}
                </p>
                <div className="mb-4">
                  <input
                    type="text"
                    value={mfaCode}
                    onChange={(e) => setMfaCode(e.target.value)}
                    placeholder="Enter verification code"
                    className="w-full px-3 py-2 border rounded-md"
                  />
                </div>
                <div className="flex justify-end gap-4">
                  <button
                    onClick={() => setShowMfaSetup(false)}
                    className="px-4 py-2 text-gray-600 hover:text-gray-800"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={verifyMFA}
                    className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
                  >
                    Verify
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Main Content */}
        <div className="lg:col-span-2 space-y-8">
          {/* Disaster Reports */}
          <div className="bg-white rounded-lg shadow-lg p-6">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-semibold">Your Disaster Reports</h2>
              <Link
                href="/disasters/report"
                className="text-blue-600 hover:text-blue-800"
              >
                Report New Disaster
              </Link>
            </div>
            <div className="divide-y">
              {reports.map((report) => (
                <div key={report.id} className="py-4">
                  <div className="flex justify-between items-start">
                    <div>
                      <Link
                        href={`/disasters/${report.id}`}
                        className="font-semibold hover:text-blue-600"
                      >
                        {report.title}
                      </Link>
                      <p className="text-sm text-gray-500">
                        {new Date(report.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="flex gap-2">
                      <span
                        className={`px-2 py-1 rounded-full text-xs font-semibold
                          ${
                            report.severity === 'critical'
                              ? 'bg-red-100 text-red-800'
                              : report.severity === 'high'
                              ? 'bg-orange-100 text-orange-800'
                              : report.severity === 'medium'
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-green-100 text-green-800'
                          }`}
                      >
                        {report.severity}
                      </span>
                      <span
                        className={`px-2 py-1 rounded-full text-xs font-semibold
                          ${
                            report.status === 'verified'
                              ? 'bg-green-100 text-green-800'
                              : report.status === 'pending'
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-blue-100 text-blue-800'
                          }`}
                      >
                        {report.status}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
              {reports.length === 0 && (
                <p className="text-gray-500 py-4">No reports yet</p>
              )}
            </div>
          </div>

          {/* Donations */}
          <div className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Your Donations</h2>
            <div className="divide-y">
              {donations.map((donation) => (
                <div key={donation.id} className="py-4">
                  <div className="flex justify-between items-start">
                    <div>
                      <Link
                        href={`/disasters/${donation.disasterReportId}`}
                        className="font-semibold hover:text-blue-600"
                      >
                        Donation to Disaster Relief
                      </Link>
                      <p className="text-sm text-gray-500">
                        {new Date(donation.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="font-semibold">
                        {new Intl.NumberFormat('id-ID', {
                          style: 'currency',
                          currency: donation.currency,
                        }).format(donation.amount)}
                      </p>
                      <span
                        className={`inline-block px-2 py-1 rounded-full text-xs font-semibold
                          ${
                            donation.status === 'completed'
                              ? 'bg-green-100 text-green-800'
                              : donation.status === 'pending'
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-red-100 text-red-800'
                          }`}
                      >
                        {donation.status}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
              {donations.length === 0 && (
                <p className="text-gray-500 py-4">No donations yet</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
