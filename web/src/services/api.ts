import { clearAuth } from "@/lib/auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

interface RequestOptions extends RequestInit {
  token?: string;
}

let isRefreshing = false;
let refreshPromise: Promise<string> | null = null;

async function tryRefreshToken(): Promise<string> {
  const savedRefresh = localStorage.getItem("web_refresh_token");
  if (!savedRefresh) throw new Error("No refresh token");

  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: savedRefresh }),
  });

  if (!res.ok) throw new Error("Refresh failed");

  const data = await res.json();
  localStorage.setItem("web_token", data.access_token);
  localStorage.setItem("web_refresh_token", data.refresh_token);
  return data.access_token;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const { token, ...init } = options;
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init.headers as Record<string, string>),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers });

  if (res.status === 401 && token) {
    try {
      if (!isRefreshing) {
        isRefreshing = true;
        refreshPromise = tryRefreshToken();
      }
      const newToken = await refreshPromise!;
      isRefreshing = false;
      refreshPromise = null;
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
  constructor(public status: number, public code: string, message: string) {
    super(message);
  }
}

// --- Types ---
export interface User {
  id: string;
  phone: string;
  role: string;
  name: string;
  avatar_url: string;
  address: string;
  province: string;
  district: string;
  ward: string;
  description: string;
  org_name: string;
  is_blocked: boolean;
  subscription_expires_at: string | null;
  created_at: string;
}

export interface PublicProfile {
  id: string;
  role: string;
  name: string;
  avatar_url: string;
  province: string;
  description: string;
  org_name: string;
  created_at: string;
}

export interface Listing {
  id: string;
  user_id: string;
  title: string;
  category?: string;
  rice_type: string;
  province: string;
  district?: string;
  ward?: string;
  quantity_kg: number;
  price_per_kg: number;
  harvest_season?: string;
  description?: string;
  certifications?: string;
  images: string[];
  status: string;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface ListingDetail extends Listing {
  seller?: PublicProfile;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
}

export interface Conversation {
  id: string;
  member_id: string;
  seller_id: string;
  listing_id?: string;
  last_message_at: string;
  created_at: string;
  other_user?: PublicProfile;
  unread_count: number;
}

export interface Message {
  id: string;
  conversation_id: string;
  sender_id: string;
  content: string;
  type: string;
  read_at?: string;
  created_at: string;
}

export interface Rating {
  id: string;
  reviewer_id: string;
  seller_id: string;
  stars: number;
  comment: string;
  reviewer_name?: string;
  reviewer_avatar?: string;
  created_at: string;
}

export interface RatingSummary {
  average: number;
  total: number;
  distribution: Record<string, number>;
}

export interface AppNotification {
  id: string;
  user_id: string;
  type: string;
  title: string;
  body: string;
  data?: Record<string, string>;
  is_read: boolean;
  created_at: string;
}

export interface PriceBoardEntry {
  product_key: string;
  product_label: string;
  min_price: number | null;
  listing_count: number;
  sponsor_logo?: string;
}

export interface PriceBoardCategory {
  category_key: string;
  category_label: string;
  products: PriceBoardEntry[];
}

export interface PriceBoardResponse {
  categories: PriceBoardCategory[];
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

export interface SubscriptionPlan {
  id: string;
  months: number;
  amount: number;
  label: string;
  is_active: boolean;
}

export interface SubscriptionStatus {
  has_active: boolean;
  expires_at: string | null;
  plan: string;
  days_remaining: number;
}

export interface Feedback {
  id: string;
  user_id: string;
  content: string;
  reply: string | null;
  replied_at: string | null;
  created_at: string;
}

// --- Auth ---
export async function loginPassword(phone: string, password: string) {
  return request<{
    user: { id: string; phone: string; name?: string; avatar_url?: string; role: string };
    tokens: { access_token: string; refresh_token: string; expires_in: number };
  }>("/auth/login", { method: "POST", body: JSON.stringify({ phone, password }) });
}

export async function register(phone: string) {
  return request<{ message: string }>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ phone }),
  });
}

export async function completeRegister(data: {
  phone: string;
  code: string;
  name: string;
  password: string;
  province?: string;
  ward?: string;
  address?: string;
}) {
  return request<{
    user: { id: string; phone: string; name?: string; avatar_url?: string; role: string };
    tokens: { access_token: string; refresh_token: string; expires_in: number };
  }>("/auth/complete-register", { method: "POST", body: JSON.stringify(data) });
}

export async function sendOTP(phone: string) {
  return request<{ message: string }>("/auth/send-otp", {
    method: "POST",
    body: JSON.stringify({ phone }),
  });
}

export async function resetPassword(phone: string, code: string, new_password: string) {
  return request<{ message: string }>("/auth/reset-password", {
    method: "POST",
    body: JSON.stringify({ phone, code, new_password }),
  });
}

// --- Account ---
export async function changePassword(token: string, currentPassword: string, newPassword: string) {
  return request<{ message: string }>("/users/me/password", {
    token,
    method: "POST",
    body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
  });
}

export async function changePhone(token: string, newPhone: string, code: string) {
  return request<User>("/users/me/phone", {
    token,
    method: "POST",
    body: JSON.stringify({ new_phone: newPhone, code }),
  });
}

// --- Public ---
export async function getPriceBoard() {
  return request<PriceBoardResponse>("/marketplace/price-board");
}

export async function getProductCatalog() {
  return request<RiceCategory[]>("/marketplace/product-catalog");
}

export async function browseMarketplace(page: number, limit: number, category?: string) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  if (category) params.set("category", category);
  return request<PaginatedResponse<Listing>>(`/marketplace/search?${params}`);
}

export async function searchMarketplace(params: {
  q?: string;
  category?: string;
  rice_type?: string;
  province?: string;
  ward?: string;
  min_price?: number;
  max_price?: number;
  sort?: string;
  page?: number;
  limit?: number;
}) {
  const sp = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== "") sp.set(k, String(v));
  });
  if (!sp.has("page")) sp.set("page", "1");
  if (!sp.has("limit")) sp.set("limit", "20");
  return request<PaginatedResponse<Listing>>(`/marketplace/search?${sp}`);
}

export async function getListingDetail(id: string) {
  return request<ListingDetail>(`/marketplace/${id}`);
}

export async function getPublicProfile(userId: string) {
  return request<PublicProfile>(`/users/${userId}/profile`);
}

export async function getSellerRatings(userId: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Rating>>(`/users/${userId}/ratings?${params}`);
}

export async function getRatingSummary(userId: string) {
  return request<RatingSummary>(`/users/${userId}/rating-summary`);
}

// --- Protected ---
export async function getMe(token: string) {
  return request<User>("/users/me", { token });
}

export async function updateMe(token: string, data: Record<string, string>) {
  return request<User>("/users/me", { token, method: "PUT", body: JSON.stringify(data) });
}

export async function updateMyAvatar(token: string, url: string) {
  return request<User>("/users/me/avatar", { token, method: "POST", body: JSON.stringify({ url }) });
}

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

export async function uploadAudio(token: string, blob: Blob) {
  const formData = new FormData();
  formData.append("audio", blob, "recording.webm");

  const doUpload = (t: string) =>
    fetch(`${API_BASE}/upload/audio`, {
      method: "POST",
      headers: { Authorization: `Bearer ${t}` },
      body: formData,
    });

  let res = await doUpload(token);

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

// --- Listings ---
export async function getMyListings(token: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Listing>>(`/listings/my?${params}`, { token });
}

export async function createListing(token: string, data: Record<string, unknown>) {
  return request<Listing>("/listings", { token, method: "POST", body: JSON.stringify(data) });
}

export async function batchCreateListings(token: string, items: Record<string, unknown>[]) {
  return request<{ listings: Listing[] }>("/listings/batch", {
    token,
    method: "POST",
    body: JSON.stringify({ items }),
  });
}

export async function updateListing(token: string, id: string, data: Record<string, unknown>) {
  return request<Listing>(`/listings/${id}`, { token, method: "PUT", body: JSON.stringify(data) });
}

export async function deleteListing(token: string, id: string) {
  return request<void>(`/listings/${id}`, { token, method: "DELETE" });
}

export async function addListingImage(token: string, listingId: string, url: string) {
  return request<Listing>(`/listings/${listingId}/images`, {
    token,
    method: "POST",
    body: JSON.stringify({ url }),
  });
}

// --- Chat ---
export async function getConversations(token: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Conversation>>(`/conversations?${params}`, { token });
}

export async function createConversation(token: string, sellerId: string, listingId?: string) {
  const body: Record<string, string> = { seller_id: sellerId };
  if (listingId) body.listing_id = listingId;
  return request<Conversation>("/conversations", { token, method: "POST", body: JSON.stringify(body) });
}

export async function getMessages(token: string, convId: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Message>>(`/conversations/${convId}/messages?${params}`, { token });
}

export async function sendMessage(token: string, convId: string, content: string, type = "text") {
  return request<Message>(`/conversations/${convId}/messages`, {
    token,
    method: "POST",
    body: JSON.stringify({ content, type }),
  });
}

export async function deleteMessage(token: string, convId: string, msgId: string) {
  return request<void>(`/conversations/${convId}/messages/${msgId}`, { token, method: "DELETE" });
}

export async function recallMessage(token: string, convId: string, msgId: string) {
  return request<void>(`/conversations/${convId}/messages/${msgId}/recall`, { token, method: "PUT" });
}

export async function batchDeleteMessages(token: string, convId: string, messageIds: string[]) {
  return request<void>(`/conversations/${convId}/messages/batch-delete`, {
    token,
    method: "POST",
    body: JSON.stringify({ message_ids: messageIds }),
  });
}

export async function batchRecallMessages(token: string, convId: string, messageIds: string[]) {
  return request<void>(`/conversations/${convId}/messages/batch-recall`, {
    token,
    method: "POST",
    body: JSON.stringify({ message_ids: messageIds }),
  });
}

export async function markConversationRead(token: string, convId: string) {
  return request<void>(`/conversations/${convId}/read`, { token, method: "PUT" });
}

// --- Notifications ---
export async function getNotifications(token: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<AppNotification>>(`/notifications?${params}`, { token });
}

export async function markNotificationRead(token: string, id: string) {
  return request<void>(`/notifications/${id}/read`, { token, method: "PUT" });
}

// --- Subscription ---
export async function getSubscriptionStatus(token: string) {
  return request<SubscriptionStatus>("/subscription/status", { token });
}

export async function getSubscriptionPlans(token: string) {
  return request<{ plans: SubscriptionPlan[] }>("/subscription/plans", { token });
}

export interface SubscriptionHistory {
  id: string;
  plan_months: number;
  amount: number;
  starts_at: string;
  expires_at: string;
  status: string;
  created_at: string;
}

export async function getSubscriptionHistory(token: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<SubscriptionHistory>>(`/subscription/history?${params}`, { token });
}

// --- Ratings ---
export async function createRating(token: string, sellerId: string, stars: number, comment: string) {
  return request<Rating>("/ratings", {
    token,
    method: "POST",
    body: JSON.stringify({ seller_id: sellerId, stars, comment }),
  });
}

// --- Reports ---
export async function createReport(
  token: string,
  targetType: string,
  targetId: string,
  reason: string,
  description?: string
) {
  return request<void>("/reports", {
    token,
    method: "POST",
    body: JSON.stringify({ target_type: targetType, target_id: targetId, reason, description }),
  });
}

// --- Feedback ---
export async function createFeedback(token: string, content: string) {
  return request<Feedback>("/feedbacks", { token, method: "POST", body: JSON.stringify({ content }) });
}

export async function getMyFeedbacks(token: string, page: number, limit: number) {
  const params = new URLSearchParams({ page: String(page), limit: String(limit) });
  return request<PaginatedResponse<Feedback>>(`/feedbacks/my?${params}`, { token });
}

// --- Permissions ---
export type PermissionMap = Record<string, boolean>;

export async function getMyPermissions(token: string): Promise<{ role: string; permissions: PermissionMap }> {
  return request<{ role: string; permissions: PermissionMap }>("/permissions/me", { token });
}
