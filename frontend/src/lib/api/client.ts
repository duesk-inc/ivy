import axios from 'axios';
import type {
  LoginResponse,
  User,
  MatchingRequest,
  MatchingResponse,
  MatchingListResponse,
  FileParseResponse,
  SettingsResponse,
  SupplementInfo,
  JobGroup,
} from '../../types';

const TOKEN_KEY = 'ivy_access_token';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8081/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor: attach Bearer token
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem(TOKEN_KEY);
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor: redirect to /login on 401
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem(TOKEN_KEY);
      window.location.href = '/login';
    }
    return Promise.reject(error);
  },
);

// ---------- Auth ----------

export async function login(email: string, password: string): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>('/auth/login', { email, password });
  localStorage.setItem(TOKEN_KEY, data.access_token);
  return data;
}

export async function refreshToken(token: string): Promise<LoginResponse> {
  const { data } = await apiClient.post<LoginResponse>('/auth/refresh', { refresh_token: token });
  localStorage.setItem(TOKEN_KEY, data.access_token);
  return data;
}

export async function logout(): Promise<void> {
  try {
    await apiClient.post('/auth/logout');
  } finally {
    localStorage.removeItem(TOKEN_KEY);
  }
}

export async function getMe(): Promise<User> {
  const { data } = await apiClient.get<User>('/me');
  return data;
}

// ---------- Matching ----------

export interface MatchingDetailResponseFull extends MatchingResponse {
  job_text: string;
  engineer_text: string;
  supplement?: SupplementInfo;
}

export async function executeMatching(req: MatchingRequest): Promise<MatchingResponse> {
  const { data } = await apiClient.post<MatchingResponse>('/matchings', req);
  return data;
}

export async function getMatchings(
  page: number = 1,
  pageSize: number = 20,
  grade?: string,
): Promise<MatchingListResponse> {
  const params: Record<string, any> = { page, page_size: pageSize };
  if (grade) {
    params.grade = grade;
  }
  const { data } = await apiClient.get<MatchingListResponse>('/matchings', { params });
  return data;
}

export async function getMatchingDetail(id: string): Promise<MatchingDetailResponseFull> {
  const { data } = await apiClient.get<MatchingDetailResponseFull>(`/matchings/${id}`);
  return data;
}

export async function deleteMatching(id: string): Promise<void> {
  await apiClient.delete(`/matchings/${id}`);
}

// ---------- File ----------

export async function parseFile(file: File): Promise<FileParseResponse> {
  const formData = new FormData();
  formData.append('file', file);
  const { data } = await apiClient.post<FileParseResponse>('/files/parse', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
  return data;
}

// ---------- Settings ----------

export async function getSettings(): Promise<SettingsResponse> {
  const { data } = await apiClient.get<SettingsResponse>('/settings');
  return data;
}

export async function updateSetting(key: string, value: any): Promise<void> {
  await apiClient.put(`/settings/${key}`, { value });
}

// ---------- Job Groups ----------

export async function createJobGroup(name: string, matchingId: string): Promise<JobGroup> {
  const { data } = await apiClient.post<JobGroup>('/job-groups', { name, matching_id: matchingId });
  return data;
}

export async function getJobGroups(): Promise<JobGroup[]> {
  const { data } = await apiClient.get<JobGroup[]>('/job-groups');
  return data;
}

export async function getJobGroup(id: string): Promise<JobGroup> {
  const { data } = await apiClient.get<JobGroup>(`/job-groups/${id}`);
  return data;
}

export async function deleteJobGroup(id: string): Promise<void> {
  await apiClient.delete(`/job-groups/${id}`);
}

export async function linkMatchingToJobGroup(matchingId: string, jobGroupId: string): Promise<void> {
  await apiClient.put(`/matchings/${matchingId}/job-group`, { job_group_id: jobGroupId });
}

export async function unlinkMatchingFromJobGroup(matchingId: string): Promise<void> {
  await apiClient.delete(`/matchings/${matchingId}/job-group`);
}

export default apiClient;
