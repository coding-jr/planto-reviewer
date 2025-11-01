import axios from 'axios';

const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3000';
const API_KEY = process.env.NEXT_PUBLIC_API_KEY || '';

const api = axios.create({
  baseURL: `${API_URL}/api`,
  headers: {
    'Content-Type': 'application/json',
    ...(API_KEY && { 'Authorization': `Bearer ${API_KEY}` }),
  },
});

export interface Organization {
  id: number;
  name: string;
  github_org_name: string;
  is_active: boolean;
  created_at: string;
}

export interface DeveloperMetrics {
  developer_id: number;
  period: {
    start: string;
    end: string;
  };
  summary: {
    total_prs: number;
    total_issues: number;
    avg_quality_score: number;
  };
  daily_metrics: any[];
}

export interface LeaderboardEntry {
  developer_id: number;
  github_username: string;
  total_prs: number;
  total_issues: number;
  code_quality_score: number;
}

export interface TopIssue {
  issue_type: string;
  severity: string;
  title: string;
  description: string;
  count: number;
}

export interface OrgSummary {
  organization_id: number;
  summary: {
    total_developers: number;
    total_prs: number;
    issues_by_type: Array<{ issue_type: string; count: number }>;
    issues_by_severity: Array<{ severity: string; count: number }>;
  };
}

// Organizations
export const getOrganizations = async (): Promise<Organization[]> => {
  const response = await api.get('/organizations');
  return response.data.data;
};

export const getOrganization = async (id: number): Promise<Organization> => {
  const response = await api.get(`/organizations/${id}`);
  return response.data.data;
};

export const createOrganization = async (data: {
  name: string;
  github_org_name: string;
  github_token: string;
  repos: string[];
}): Promise<Organization> => {
  const response = await api.post('/organizations', data);
  return response.data.data;
};

// Metrics
export const getDeveloperMetrics = async (
  developerId: number,
  startDate?: string,
  endDate?: string
): Promise<DeveloperMetrics> => {
  const params = new URLSearchParams();
  if (startDate) params.append('start_date', startDate);
  if (endDate) params.append('end_date', endDate);

  const response = await api.get(`/metrics/developer/${developerId}?${params}`);
  return response.data;
};

export const getOrgSummary = async (orgId: number): Promise<OrgSummary> => {
  const response = await api.get(`/metrics/organization/${orgId}/summary`);
  return response.data;
};

export const getLeaderboard = async (orgId: number): Promise<{ organization_id: number; leaderboard: LeaderboardEntry[] }> => {
  const response = await api.get(`/metrics/organization/${orgId}/leaderboard`);
  return response.data;
};

export const getTopIssues = async (orgId: number, limit = 10): Promise<{ organization_id: number; top_issues: TopIssue[] }> => {
  const response = await api.get(`/metrics/organization/${orgId}/top-issues?limit=${limit}`);
  return response.data;
};

export default api;
