const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3001/api/v1";

class ApiClient {
  private token: string | null = null;

  constructor() {
    if (typeof window !== "undefined") {
      this.token = localStorage.getItem("token");
    }
  }

  setToken(token: string) {
    this.token = token;
    if (typeof window !== "undefined") {
      localStorage.setItem("token", token);
    }
  }

  clearToken() {
    this.token = null;
    if (typeof window !== "undefined") {
      localStorage.removeItem("token");
    }
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const headers: Record<string, string> = { "Content-Type": "application/json", ...((options.headers as Record<string, string>) || {}) };
    if (this.token) headers["Authorization"] = `Bearer ${this.token}`;
    const res = await fetch(`${API_BASE}${endpoint}`, { ...options, headers });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Request failed" }));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    return res.json();
  }

  // Auth
  register = (data: { email: string; password: string; name: string; role?: string }) =>
    this.request<{ token: string; user: any }>(`/auth/register?role=${data.role || "client"}`, { method: "POST", body: JSON.stringify(data) });

  login = (data: { email: string; password: string }) =>
    this.request<{ token: string; user: any }>("/auth/login", { method: "POST", body: JSON.stringify(data) });

  // Profile
  getProfile = (id: string) => this.request<any>(`/users/${id}`);
  getMe = () => this.request<any>("/me");
  updateProfile = (data: any) => this.request<any>("/me", { method: "PUT", body: JSON.stringify(data) });

  // Categories & Freelancers
  getCategories = () => this.request<any[]>("/categories");
  getFreelancers = () => this.request<any[]>("/freelancers");

  // Gigs
  getGigs = (params?: Record<string, string>) => {
    const qs = params ? "?" + new URLSearchParams(params).toString() : "";
    return this.request<any>(`/gigs${qs}`);
  };
  getGig = (id: string) => this.request<any>(`/gigs/${id}`);
  createGig = (data: any) => this.request<any>("/gigs", { method: "POST", body: JSON.stringify(data) });
  updateGig = (id: string, data: any) => this.request<any>(`/gigs/${id}`, { method: "PUT", body: JSON.stringify(data) });
  deleteGig = (id: string) => this.request<any>(`/gigs/${id}`, { method: "DELETE" });
  getGigReviews = (id: string) => this.request<any[]>(`/gigs/${id}/reviews`);
  toggleFavorite = (id: string) => this.request<any>(`/favorites/${id}`, { method: "POST" });

  // Dashboard
  getDashboard = () => this.request<any>("/dashboard");
  getMyGigs = () => this.request<any[]>("/my-gigs");
  getEarnings = () => this.request<any>("/earnings");
  getSpending = () => this.request<any>("/spending");

  // Orders
  createOrder = (data: any) => this.request<any>("/orders", { method: "POST", body: JSON.stringify(data) });
  getOrders = (status?: string) => this.request<any[]>(`/orders${status ? `?status=${status}` : ""}`);
  getOrder = (id: string) => this.request<any>(`/orders/${id}`);
  deliverOrder = (id: string, data: any) => this.request<any>(`/orders/${id}/deliver`, { method: "POST", body: JSON.stringify(data) });
  approveOrder = (id: string) => this.request<any>(`/orders/${id}/approve`, { method: "POST" });
  requestRevision = (id: string, data: any) => this.request<any>(`/orders/${id}/revision`, { method: "POST", body: JSON.stringify(data) });
  createReview = (id: string, data: any) => this.request<any>(`/orders/${id}/review`, { method: "POST", body: JSON.stringify(data) });

  // Messages
  getConversations = () => this.request<any[]>("/conversations");
  getMessages = (userId: string) => this.request<any[]>(`/messages/${userId}`);
  sendMessage = (data: { receiverId: string; content: string; orderId?: string }) =>
    this.request<any>("/messages", { method: "POST", body: JSON.stringify(data) });
  getUnreadCount = () => this.request<{ count: number }>("/messages/unread");
  markConversationRead = (userId: string) => this.request<any>(`/conversations/${userId}/read`, { method: "POST" });

  // Notifications
  getNotifications = () => this.request<any[]>("/notifications");
  markNotificationsRead = () => this.request<any>("/notifications/read", { method: "POST" });

  // Spotify
  getSpotifyStatus = () => this.request<any>("/spotify/status");
  shareSpotifyTrack = (share: boolean) => this.request<any>("/spotify/share", { method: "POST", body: JSON.stringify({ share }) });
  disconnectSpotify = () => this.request<any>("/spotify/disconnect", { method: "POST" });
  spotifyPlayback = (action: string) => this.request<any>("/spotify/playback", { method: "POST", body: JSON.stringify({ action }) });
  searchSpotify = (query: string) => this.request<any>(`/spotify/search?q=${encodeURIComponent(query)}`);

  // Meetings
  createMeeting = (data: any) => this.request<any>("/meetings", { method: "POST", body: JSON.stringify(data) });
  getMeetings = () => this.request<any[]>("/meetings");
  getMeeting = (id: string) => this.request<any>(`/meetings/${id}`);
  updateMeeting = (id: string, data: any) => this.request<any>(`/meetings/${id}`, { method: "PUT", body: JSON.stringify(data) });
  deleteMeeting = (id: string) => this.request<any>(`/meetings/${id}`, { method: "DELETE" });
  joinMeeting = (id: string) => this.request<any>(`/meetings/${id}/join`, { method: "POST" });

  // Admin
  getMembers = () => this.request<{ members: any[]; total: number }>("/admin");
  inviteMember = (data: { email: string; role: string }) =>
    this.request<any>("/admin/members/invite", { method: "POST", body: JSON.stringify(data) });
  changeMemberRole = (id: string, role: string) =>
    this.request<any>(`/admin/members/${id}/role`, { method: "PUT", body: JSON.stringify({ role }) });
  removeMember = (id: string) =>
    this.request<any>(`/admin/members/${id}`, { method: "DELETE" });
  demoteMember = (id: string) =>
    this.request<any>(`/admin/members/${id}/demote`, { method: "POST" });
  getAdminSettings = () => this.request<any>("/admin/settings");
  updateAdminSettings = (data: any) =>
    this.request<any>("/admin/settings", { method: "PUT", body: JSON.stringify(data) });

  // Coworking
  getCoworkingSessions = () => this.request<any[]>("/coworking");
  getCoworkingSession = (id: string) => this.request<any>(`/coworking/${id}`);
  createCoworkingSession = (data: { type: string; title: string; zoneId: string; duration: number }) =>
    this.request<any>("/coworking", { method: "POST", body: JSON.stringify(data) });
  joinCoworkingSession = (id: string) => this.request<any>(`/coworking/${id}/join`, { method: "POST" });
  leaveCoworkingSession = (id: string) => this.request<any>(`/coworking/${id}/leave`, { method: "POST" });
  updateCoworkingTimer = (id: string, data: { isPaused?: boolean; phase?: string }) =>
    this.request<any>(`/coworking/${id}/timer`, { method: "POST", body: JSON.stringify(data) });
  endCoworkingSession = (id: string) => this.request<any>(`/coworking/${id}/end`, { method: "POST" });

  // Realtime WebSocket
  connectRealtime = (userId: string, onMessage: (msg: any) => void): WebSocket | null => {
    if (typeof window === "undefined") return null;
    const wsBase = process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:3001";
    if (!this.token) return null;
    const ws = new WebSocket(`${wsBase}/api/v1/realtime/ws?token=${this.token}`);
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        onMessage(data);
      } catch {}
    };
    ws.onerror = () => {};
    return ws;
  };
}

export const api = new ApiClient();
