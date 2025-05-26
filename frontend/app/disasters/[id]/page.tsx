'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { MapContainer, TileLayer, Marker, Popup } from 'react-leaflet';
import 'leaflet/dist/leaflet.css';
import { toast } from 'react-toastify';
import { useAuth } from '../../contexts/AuthContext';

interface DisasterReport {
  id: string;
  reporterId: string;
  title: string;
  description: string;
  latitude: number;
  longitude: number;
  severity: string;
  status: string;
  verifiedBy: string | null;
  createdAt: string;
  updatedAt: string;
  files: {
    id: string;
    filename: string;
    fileSize: number;
    mimeType: string;
    createdAt: string;
  }[];
}

interface Donation {
  id: string;
  donorId: string;
  amount: number;
  currency: string;
  status: string;
  createdAt: string;
}

export default function DisasterDetailPage() {
  const params = useParams();
  const { user } = useAuth();
  const [report, setReport] = useState<DisasterReport | null>(null);
  const [donations, setDonations] = useState<Donation[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch report details
        const reportResponse = await fetch(
          `http://localhost:8080/api/reports/${params.id}`,
          {
            credentials: 'include',
          }
        );
        if (!reportResponse.ok) throw new Error('Failed to fetch report');
        const reportData = await reportResponse.json();
        setReport(reportData);

        // Fetch donations for this report
        const donationsResponse = await fetch(
          `http://localhost:8080/api/donations?reportId=${params.id}`,
          {
            credentials: 'include',
          }
        );
        if (!donationsResponse.ok) throw new Error('Failed to fetch donations');
        const donationsData = await donationsResponse.json();
        setDonations(donationsData);
      } catch (error) {
        toast.error('Error fetching disaster details');
      } finally {
        setLoading(false);
      }
    };

    if (params.id) {
      fetchData();
    }
  }, [params.id]);

  const handleVerify = async () => {
    try {
      const response = await fetch(
        `http://localhost:8080/api/reports/${params.id}/verify`,
        {
          method: 'POST',
          credentials: 'include',
        }
      );

      if (!response.ok) throw new Error('Failed to verify report');

      toast.success('Report verified successfully');
      // Refresh report data
      const updatedReport = await response.json();
      setReport(updatedReport);
    } catch (error) {
      toast.error('Failed to verify report');
    }
  };

  const formatAmount = (amount: number, currency: string) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: currency,
    }).format(amount);
  };

  if (loading || !report) {
    return <div>Loading...</div>;
  }

  const totalDonations = donations.reduce((sum, donation) => {
    if (donation.status === 'completed') {
      return sum + donation.amount;
    }
    return sum;
  }, 0);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-8">
          <div className="bg-white rounded-lg shadow-lg p-6">
            <div className="flex justify-between items-start mb-6">
              <h1 className="text-3xl font-bold">{report.title}</h1>
              <div className="flex gap-2">
                <span
                  className={`px-3 py-1 rounded-full text-sm font-semibold
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
                  {report.severity.charAt(0).toUpperCase() + report.severity.slice(1)} Severity
                </span>
                <span
                  className={`px-3 py-1 rounded-full text-sm font-semibold
                    ${
                      report.status === 'verified'
                        ? 'bg-green-100 text-green-800'
                        : report.status === 'pending'
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-blue-100 text-blue-800'
                    }`}
                >
                  {report.status.charAt(0).toUpperCase() + report.status.slice(1)}
                </span>
              </div>
            </div>

            <p className="text-gray-700 whitespace-pre-wrap mb-6">
              {report.description}
            </p>

            <div className="border-t pt-4">
              <div className="flex justify-between text-sm text-gray-500">
                <span>Reported on {new Date(report.createdAt).toLocaleDateString()}</span>
                {report.verifiedBy && (
                  <span>Verified by admin on {new Date(report.updatedAt).toLocaleDateString()}</span>
                )}
              </div>
            </div>
          </div>

          {/* Map */}
          <div className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Location</h2>
            <div className="h-[400px] rounded-lg overflow-hidden">
              <MapContainer
                center={[report.latitude, report.longitude]}
                zoom={13}
                style={{ height: '100%', width: '100%' }}
              >
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />
                <Marker position={[report.latitude, report.longitude]}>
                  <Popup>{report.title}</Popup>
                </Marker>
              </MapContainer>
            </div>
          </div>

          {/* Images */}
          {report.files.length > 0 && (
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-xl font-semibold mb-4">Images</h2>
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {report.files.map((file) => (
                  <div
                    key={file.id}
                    className="relative aspect-square rounded-lg overflow-hidden"
                  >
                    <img
                      src={`http://localhost:8080/api/uploads/${file.id}`}
                      alt={file.filename}
                      className="object-cover w-full h-full"
                    />
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Admin Actions */}
          {user && report.status === 'pending' && (
            <div className="bg-white rounded-lg shadow-lg p-6">
              <h2 className="text-xl font-semibold mb-4">Admin Actions</h2>
              <button
                onClick={handleVerify}
                className="w-full bg-green-600 text-white px-4 py-2 rounded-md hover:bg-green-700"
              >
                Verify Report
              </button>
            </div>
          )}

          {/* Donation Summary */}
          <div className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Donations</h2>
            <div className="space-y-4">
              <div>
                <p className="text-gray-600">Total Donations</p>
                <p className="text-2xl font-bold">
                  {formatAmount(totalDonations, 'IDR')}
                </p>
              </div>
              <div>
                <p className="text-gray-600">Number of Donors</p>
                <p className="text-2xl font-bold">
                  {donations.filter((d) => d.status === 'completed').length}
                </p>
              </div>
              {report.status === 'verified' && (
                <Link
                  href={`/donate/${report.id}`}
                  className="block w-full text-center bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                >
                  Make a Donation
                </Link>
              )}
            </div>
          </div>

          {/* Recent Donations */}
          <div className="bg-white rounded-lg shadow-lg p-6">
            <h2 className="text-xl font-semibold mb-4">Recent Donations</h2>
            <div className="space-y-4">
              {donations
                .filter((d) => d.status === 'completed')
                .slice(0, 5)
                .map((donation) => (
                  <div
                    key={donation.id}
                    className="flex justify-between items-center"
                  >
                    <div className="text-sm">
                      <p className="font-semibold">
                        {donation.donorId.slice(0, 8)}...
                      </p>
                      <p className="text-gray-500">
                        {new Date(donation.createdAt).toLocaleDateString()}
                      </p>
                    </div>
                    <span className="font-semibold">
                      {formatAmount(donation.amount, donation.currency)}
                    </span>
                  </div>
                ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
