'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { getOrganizations, Organization } from '@/lib/api';

export default function Home() {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchOrganizations = async () => {
      try {
        setLoading(true);
        const data = await getOrganizations();
        setOrganizations(data);
      } catch (err: any) {
        setError(err.message || 'Failed to fetch organizations');
      } finally {
        setLoading(false);
      }
    };

    fetchOrganizations();
  }, []);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        <p className="font-bold">Error</p>
        <p>{error}</p>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Organizations</h1>
          <p className="mt-2 text-gray-600">
            View and manage your organizations and their code quality metrics
          </p>
        </div>
      </div>

      {organizations.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-500 text-lg">No organizations found</p>
          <p className="text-gray-400 mt-2">
            Use the API to add organizations: POST /api/organizations
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {organizations.map((org) => (
            <Link key={org.id} href={`/org/${org.id}`}>
              <div className="card hover:shadow-md transition-shadow cursor-pointer">
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900">
                      {org.name}
                    </h3>
                    <p className="text-sm text-gray-500 mt-1">
                      @{org.github_org_name}
                    </p>
                  </div>
                  <span
                    className={`badge ${
                      org.is_active ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                    }`}
                  >
                    {org.is_active ? 'Active' : 'Inactive'}
                  </span>
                </div>
                <div className="mt-4 pt-4 border-t border-gray-100">
                  <p className="text-sm text-primary-600 font-medium">
                    View Dashboard →
                  </p>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
