// Supply Chain
export type SupplyChainLevel = 0 | 1 | 2 | 3 | 4;

export const SUPPLY_CHAIN_LABELS: Record<number, string> = {
  0: '不明',
  1: 'エンド直',
  2: '1次請け',
  3: '2次請け',
  4: '3次以上',
};

// User
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'sales';
}

// Login
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  user: User;
}

// Matching
export interface SupplementInfo {
  affiliation_type?: 'duesk' | 'partner';
  affiliation_name?: string;
  rate?: number;
  nationality?: string;
  employment_type?: string;
  available_from?: string;
  supply_chain_level?: number;
  supply_chain_source?: string;
}

export interface MatchingRequest {
  job_text: string;
  engineer_text?: string;
  engineer_file_key?: string;
  supplement?: SupplementInfo;
}

export interface MatchingResponse {
  id: string;
  total_score: number;
  grade: string;
  grade_label: string;
  result: MatchResult;
  model_used: string;
  tokens_used: number;
  created_at: string;
  job_group_id?: string;
  supply_chain_level: number;
  supply_chain_source?: string;
}

export interface MatchResult {
  job_summary: {
    name: string;
    location: string;
    remote: string;
    rate: string;
    start: string;
    conditions: string;
  };
  engineer_summary: {
    initials: string;
    age: number;
    gender: string;
    nearest_station: string;
    affiliation: string;
    rate: string;
    available_from: string;
  };
  total_score: number;
  grade: string;
  grade_label: string;
  scores: {
    skill: {
      score: number;
      max: number;
      reason: string;
      required_skills: Array<{
        skill: string;
        status: 'met' | 'partial' | 'unmet';
        detail: string;
      }>;
      optional_skills: Array<{
        skill: string;
        status: string;
        detail: string;
      }>;
    };
    timing: {
      score: number;
      max: number;
      reason: string;
    };
    rate: {
      score: number;
      max: number;
      reason: string;
      calculation: string;
    };
    experience_years: {
      score: number;
      max: number;
      reason: string;
    };
    location: {
      score: number;
      max: number;
      reason: string;
      commute_time?: string;
    };
    industry: {
      score: number;
      max: number;
      reason: string;
    };
  };
  ng_flags: Record<
    string,
    {
      status: 'ok' | 'ng' | 'warning' | 'unknown';
      detail: string;
    }
  >;
  negatives: string[];
  positives: string[];
  warnings: string[];
  advice: string;
  confirmation_hints: Array<{
    target: string;
    question: string;
    reason: string;
  }>;
}

export interface MatchingListResponse {
  items: MatchingListItem[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface MatchingListItem {
  id: string;
  total_score: number;
  grade: string;
  grade_label: string;
  job_summary?: {
    name: string;
    location?: string;
    rate?: string;
  };
  model_used: string;
  created_at: string;
  job_group_id?: string;
  supply_chain_level: number;
  supply_chain_source?: string;
}

// Job Group
export interface JobGroup {
  id: string;
  name: string;
  matchings: MatchingListItem[];
  best_route?: MatchingListItem;
  created_at: string;
}

// File
export interface FileParseResponse {
  text: string;
  file_key: string;
  file_name: string;
  parse_warnings: string[];
}

// Settings
export interface SettingItem {
  key: string;
  value: any;
}

export interface SettingsResponse {
  settings: SettingItem[];
}
