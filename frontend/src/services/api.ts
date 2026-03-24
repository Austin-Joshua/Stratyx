import axios, { AxiosError, type AxiosRequestConfig } from "axios";

export type User = {
  id: string;
  name: string;
  email: string;
  role: "SUPER_ADMIN" | "ADMIN" | "MODERATOR" | "USER" | "GUEST";
  isActive: boolean;
  isEmailVerified: boolean;
  avatarUrl?: string;
  twoFAEnabled?: boolean;
  createdAt: string;
};

export type AuthResponse = {
  accessToken: string;
  refreshToken: string;
  user: User;
  requires2FA?: boolean;
  challengeToken?: string;
};

export type Item = {
  id: string;
  title: string;
  description: string;
  ownerId: string;
  createdAt: string;
  updatedAt: string;
};

export type Session = {
  id: string;
  userId: string;
  deviceName: string;
  ipAddress: string;
  createdAt: string;
  lastSeenAt: string;
  expiresAt: string;
};

export type Comment = {
  id: string;
  itemId: string;
  authorId: string;
  parentId?: string;
  content: string;
  mentions: string[];
  createdAt: string;
};

export type Notification = {
  id: string;
  userId: string;
  type: string;
  title: string;
  body: string;
  read: boolean;
  createdAt: string;
};

export type Activity = {
  id: string;
  actorId: string;
  entity: string;
  entityId: string;
  action: string;
  message: string;
  createdAt: string;
};

export type DashboardSummary = {
  itemsCount: number;
  commentsCount: number;
  unreadNotifications: number;
};

export type AdminMetrics = {
  usersTotal: number;
  usersActive: number;
  itemsTotal: number;
  sessionsActive: number;
  commentsTotal: number;
  notificationsTotal: number;
};

export type OAuthAccount = {
  id: string;
  provider: string;
  email: string;
  providerUserId: string;
  createdAt: string;
};

export type FileAsset = {
  id: string;
  ownerId: string;
  itemId: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  url: string;
  storageKind: string;
  createdAt: string;
};

export type ModerationReport = {
  id: string;
  commentId: string;
  reporterId: string;
  reason: string;
  status: string;
  reviewedBy?: string;
  reviewedAt?: string;
  createdAt: string;
};

export type ItemListResponse = {
  items: Item[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type ItemListQuery = {
  page?: number;
  pageSize?: number;
  search?: string;
  sortBy?: "updatedAt" | "createdAt" | "title";
  sortOrder?: "asc" | "desc";
};

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

const ACCESS_TOKEN_KEY = "stratyx_access_token";
const REFRESH_TOKEN_KEY = "stratyx_refresh_token";

function getLocalStorage() {
  if (typeof window === "undefined") {
    return null;
  }
  return window.localStorage;
}

export function getAccessToken() {
  return getLocalStorage()?.getItem(ACCESS_TOKEN_KEY) ?? "";
}

export function getRefreshToken() {
  return getLocalStorage()?.getItem(REFRESH_TOKEN_KEY) ?? "";
}

export function setTokens(accessToken: string, refreshToken: string) {
  const store = getLocalStorage();
  if (!store) {
    return;
  }
  store.setItem(ACCESS_TOKEN_KEY, accessToken);
  store.setItem(REFRESH_TOKEN_KEY, refreshToken);
}

export function clearTokens() {
  const store = getLocalStorage();
  if (!store) {
    return;
  }
  store.removeItem(ACCESS_TOKEN_KEY);
  store.removeItem(REFRESH_TOKEN_KEY);
}

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 12000,
});

let isRefreshing = false;
let requestQueue: Array<(token: string) => void> = [];

function drainQueue(token: string) {
  requestQueue.forEach((resolve) => resolve(token));
  requestQueue = [];
}

api.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as AxiosRequestConfig & {
      _retry?: boolean;
    };
    if (
      error.response?.status !== 401 ||
      originalRequest._retry ||
      originalRequest.url?.includes("/api/auth/login") ||
      originalRequest.url?.includes("/api/auth/register") ||
      originalRequest.url?.includes("/api/auth/refresh")
    ) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    if (!isRefreshing) {
      isRefreshing = true;
      try {
        const refreshToken = getRefreshToken();
        if (!refreshToken) {
          throw new Error("No refresh token");
        }
        const response = await axios.post<{ accessToken: string; refreshToken: string }>(
          `${API_BASE_URL}/api/auth/refresh`,
          { refreshToken },
          { timeout: 12000 },
        );
        setTokens(response.data.accessToken, response.data.refreshToken);
        drainQueue(response.data.accessToken);
      } catch (refreshError) {
        clearTokens();
        if (typeof window !== "undefined") {
          window.location.href = "/login";
        }
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return new Promise((resolve) => {
      requestQueue.push((newToken: string) => {
        if (!originalRequest.headers) {
          originalRequest.headers = {};
        }
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        resolve(api(originalRequest));
      });
    });
  },
);

async function requestWithRetry<T>(
  operation: () => Promise<{ data: T }>,
  retries = 1,
): Promise<T> {
  try {
    const response = await operation();
    return response.data;
  } catch (error) {
    const networkError = error as AxiosError;
    if (retries > 0 && (!networkError.response || networkError.code === "ECONNABORTED")) {
      return requestWithRetry(operation, retries - 1);
    }
    throw error;
  }
}

export const authApi = {
  register: (payload: { name: string; email: string; password: string }) =>
    requestWithRetry(() => api.post<{ message: string; user: User }>("/api/auth/register", payload)),
  login: (payload: { email: string; password: string; remember?: boolean }) =>
    requestWithRetry(() => api.post<AuthResponse>("/api/auth/login", payload)),
  me: () => requestWithRetry(() => api.get<User>("/api/auth/me")),
  refresh: () =>
    requestWithRetry(() =>
      api.post<{ accessToken: string; refreshToken: string }>("/api/auth/refresh", {
        refreshToken: getRefreshToken(),
      }),
    ),
  logout: () =>
    requestWithRetry(() =>
      api.post<{ message: string }>("/api/auth/logout", {
        refreshToken: getRefreshToken(),
      }),
    ),
  forgotPassword: (email: string) =>
    requestWithRetry(() =>
      api.post<{ message: string; resetToken?: string }>("/api/auth/forgot-password", {
        email,
      }),
    ),
  resetPassword: (payload: { token: string; newPassword: string }) =>
    requestWithRetry(() => api.post<{ message: string }>("/api/auth/reset-password", payload)),
  verifyEmail: (token: string) =>
    requestWithRetry(() => api.post<{ message: string }>("/api/auth/verify-email", { token })),
  sessions: () => requestWithRetry(() => api.get<Session[]>("/api/auth/sessions")),
  logoutAll: () => requestWithRetry(() => api.post<{ message: string }>("/api/auth/logout-all")),
  setup2FA: () =>
    requestWithRetry(() => api.post<{ secret: string; otpAuthURL: string }>("/api/auth/2fa/setup")),
  verify2FASetup: (code: string) =>
    requestWithRetry(() => api.post<{ message: string }>("/api/auth/2fa/verify-setup", { code })),
  disable2FA: (code: string) =>
    requestWithRetry(() => api.post<{ message: string }>("/api/auth/2fa/disable", { code })),
  complete2FALogin: (challengeToken: string, code: string) =>
    requestWithRetry(() => api.post<AuthResponse>("/api/auth/2fa/complete-login", { challengeToken, code })),
  connectedAccounts: () => requestWithRetry(() => api.get<OAuthAccount[]>("/api/auth/connected-accounts")),
  disconnectAccount: (provider: string) =>
    requestWithRetry(() => api.delete<{ message: string }>(`/api/auth/connected-accounts/${provider}`)),
  updateProfile: (payload: {
    name: string;
    avatarUrl?: string;
    privacySettings?: { showEmail: boolean; showProfile: boolean };
    notificationPrefs?: { emailMentions: boolean; inAppMentions: boolean; productNews: boolean };
  }) => requestWithRetry(() => api.put<User>("/api/auth/profile", payload)),
  uploadAvatar: (file: File) =>
    requestWithRetry(async () => {
      const form = new FormData();
      form.append("file", file);
      return api.post<{ avatarUrl: string }>("/api/auth/avatar", form, {
        headers: { "Content-Type": "multipart/form-data" },
      });
    }),
};

export const itemsApi = {
  create: (payload: { title: string; description: string }) =>
    requestWithRetry(() => api.post<Item>("/api/items", payload)),
  list: (query: ItemListQuery = {}) =>
    requestWithRetry(() =>
      api.get<ItemListResponse>("/api/items", {
        params: {
          page: query.page ?? 1,
          pageSize: query.pageSize ?? 10,
          search: query.search ?? "",
          sortBy: query.sortBy ?? "updatedAt",
          sortOrder: query.sortOrder ?? "desc",
        },
      }),
    ),
  getByID: (id: string) => requestWithRetry(() => api.get<Item>(`/api/items/${id}`)),
  update: (id: string, payload: { title: string; description: string }) =>
    requestWithRetry(() => api.put<Item>(`/api/items/${id}`, payload)),
  remove: (id: string) => requestWithRetry(() => api.delete<{ message: string }>(`/api/items/${id}`)),
  comments: (id: string) => requestWithRetry(() => api.get<Comment[]>(`/api/items/${id}/comments`)),
  addComment: (id: string, payload: { content: string; parentId?: string }) =>
    requestWithRetry(() => api.post<Comment>(`/api/items/${id}/comments`, payload)),
  uploadAttachment: (id: string, file: File) =>
    requestWithRetry(async () => {
      const form = new FormData();
      form.append("file", file);
      return api.post<FileAsset>(`/api/items/${id}/attachments`, form, {
        headers: { "Content-Type": "multipart/form-data" },
      });
    }),
  reportComment: (commentID: string, reason: string) =>
    requestWithRetry(() => api.post<ModerationReport>(`/api/comments/${commentID}/report`, { reason })),
};

export const dashboardApi = {
  summary: () => requestWithRetry(() => api.get<DashboardSummary>("/api/dashboard/summary")),
  activity: () => requestWithRetry(() => api.get<Activity[]>("/api/activity")),
};

export const notificationsApi = {
  list: () => requestWithRetry(() => api.get<Notification[]>("/api/notifications")),
  markRead: (id: string, read: boolean) =>
    requestWithRetry(() => api.put<{ message: string }>(`/api/notifications/${id}/read`, { read })),
};

export const adminApi = {
  metrics: () => requestWithRetry(() => api.get<AdminMetrics>("/api/admin/metrics")),
  users: () => requestWithRetry(() => api.get<User[]>("/api/admin/users")),
  setRole: (id: string, role: User["role"]) =>
    requestWithRetry(() => api.put<{ message: string }>(`/api/admin/users/${id}/role`, { role })),
  setActive: (id: string, active: boolean) =>
    requestWithRetry(() => api.put<{ message: string }>(`/api/admin/users/${id}/active`, { active })),
  moderationQueue: (status = "") =>
    requestWithRetry(() => api.get<ModerationReport[]>("/api/admin/moderation/reports", { params: { status } })),
  reviewModeration: (id: string, status: "APPROVED" | "REJECTED") =>
    requestWithRetry(() => api.put<{ message: string }>(`/api/admin/moderation/reports/${id}/review`, { status })),
  auditLogs: () => requestWithRetry(() => api.get<Activity[]>("/api/admin/audit-logs")),
};

export const filesApi = {
  createAccessUrl: (id: string) =>
    requestWithRetry(() => api.post<{ url: string }>(`/api/files/${id}/access`)),
};

export { API_BASE_URL };
