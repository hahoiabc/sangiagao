import { clearAuth } from "@/lib/auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

interface RequestOptions extends RequestInit {
  token?: string;
}

let isRefreshing = false;
let refreshPromise: Promise<string> | null = null;

async function tryRefreshToken(): Promise<string> {
  const savedRefresh = localStorage.getItem("admin_refresh_token");
  if (!savedRefresh) {
    throw new Error("No refresh token");
  }

  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: savedRefresh }),
  });

  if (!res.ok) {
    throw new Error("Refresh failed");
  }

  const data = await res.json();
  localStorage.setItem("admin_token", data.access_token);
  localStorage.setItem("admin_refresh_token", data.refresh_token);
  return data.access_token;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { token, ...init } = options;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  // Auto-refresh on 401 for authenticated requests
  if (res.status === 401 && token) {
    try {
      if (!isRefreshing) {
        isRefreshing = true;
        refreshPromise = tryRefreshToken();
      }
      const newToken = await refreshPromise!;
      isRefreshing = false;
      refreshPromise = null;

      // Retry with new token
      headers["Authorization"] = `Bearer ${newToken}`;
      const retryRes = await fetch(`${API_BASE}${path}`, { ...init, headers });

      if (!retryRes.ok) {
        const body = await retryRes.json().catch(() => ({}));
        throw new ApiError(retryRes.status, body.error || "unknown", body.message || retryRes.statusText);
      }
      return retryRes.json();
    } catch {
      isRefreshing = false;
      refreshPromise = null;
      clearAuth();
      throw new ApiError(401, "session_expired", "Phiên đăng nhập đã hết hạn.");
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.error || "unknown", body.message || res.statusText);
  }

  return res.json();
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string
  ) {
    super(message);
  }
}

// --- Auth ---
export async function sendOTP(phone: string) {
  return request<{ message: string; expires_in: number }>("/auth/send-otp", {
    method: "POST",
    body: JSON.stringify({ phone }),
  });
}

export async function verifyOTP(phone: string, code: string) {
  return request<{
    user: { id: string; phone: string; name?: string; avatar_url?: string; role: string };
    tokens: { access_token: string; refresh_token: string; expires_in: number };
  }>("/auth/verify-otp", {
    method: "POST",
    body: JSON.stringify({ phone, code }),
  });
}

export async function loginPassword(phone: string, password: string) {
  return request<{
    user: { id: string; phone: string; name?: string; avatar_url?: string; role: string };
    tokens: { access_token: string; refresh_token: string; expires_in: number };
  }>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ phone, password }),
  });
}

export async function refreshToken(refreshToken: string) {
  return request<{ access_token: string; refresh_token: string; expires_in: number }>("/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refresh_token: refreshToken }),
  });
}

// --- Dashboard ---
export async function getDashboardStats(token: string) {
  return request<Record<string, number>>("/admin/dashboard/stats", { token });
}

export interface DashboardCharts {
  users_by_month: { month: string; count: number }[];
  listings_by_month: { month: string; count: number }[];
  subs_by_month: { month: string; count: number }[];
  users_by_role: { label: string; count: number }[];
  listings_by_rice_type: { label: string; count: number }[];
  listings_by_province: { label: string; count: number }[];
}

export async function getDashboardCharts(token: string) {
  return request<DashboardCharts>("/admin/dashboard/charts", { token });
}

// --- Users ---
export interface User {
  id: string;
  phone: string;
  role: string;
  name: string;
  avatar_url: string;
  address: string;
  province: string;
  ward: string;
  description: string;
  org_name: string;
  is_blocked: boolean;
  block_reason: string;
  subscription_expires_at: string | null;
  created_at: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export async function listUsers(token: string, search: string, page: number, limit: number) {
  const params = new URLSearchParams({ search, page: String(page), limit: String(limit) });
  return request<PaginatedResponse<User>>(`/admin/users?${params}`, { token });
}

export async function blockUser(token: string, userId: string, reason: string) {
  return request<User>(`/admin/users/${userId}/block`, {
    token,
    method: "PUT",
    body: JSON.stringify({ reason }),
  });
}

export async function unblockUser(token: string, userId: string) {
  return request<User>(`/admin/users/${userId}/unblock`, { token, method: "PUT" });
}

export async function getUserDetail(token: string, userId: string) {
  return request<User>(`/admin/users/${userId}`, { token });
}

export async function listUserListings(token: string, userId: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Listing>>(`/admin/users/${userId}/listings?${params}`, { token });
}

export interface Subscription {
  id: string;
  user_id: string;
  plan: string;
  duration_months: number;
  amount: number;
  started_at: string;
  expires_at: string;
  status: string;
  created_at: string;
}

export async function listUserSubscriptions(token: string, userId: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Subscription>>(`/admin/users/${userId}/subscriptions?${params}`, { token });
}

// --- Listings ---
export interface Listing {
  id: string;
  user_id: string;
  title: string;
  category?: string;
  rice_type: string;
  province: string;
  quantity_kg: number;
  price_per_kg: number;
  description?: string;
  certifications?: string;
  images: string[];
  status: string;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface ListingDetail extends Listing {
  seller?: {
    id: string;
    phone: string;
    name: string;
    avatar_url: string;
    province: string;
    org_name: string;
  };
}

export async function browseListings(token: string, page: number, limit: number, category?: string) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (category) params.set("category", category);
  return request<PaginatedResponse<Listing>>(`/marketplace/search?${params}`, { token });
}

export async function getListingDetail(token: string, id: string) {
  return request<ListingDetail>(`/marketplace/${id}`, { token });
}

export async function deleteListing(token: string, listingId: string) {
  return request<void>(`/admin/listings/${listingId}`, { token, method: "DELETE" });
}

// --- Reports ---
export interface Report {
  id: string;
  reporter_id: string;
  target_type: string;
  target_id: string;
  reason: string;
  description: string;
  status: string;
  admin_action: string;
  admin_note: string;
  resolved_by: string;
  resolved_at: string;
  created_at: string;
}

export async function listReports(token: string, page: number, limit: number, status: string = "pending") {
  const params = new URLSearchParams({ page: String(page), limit: String(limit), status });
  return request<PaginatedResponse<Report>>(`/admin/reports?${params}`, { token });
}

export async function resolveReport(token: string, reportId: string, action: string, adminNote?: string) {
  return request<Report>(`/admin/reports/${reportId}`, {
    token,
    method: "PUT",
    body: JSON.stringify({ admin_action: action, admin_note: adminNote || undefined }),
  });
}

export async function changeUserRole(token: string, userId: string, role: string) {
  return request<User>(`/admin/users/${userId}/role`, {
    token,
    method: "PUT",
    body: JSON.stringify({ role }),
  });
}

// --- Subscriptions ---
export interface SubscriptionPlan {
  id: string;
  months: number;
  amount: number;
  label: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export async function activateSubscription(token: string, userId: string, months: number) {
  return request<{ message: string; subscription: Subscription }>(`/admin/subscriptions/${userId}/activate`, {
    token,
    method: "POST",
    body: JSON.stringify({ months }),
  });
}

export async function getSubscriptionPlans(token: string) {
  return request<{ plans: SubscriptionPlan[] }>("/subscription/plans", { token });
}

// --- Plan Management (owner only) ---
export async function listAllPlans(token: string) {
  return request<{ plans: SubscriptionPlan[] }>("/admin/plans", { token });
}

export async function createPlan(token: string, data: { months: number; amount: number; label: string }) {
  return request<SubscriptionPlan>("/admin/plans", { token, method: "POST", body: JSON.stringify(data) });
}

export async function updatePlan(token: string, id: string, data: { months?: number; amount?: number; label?: string; is_active?: boolean }) {
  return request<SubscriptionPlan>(`/admin/plans/${id}`, { token, method: "PUT", body: JSON.stringify(data) });
}

export async function deletePlan(token: string, id: string) {
  return request<void>(`/admin/plans/${id}`, { token, method: "DELETE" });
}

// --- Subscription Revenue Stats ---
export interface SubRevenueMonth {
  month: string;
  paid_count: number;
  trial_count: number;
  revenue: number;
}

export interface SubRevenueStats {
  total_subscriptions: number;
  active_count: number;
  expired_count: number;
  paid_count: number;
  trial_count: number;
  total_revenue: number;
  monthly_revenue: SubRevenueMonth[];
}

export async function getSubscriptionRevenueStats(token: string) {
  return request<SubRevenueStats>("/admin/subscriptions/stats", { token });
}

// --- Daily Revenue ---
export interface SubRevenueDay {
  date: string;
  paid_count: number;
  trial_count: number;
  revenue: number;
}

export interface SubDailyRevenueReport {
  from: string;
  to: string;
  total_paid: number;
  total_trial: number;
  total_revenue: number;
  days: SubRevenueDay[];
}

export async function getDailyRevenue(token: string, from: string, to: string) {
  const params = new URLSearchParams({ from, to });
  return request<SubDailyRevenueReport>(`/admin/subscriptions/daily-revenue?${params}`, { token });
}

// --- Sponsors ---
export interface ProductSponsor {
  id: string;
  product_key: string;
  logo_url: string;
  sponsor_name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RiceProduct {
  key: string;
  label: string;
  category: string;
}

export interface RiceCategory {
  key: string;
  label: string;
  products: RiceProduct[];
}

export async function listSponsors(token: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<ProductSponsor>>(`/admin/sponsors?${params}`, { token });
}

export async function createSponsor(token: string, data: { product_key: string; logo_url: string; sponsor_name: string }) {
  return request<ProductSponsor>("/admin/sponsors", { token, method: "POST", body: JSON.stringify(data) });
}

export async function updateSponsor(token: string, id: string, data: { logo_url?: string; sponsor_name?: string; is_active?: boolean }) {
  return request<ProductSponsor>(`/admin/sponsors/${id}`, { token, method: "PUT", body: JSON.stringify(data) });
}

export async function deleteSponsor(token: string, id: string) {
  return request<void>(`/admin/sponsors/${id}`, { token, method: "DELETE" });
}

export async function getProductCatalog(token: string) {
  return request<RiceCategory[]>("/marketplace/product-catalog", { token });
}

// --- Catalog Management ---
export interface CatalogCategory {
  id: string;
  key: string;
  label: string;
  icon: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface CatalogProduct {
  id: string;
  key: string;
  label: string;
  category_id: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export async function listCatalogCategories(token: string) {
  return request<CatalogCategory[]>("/admin/catalog/categories", { token });
}

export async function createCatalogCategory(token: string, data: { key: string; label: string; icon?: string }) {
  return request<CatalogCategory>("/admin/catalog/categories", { token, method: "POST", body: JSON.stringify(data) });
}

export async function updateCatalogCategory(token: string, id: string, data: { label?: string; icon?: string; sort_order?: number; is_active?: boolean }) {
  return request<CatalogCategory>(`/admin/catalog/categories/${id}`, { token, method: "PUT", body: JSON.stringify(data) });
}

export async function deleteCatalogCategory(token: string, id: string) {
  return request<void>(`/admin/catalog/categories/${id}`, { token, method: "DELETE" });
}

export async function listCatalogProducts(token: string) {
  return request<CatalogProduct[]>("/admin/catalog/products", { token });
}

export async function createCatalogProduct(token: string, data: { key: string; label: string; category_id: string }) {
  return request<CatalogProduct>("/admin/catalog/products", { token, method: "POST", body: JSON.stringify(data) });
}

export async function updateCatalogProduct(token: string, id: string, data: { label?: string; category_id?: string; sort_order?: number; is_active?: boolean }) {
  return request<CatalogProduct>(`/admin/catalog/products/${id}`, { token, method: "PUT", body: JSON.stringify(data) });
}

export async function deleteCatalogProduct(token: string, id: string) {
  return request<void>(`/admin/catalog/products/${id}`, { token, method: "DELETE" });
}

// --- Feedbacks ---
export interface Feedback {
  id: string;
  user_id: string;
  user_name: string;
  user_phone: string;
  content: string;
  reply: string | null;
  replied_at: string | null;
  created_at: string;
}

export async function listFeedbacks(token: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Feedback>>(`/admin/feedbacks?${params}`, { token });
}

export async function replyFeedback(token: string, id: string, reply: string) {
  return request<Feedback>(`/admin/feedbacks/${id}/reply`, {
    token,
    method: "PUT",
    body: JSON.stringify({ reply }),
  });
}

// --- Profile (current admin) ---
export async function getMe(token: string) {
  return request<User>("/users/me", { token });
}

export interface UpdateProfileData {
  name?: string;
  address?: string;
  province?: string;
  ward?: string;
  description?: string;
  org_name?: string;
}

export async function updateMe(token: string, data: UpdateProfileData) {
  return request<User>("/users/me", { token, method: "PUT", body: JSON.stringify(data) });
}

export async function updateMyAvatar(token: string, url: string) {
  return request<User>("/users/me/avatar", { token, method: "POST", body: JSON.stringify({ url }) });
}

// --- Upload ---
export async function uploadImage(token: string, file: File, folder: "avatars" | "listings") {
  const formData = new FormData();
  formData.append("image", file);
  formData.append("folder", folder);

  const doUpload = (t: string) =>
    fetch(`${API_BASE}/upload/image`, {
      method: "POST",
      headers: { Authorization: `Bearer ${t}` },
      body: formData,
    });

  let res = await doUpload(token);

  // Auto-refresh on 401
  if (res.status === 401) {
    try {
      if (!isRefreshing) {
        isRefreshing = true;
        refreshPromise = tryRefreshToken();
      }
      const newToken = await refreshPromise!;
      isRefreshing = false;
      refreshPromise = null;
      res = await doUpload(newToken);
    } catch {
      isRefreshing = false;
      refreshPromise = null;
      clearAuth();
      throw new ApiError(401, "session_expired", "Phiên đăng nhập đã hết hạn.");
    }
  }

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new ApiError(res.status, body.error || "unknown", body.message || res.statusText);
  }

  return res.json() as Promise<{ url: string }>;
}

// --- System Monitoring ---

export interface SystemStats {
  uptime: string;
  go_version: string;
  hostname: string;
  cpu_cores: number;
  cpu_percent: number;
  mem_total_mb: number;
  mem_used_mb: number;
  mem_percent: number;
  disk_total_gb: number;
  disk_used_gb: number;
  disk_percent: number;
  goroutines: number;
  heap_alloc_mb: number;
  heap_sys_mb: number;
  gc_cycles: number;
  online_users: number;
  online_ids: string[];
}

export async function getSystemStats(token: string): Promise<SystemStats> {
  return request<SystemStats>("/admin/system/stats", { token });
}
