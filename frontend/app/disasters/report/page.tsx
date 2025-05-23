'use client';

import { useEffect, useState } from 'react';
import { MapContainer, TileLayer, Marker, Popup, useMap } from 'react-leaflet';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { toast } from 'react-toastify';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

// Fix for default marker icons in Leaflet with Next.js
delete (L.Icon.Default.prototype as any)._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: '/marker-icon-2x.png',
  iconUrl: '/marker-icon.png',
  shadowUrl: '/marker-shadow.png',
});

const reportSchema = z.object({
  title: z.string().min(5, 'Title must be at least 5 characters').max(255),
  description: z.string().min(20, 'Description must be at least 20 characters'),
  severity: z.enum(['low', 'medium', 'high', 'critical']),
  images: z
    .any()
    .refine((files) => files?.length <= 5, 'Maximum of 5 files are allowed.')
    .refine(
      (files) => {
        for (const file of files) {
          if (file?.size > 5 * 1024 * 1024) return false;
        }
        return true;
      },
      'Each file must be less than 5MB'
    ),
  latitude: z.number(),
  longitude: z.number(),
});

type ReportFormData = z.infer<typeof reportSchema>;

const MapComponent = ({
  onLocationSelect,
}: {
  onLocationSelect: (lat: number, lng: number) => void;
}) => {
  const map = useMap();

  useEffect(() => {
    map.on('click', (e) => {
      onLocationSelect(e.latlng.lat, e.latlng.lng);
    });
  }, [map, onLocationSelect]);

  return null;
};

export default function ReportDisasterPage() {
  const [position, setPosition] = useState<[number, number] | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<ReportFormData>({
    resolver: zodResolver(reportSchema),
  });

  useEffect(() => {
    // Get user's location if they allow it
    navigator.geolocation.getCurrentPosition(
      (position) => {
        setPosition([position.coords.latitude, position.coords.longitude]);
      },
      () => {
        // Default to Indonesia's coordinates if location access is denied
        setPosition([-6.2088, 106.8456]);
      }
    );
  }, []);

  const handleLocationSelect = (lat: number, lng: number) => {
    setValue('latitude', lat);
    setValue('longitude', lng);
    setPosition([lat, lng]);
  };

  const onSubmit = async (data: ReportFormData) => {
    try {
      setIsSubmitting(true);

      const formData = new FormData();
      formData.append('title', data.title);
      formData.append('description', data.description);
      formData.append('severity', data.severity);
      formData.append('latitude', String(data.latitude));
      formData.append('longitude', String(data.longitude));

      // Append images if any
      if (data.images) {
        Array.from(data.images).forEach((file) => {
          formData.append('images', file);
        });
      }

      const response = await fetch('http://localhost:8080/api/reports', {
        method: 'POST',
        credentials: 'include',
        body: formData,
      });

      if (!response.ok) {
        throw new Error('Failed to submit report');
      }

      toast.success('Disaster report submitted successfully');
    } catch (error: any) {
      toast.error(error.message || 'Failed to submit report');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!position) {
    return <div>Loading map...</div>;
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <h1 className="text-3xl font-bold mb-8">Report a Disaster</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div>
            <label className="block text-sm font-medium text-gray-700">
              Title
            </label>
            <input
              type="text"
              {...register('title')}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            />
            {errors.title && (
              <p className="mt-1 text-sm text-red-600">{errors.title.message}</p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              Description
            </label>
            <textarea
              {...register('description')}
              rows={4}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            />
            {errors.description && (
              <p className="mt-1 text-sm text-red-600">
                {errors.description.message}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              Severity
            </label>
            <select
              {...register('severity')}
              className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
            {errors.severity && (
              <p className="mt-1 text-sm text-red-600">
                {errors.severity.message}
              </p>
            )}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700">
              Images (Max 5 files, 5MB each)
            </label>
            <input
              type="file"
              multiple
              accept="image/*"
              {...register('images')}
              className="mt-1 block w-full"
            />
            {errors.images && (
              <p className="mt-1 text-sm text-red-600">
                {errors.images.message as string}
              </p>
            )}
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            {isSubmitting ? 'Submitting...' : 'Submit Report'}
          </button>
        </form>

        <div className="h-[600px] rounded-lg overflow-hidden shadow-lg">
          <MapContainer
            center={position}
            zoom={13}
            style={{ height: '100%', width: '100%' }}
          >
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
            {position && (
              <Marker position={position}>
                <Popup>Disaster location</Popup>
              </Marker>
            )}
            <MapComponent onLocationSelect={handleLocationSelect} />
          </MapContainer>
        </div>
      </div>
    </div>
  );
}
