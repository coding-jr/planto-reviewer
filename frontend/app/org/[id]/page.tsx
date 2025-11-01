'use client';

import { useEffect, useState } from 'react';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import {
  getOrganization,
  getOrgSummary,
  getLeaderboard,
  getTopIssues,
  Organization,
  OrgSummary,
  LeaderboardEntry,
  TopIssue,
} from '@/lib/api';

export default function OrgDashboard() {
  const params = useParams();
  const orgId = parseInt(params.id as string);

  const [org, setOrg] = useState<Organization | null>(null);
  const [summary, setSummary] = useState<OrgSummary | null>(null);
  const [leaderboard, setLeaderboard] = useState<LeaderboardEntry[]>([]);
  const [topIssues, setTopIssues] = useState<TopIssue[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        const [orgData, summaryData, leaderboardData, issuesData] = await Promise.all([
          getOrganization(orgId),
          getOrgSummary(orgId),
          getLeaderboard(orgId),
          getTopIssues(orgId, 5),
        ]);

        setOrg(orgData);
        setSummary(summaryData);
        setLeaderboard(leaderboardData.leaderboard);
        setTopIssues(issuesData.top_issues);
      } catch (err: any) {
        setError(err.message || 'Failed to fetch data');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [orgId]);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  if (error || !org || !summary) {
    return (
      <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
        <p className="font-bold">Error</p>
        <p>{error || 'Failed to load organization'}</p>
      </div>
    );
  }

  const getSeverityBadge = (severity: string) => {
    const classes = {
      critical: 'badge-critical',
      high: 'badge-high',
      medium: 'badge-medium',
      low: 'badge-low',
    }[severity] || 'badge-low';
    return `badge ${classes}`;
  };

  const getQualityColor = (score: number) => {
    if (score >= 90) return 'text-green-600';
    if (score >= 70) return 'text-yellow-600';
    return 'text-red-600';
  };

  return (
    <div>
      {/* Header */}
      <div className="mb-8">
        <Link href="/" className="text-primary-600 hover:text-primary-700 text-sm mb-2 inline-block">
          ← Back to Organizations
        </Link>
        <h1 className="text-3xl font-bold text-gray-900">{org.name}</h1>
        <p className="text-gray-600 mt-1">@{org.github_org_name}</p>
      </div>

      {/* Summary Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <div className="card">
          <p className="text-sm text-gray-600 mb-1">Total Developers</p>
          <p className="text-3xl font-bold text-gray-900">
            {summary.summary.total_developers}
          </p>
        </div>
        <div className="card">
          <p className="text-sm text-gray-600 mb-1">Total PRs</p>
          <p className="text-3xl font-bold text-gray-900">
            {summary.summary.total_prs}
          </p>
        </div>
        <div className="card">
          <p className="text-sm text-gray-600 mb-1">Total Issues</p>
          <p className="text-3xl font-bold text-red-600">
            {summary.summary.issues_by_severity.reduce((sum, item) => sum + item.count, 0)}
          </p>
        </div>
        <div className="card">
          <p className="text-sm text-gray-600 mb-1">Critical Issues</p>
          <p className="text-3xl font-bold text-red-600">
            {summary.summary.issues_by_severity.find(i => i.severity === 'critical')?.count || 0}
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
        {/* Issues by Type */}
        <div className="card">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Issues by Type</h2>
          <div className="space-y-3">
            {summary.summary.issues_by_type.map((item) => (
              <div key={item.issue_type} className="flex justify-between items-center">
                <span className="text-gray-700 capitalize">
                  {item.issue_type.replace(/_/g, ' ')}
                </span>
                <span className="font-semibold text-gray-900">{item.count}</span>
              </div>
            ))}
            {summary.summary.issues_by_type.length === 0 && (
              <p className="text-gray-400 text-center py-4">No issues found</p>
            )}
          </div>
        </div>

        {/* Issues by Severity */}
        <div className="card">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Issues by Severity</h2>
          <div className="space-y-3">
            {summary.summary.issues_by_severity.map((item) => (
              <div key={item.severity} className="flex justify-between items-center">
                <span className={getSeverityBadge(item.severity)}>
                  {item.severity.toUpperCase()}
                </span>
                <span className="font-semibold text-gray-900">{item.count}</span>
              </div>
            ))}
            {summary.summary.issues_by_severity.length === 0 && (
              <p className="text-gray-400 text-center py-4">No issues found</p>
            )}
          </div>
        </div>
      </div>

      {/* Developer Leaderboard */}
      <div className="card mb-8">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">
          Developer Leaderboard (Top 10)
        </h2>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead>
              <tr className="bg-gray-50">
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Rank
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Developer
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  PRs
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Issues
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase">
                  Quality Score
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {leaderboard.map((dev, index) => (
                <tr key={dev.developer_id} className="hover:bg-gray-50">
                  <td className="px-4 py-3 whitespace-nowrap">
                    <span className="font-semibold text-gray-700">#{index + 1}</span>
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap">
                    <span className="text-gray-900">{dev.github_username}</span>
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap text-gray-700">
                    {dev.total_prs}
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap text-gray-700">
                    {dev.total_issues}
                  </td>
                  <td className="px-4 py-3 whitespace-nowrap">
                    <span className={`font-semibold ${getQualityColor(dev.code_quality_score)}`}>
                      {dev.code_quality_score.toFixed(1)}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {leaderboard.length === 0 && (
            <p className="text-gray-400 text-center py-8">No developers found</p>
          )}
        </div>
      </div>

      {/* Top Issues */}
      <div className="card">
        <h2 className="text-xl font-semibold text-gray-900 mb-4">
          Most Common Issues (Top 5)
        </h2>
        <div className="space-y-4">
          {topIssues.map((issue, index) => (
            <div key={index} className="border-l-4 border-primary-500 pl-4 py-2">
              <div className="flex items-start justify-between mb-2">
                <div className="flex-1">
                  <h3 className="font-semibold text-gray-900">{issue.title}</h3>
                  <p className="text-sm text-gray-600 mt-1">{issue.description}</p>
                </div>
                <div className="flex items-center gap-2 ml-4">
                  <span className={getSeverityBadge(issue.severity)}>
                    {issue.severity}
                  </span>
                  <span className="text-sm font-semibold text-gray-700">
                    {issue.count}x
                  </span>
                </div>
              </div>
              <span className="text-xs text-gray-500 capitalize">
                {issue.issue_type.replace(/_/g, ' ')}
              </span>
            </div>
          ))}
          {topIssues.length === 0 && (
            <p className="text-gray-400 text-center py-8">No issues found</p>
          )}
        </div>
      </div>
    </div>
  );
}
