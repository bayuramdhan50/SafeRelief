'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'react-toastify';

const donationSchema = z.object({
  amount: z.number().min(1, 'Amount must be at least 1'),
  currency: z.string(),
  description: z.string().optional(),
  paymentMethod: z.enum(['credit_card', 'bank_transfer', 'e_wallet']),
});

type DonationFormData = z.infer<typeof donationSchema>;

interface DisasterReport {
  id: string;
  title: string;
  description: string;
  severity: string;
  status: string;
  createdAt: string;
}

export default function DonatePage() {
  const params = useParams();
  const [report, setReport] = useState<DisasterReport | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<DonationFormData>({
    resolver: zodResolver(donationSchema),
    defaultValues: {
      currency: 'IDR',
      paymentMethod: 'bank_transfer',
    },
  });

  useEffect(() => {
    const fetchReport = async () => {
      try {
        const response = await fetch(
          `http://localhost:8080/api/reports/${params.id}`,
          {
            credentials: 'include',
          }
        );
        if (!response.ok) throw new Error('Failed to fetch report');
        const data = await response.json();
        setReport(data);
      } catch (error) {
        toast.error('Error fetching disaster report');
      }
    };

    if (params.id) {
      fetchReport();
    }
  }, [params.id]);

  const onSubmit = async (data: DonationFormData) => {
    try {
      setIsSubmitting(true);
      const response = await fetch('http://localhost:8080/api/donations', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({
          ...data,
          disasterReportId: params.id,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to process donation');
      }

      const result = await response.json();
      toast.success('Donation processed successfully!');
      // Redirect to payment gateway or confirmation page
    } catch (error) {
      toast.error('Failed to process donation. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!report) {
    return <div>Loading...</div>;
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <div className="bg-white rounded-lg shadow-lg p-6 mb-8">
        <h1 className="text-2xl font-bold mb-4">Donation Details</h1>
        <div className="mb-6">
          <h2 className="text-xl font-semibold">{report.title}</h2>
          <p className="text-gray-600 mt-2">{report.description}</p>
          <div className="mt-4">
            <span className={`inline-block px-3 py-1 rounded-full text-sm font-semibold
              ${report.severity === 'critical' ? 'bg-red-100 text-red-800' :
              report.severity === 'high' ? 'bg-orange-100 text-orange-800' :
              report.severity === 'medium' ? 'bg-yellow-100 text-yellow-800' :
              'bg-green-100 text-green-800'}`}>
              {report.severity.charAt(0).toUpperCase() + report.severity.slice(1)} Severity
            </span>
          </div>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700">
              Amount
            </label>
            <div className="mt-1 relative rounded-md shadow-sm">
              <input
                type="number"
                {...register('amount', { valueAsNumber: true })}
                className="block w-full pr-12 pl-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                placeholder="0.00"
              />
              <div className="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
                <span className="text-gray-500 sm:text-sm">IDR</span>
              </div>
            </div>
            {errors.amount && (
              <p className="mt-1 text-sm text-red-600">{errors.amount.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              Payment Method
            </label>
            <select
              {...register('paymentMethod')}
              className="mt-1 block w-full pl-3 pr-10 py-2 text-base border border-gray-300 focus:outline-none focus:ring-blue-500 focus:border-blue-500 rounded-md"
            >
              <option value="bank_transfer">Bank Transfer</option>
              <option value="credit_card">Credit Card</option>
              <option value="e_wallet">E-Wallet</option>
            </select>
            {errors.paymentMethod && (
              <p className="mt-1 text-sm text-red-600">
                {errors.paymentMethod.message}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              Message (Optional)
            </label>
            <textarea
              {...register('description')}
              rows={3}
              className="mt-1 block w-full border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
              placeholder="Add a message of support..."
            />
            {errors.description && (
              <p className="mt-1 text-sm text-red-600">
                {errors.description.message}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            {isSubmitting ? 'Processing...' : 'Make Donation'}
          </button>
        </form>
      </div>

      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 className="text-blue-800 font-medium">Secure Transaction</h3>
        <p className="text-blue-600 text-sm mt-1">
          Your donation is protected with bank-level security. We never store your
          payment details.
        </p>
      </div>
    </div>
  );
}
