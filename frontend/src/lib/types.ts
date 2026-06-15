export interface User {
  id: string;
  name: string;
  role: 'client' | 'freelancer' | 'admin';
  avatar: string;
  bio: string;
  skills: string[];
  hourlyRate: number;
  rating: number;
  reviewsCount: number;
  location: string;
  joinedAt: string;
  isVerified: boolean;
  language?: string;
}

export interface UserFull extends User {
  email: string;
  earnings: number;
  spent: number;
  website: string;
  isOnline: boolean;
  spotifyConnected: boolean;
  spotifySharing: boolean;
  spotifyTrackName?: string;
  spotifyTrackArtist?: string;
  spotifyAlbumArt?: string;
  spotifyTrackUrl?: string;
}

export interface Gig {
  id: string;
  freelancerId: string;
  freelancer?: User;
  title: string;
  description: string;
  category: string;
  subcategory: string;
  tags: string[];
  aiTools: string[];
  priceType: string;
  price: number;
  deliveryDays: number;
  revisions: number;
  images: string[];
  videoUrl: string;
  status: string;
  ordersCount: number;
  views: number;
  featured: boolean;
  rating: number;
  reviewsCount: number;
  createdAt: string;
  updatedAt: string;
  packages?: Package[];
}

export interface Package {
  id: string;
  gigId: string;
  name: string;
  description: string;
  price: number;
  deliveryDays: number;
  revisions: number;
  features: string[];
}

export interface Order {
  id: string;
  gigId: string;
  gig?: Gig;
  packageId: string;
  buyerId: string;
  buyer?: User;
  sellerId: string;
  seller?: User;
  status: 'active' | 'delivered' | 'completed' | 'cancelled' | 'revision' | 'refunded';
  requirements: string;
  price: number;
  escrowStatus: string;
  deliveryFile: string;
  deliveryNotes: string;
  createdAt: string;
  deliveryDeadline: string;
  deliveredAt?: string;
  completedAt?: string;
  disputeId?: string;
}

export interface Review {
  id: string;
  orderId: string;
  gigId: string;
  reviewerId: string;
  reviewer?: User;
  revieweeId: string;
  rating: number;
  comment: string;
  createdAt: string;
}

export interface Message {
  id: string;
  senderId: string;
  receiverId: string;
  orderId?: string;
  content: string;
  fileUrl?: string;
  isRead: boolean;
  createdAt: string;
}

export interface Conversation {
  id: string;
  user1Id: string;
  user2Id: string;
  otherUser?: User;
  lastMessage: string;
  lastMessageAt: string;
  unreadCount: number;
}

export interface Notification {
  id: string;
  userId: string;
  type: string;
  title: string;
  message: string;
  link: string;
  isRead: boolean;
  createdAt: string;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  icon: string;
  description: string;
  gigsCount: number;
}

export interface Meeting {
  id: string;
  title: string;
  description: string;
  organizerId: string;
  organizer?: User;
  roomId: string;
  startTime: string;
  endTime: string;
  recurring: boolean;
  recurrenceRule?: string;
  attendeeIds: string[];
  status: 'scheduled' | 'active' | 'cancelled';
  createdAt: string;
}

export interface CoworkingSession {
  id: string;
  hostId: string;
  host?: User;
  type: 'focused' | 'pomodoro' | 'casual';
  title: string;
  zoneId: string;
  startTime: string;
  duration: number;
  participantIds: string[];
  maxParticipants: number;
  status: 'active' | 'completed' | 'cancelled';
  timerState?: TimerState;
  createdAt: string;
}

export interface TimerState {
  remaining: number;
  isPaused: boolean;
  phase: 'work' | 'break';
  lastResumed?: string;
}

export interface Dispute {
  id: string;
  orderId: string;
  openedBy: string;
  openedRole: 'client' | 'freelancer';
  reason: string;
  description: string;
  status: 'open' | 'evidence_pending' | 'under_review' | 'resolved' | 'escalated';
  resolution?: Resolution;
  createdAt: string;
  resolvedAt?: string;
}

export interface Resolution {
  ruling: 'full_refund' | 'full_release' | 'split';
  splitClient: number;
  adminNote: string;
  resolvedBy: string;
}

export interface Subscription {
  id: string;
  planId: string;
  clientId: string;
  freelancerId: string;
  status: 'active' | 'paused' | 'cancelled' | 'expired';
  currentPeriodStart: string;
  currentPeriodEnd: string;
  nextBillingDate: string;
  totalPaid: number;
  createdAt: string;
  cancelledAt?: string;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description: string;
  logo: string;
  ownerId: string;
  memberCount: number;
  createdAt: string;
}

export interface CaseStudy {
  id: string;
  userId: string;
  title: string;
  clientName: string;
  category: string;
  challenge: string;
  approach: string;
  results: string;
  images: string[];
  skills: string[];
  duration: string;
  budgetRange: string;
  linkedGigIds: string[];
  isPublic: boolean;
  createdAt: string;
}

export interface DashboardStats {
  totalOrders: number;
  activeOrders: number;
  completedOrders: number;
  totalEarnings: number;
  totalSpent: number;
  pendingPayouts: number;
  avgRating: number;
  totalReviews: number;
  totalGigs: number;
  totalViews: number;
  conversionRate: number;
  recentOrders: Order[];
}

export interface PaginatedResponse<T> {
  gigs?: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ApiResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}
