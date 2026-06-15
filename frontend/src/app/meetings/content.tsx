"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { motion, AnimatePresence, useMotionValue, useTransform, useSpring } from "framer-motion";
import { Calendar, Clock, Plus, Users, MapPin, Edit3, Trash2, Play, X, ChevronLeft, ChevronRight, Sparkles, Zap, Video, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

const ROOMS = [
  { id: "conf-1", name: "Conference Room A", icon: "🏢", color: "#a855f7" },
  { id: "conf-2", name: "Conference Room B", icon: "🏬", color: "#06b6d4" },
  { id: "conf-3", name: "Board Room", icon: "🏛️", color: "#ec4899" },
  { id: "virtual-1", name: "Virtual Meeting Room", icon: "🌐", color: "#10b981" },
];

const STATUS_STYLES: Record<string, { bg: string; text: string; glow: string }> = {
  scheduled: { bg: "from-blue-500/20 to-cyan-500/20", text: "text-blue-300", glow: "shadow-blue-500/20" },
  active: { bg: "from-green-500/20 to-emerald-500/20", text: "text-green-300", glow: "shadow-green-500/20" },
  completed: { bg: "from-gray-500/20 to-slate-500/20", text: "text-gray-400", glow: "shadow-gray-500/10" },
  cancelled: { bg: "from-red-500/20 to-rose-500/20", text: "text-red-300", glow: "shadow-red-500/20" },
};

function AnimatedCounter({ value, duration = 1.5 }: { value: number; duration?: number }) {
  const [display, setDisplay] = useState(0);
  useEffect(() => {
    let start = 0;
    const increment = value / (duration * 60);
    const timer = setInterval(() => {
      start += increment;
      if (start >= value) { setDisplay(value); clearInterval(timer); }
      else setDisplay(Math.floor(start));
    }, 1000 / 60);
    return () => clearInterval(timer);
  }, [value, duration]);
  return <span>{display}</span>;
}

function FloatingOrb({ delay, x, y, size, color }: { delay: number; x: string; y: string; size: number; color: string }) {
  return (
    <motion.div
      className="absolute rounded-full blur-3xl opacity-20 pointer-events-none"
      style={{ left: x, top: y, width: size, height: size, background: color }}
      animate={{ y: [0, -30, 0], x: [0, 15, 0], scale: [1, 1.1, 1] }}
      transition={{ duration: 8, repeat: Infinity, delay, ease: "easeInOut" }}
    />
  );
}

export default function MeetingsContent() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [meetings, setMeetings] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingMeeting, setEditingMeeting] = useState<any>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [slideDirection, setSlideDirection] = useState(0);
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);
  const rotateX = useTransform(mouseY, [-300, 300], [5, -5]);
  const rotateY = useTransform(mouseX, [-300, 300], [-5, 5]);

  const [formData, setFormData] = useState({
    title: "", description: "", roomId: "", startTime: "", endTime: "",
    recurring: false, recurrenceRule: "", attendeeIds: [] as string[],
  });

  useEffect(() => {
    if (!authLoading && !user) { router.push("/auth/login"); return; }
    loadMeetings();
  }, [user, authLoading]);

  const loadMeetings = async () => {
    setLoading(true);
    try { setMeetings(await api.getMeetings()); } catch {}
    setLoading(false);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setActionLoading("submit");
    try {
      if (editingMeeting) await api.updateMeeting(editingMeeting.id, formData);
      else await api.createMeeting(formData);
      setShowForm(false); setEditingMeeting(null); resetForm(); loadMeetings();
    } catch (err: any) { alert(err.message); }
    setActionLoading(null);
  };

  const resetForm = () => setFormData({ title: "", description: "", roomId: "", startTime: "", endTime: "", recurring: false, recurrenceRule: "", attendeeIds: [] });

  const handleEdit = (meeting: any) => {
    setEditingMeeting(meeting);
    setFormData({
      title: meeting.title, description: meeting.description, roomId: meeting.roomId,
      startTime: meeting.startTime?.slice(0, 16) || "", endTime: meeting.endTime?.slice(0, 16) || "",
      recurring: meeting.recurring, recurrenceRule: meeting.recurrenceRule || "", attendeeIds: meeting.attendeeIds || [],
    });
    setShowForm(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Cancel this meeting?")) return;
    setActionLoading(id);
    try { await api.deleteMeeting(id); loadMeetings(); } catch (err: any) { alert(err.message); }
    setActionLoading(null);
  };

  const handleJoin = async (id: string) => {
    setActionLoading(id);
    try {
      const result = await api.joinMeeting(id);
      if (result.roomId) router.push(`/world?room=${result.roomId}`);
      loadMeetings();
    } catch (err: any) { alert(err.message); }
    setActionLoading(null);
  };

  const getDaysInMonth = (date: Date) => {
    const year = date.getFullYear(), month = date.getMonth();
    return { daysInMonth: new Date(year, month + 1, 0).getDate(), startDay: new Date(year, month, 1).getDay(), year, month };
  };

  const getMeetingsForDay = (day: number) => {
    const { year, month } = getDaysInMonth(currentMonth);
    return meetings.filter(m => m.startTime?.startsWith(`${year}-${String(month + 1).padStart(2, "0")}-${String(day).padStart(2, "0")}`));
  };

  const upcomingMeetings = meetings
    .filter(m => m.status === "scheduled" && new Date(m.startTime) > new Date())
    .sort((a, b) => new Date(a.startTime).getTime() - new Date(b.startTime).getTime())
    .slice(0, 5);

  const thisWeekMeetings = meetings.filter(m => {
    const d = new Date(m.startTime);
    const now = new Date();
    const weekEnd = new Date(now.getTime() + 7 * 86400000);
    return d >= now && d <= weekEnd;
  });

  const activeMeetings = meetings.filter(m => m.status === "active");

  const navigateMonth = (dir: number) => {
    setSlideDirection(dir);
    setCurrentMonth(new Date(currentMonth.getFullYear(), currentMonth.getMonth() + dir));
  };

  if (authLoading || loading) return (
    <div className="min-h-screen bg-[#0a0a0f] flex items-center justify-center">
      <motion.div animate={{ rotate: 360 }} transition={{ duration: 2, repeat: Infinity, ease: "linear" }} className="w-12 h-12 rounded-full border-2 border-transparent border-t-purple-500 border-r-cyan-500" />
    </div>
  );

  const { daysInMonth, startDay, year, month } = getDaysInMonth(currentMonth);
  const monthNames = ["January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"];

  return (
    <div className="min-h-screen bg-[#0a0a0f] text-white overflow-hidden">
      <FloatingOrb delay={0} x="10%" y="20%" size={400} color="linear-gradient(135deg, #a855f7, #06b6d4)" />
      <FloatingOrb delay={2} x="70%" y="60%" size={350} color="linear-gradient(135deg, #ec4899, #a855f7)" />
      <FloatingOrb delay={4} x="50%" y="10%" size={300} color="linear-gradient(135deg, #06b6d4, #10b981)" />

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 pb-16">
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ duration: 0.8 }}>
          <motion.div className="text-center mb-16" initial={{ y: 40, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}>
            <motion.div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-white/5 border border-white/10 mb-6" whileHover={{ scale: 1.05, borderColor: "rgba(168,85,247,0.5)" }}>
              <Sparkles className="w-4 h-4 text-purple-400" />
              <span className="text-sm text-gray-400">Virtual Collaboration Hub</span>
            </motion.div>
            <h1 className="text-7xl sm:text-8xl lg:text-9xl font-black tracking-tighter leading-none mb-4">
              <span className="bg-gradient-to-r from-purple-400 via-cyan-400 to-pink-400 bg-clip-text text-transparent">MEET</span>
              <span className="text-white/20">INGS</span>
            </h1>
            <p className="text-lg text-gray-500 max-w-md mx-auto">Schedule, collaborate, and connect in immersive virtual spaces</p>
          </motion.div>

          <motion.div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-12" initial={{ y: 30, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.2, duration: 0.6 }}>
            {[
              { label: "Total Meetings", value: meetings.length, icon: Video, color: "from-purple-500 to-purple-600" },
              { label: "This Week", value: thisWeekMeetings.length, icon: Calendar, color: "from-cyan-500 to-blue-500" },
              { label: "Active Now", value: activeMeetings.length, icon: Zap, color: "from-pink-500 to-rose-500" },
            ].map((stat, i) => (
              <motion.div key={stat.label} className="relative group" whileHover={{ y: -4 }} transition={{ type: "spring", stiffness: 400 }}>
                <div className={`absolute inset-0 bg-gradient-to-br ${stat.color} rounded-2xl blur-xl opacity-0 group-hover:opacity-20 transition-opacity duration-500`} />
                <div className="relative p-6 rounded-2xl bg-white/[0.03] border border-white/[0.06] backdrop-blur-xl hover:border-white/10 transition-colors">
                  <div className="flex items-center justify-between mb-4">
                    <div className={`p-2.5 rounded-xl bg-gradient-to-br ${stat.color} bg-opacity-10`}>
                      <stat.icon className="w-5 h-5 text-white" />
                    </div>
                    <motion.div className="text-4xl font-black bg-gradient-to-r from-white to-white/60 bg-clip-text text-transparent" initial={{ scale: 0.5 }} animate={{ scale: 1 }} transition={{ delay: 0.3 + i * 0.1, type: "spring" }}>
                      <AnimatedCounter value={stat.value} />
                    </motion.div>
                  </div>
                  <p className="text-sm text-gray-500 font-medium">{stat.label}</p>
                </div>
              </motion.div>
            ))}
          </motion.div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
            <motion.div className="lg:col-span-2" initial={{ y: 30, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.3 }} style={{ perspective: 1000 }}>
              <motion.div className="rounded-3xl bg-white/[0.02] border border-white/[0.06] backdrop-blur-xl overflow-hidden" style={{ rotateX, rotateY }} onMouseMove={e => { mouseX.set(e.clientX - window.innerWidth / 2); mouseY.set(e.clientY - window.innerHeight / 2); }} onMouseLeave={() => { mouseX.set(0); mouseY.set(0); }}>
                <div className="p-8">
                  <div className="flex items-center justify-between mb-8">
                    <div>
                      <h2 className="text-3xl font-bold tracking-tight">{monthNames[month]}</h2>
                      <p className="text-5xl font-black text-white/10">{year}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <motion.button whileHover={{ scale: 1.1 }} whileTap={{ scale: 0.95 }} onClick={() => navigateMonth(-1)} className="p-3 rounded-xl bg-white/5 border border-white/10 hover:bg-white/10 hover:border-purple-500/30 transition-all">
                        <ChevronLeft className="w-5 h-5" />
                      </motion.button>
                      <motion.button whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }} onClick={() => { setSlideDirection(0); setCurrentMonth(new Date()); }} className="px-5 py-3 rounded-xl bg-gradient-to-r from-purple-500/20 to-cyan-500/20 border border-purple-500/30 text-sm font-semibold hover:from-purple-500/30 hover:to-cyan-500/30 transition-all">
                        Today
                      </motion.button>
                      <motion.button whileHover={{ scale: 1.1 }} whileTap={{ scale: 0.95 }} onClick={() => navigateMonth(1)} className="p-3 rounded-xl bg-white/5 border border-white/10 hover:bg-white/10 hover:border-cyan-500/30 transition-all">
                        <ChevronRight className="w-5 h-5" />
                      </motion.button>
                    </div>
                  </div>

                  <div className="grid grid-cols-7 gap-2 mb-3">
                    {["SUN", "MON", "TUE", "WED", "THU", "FRI", "SAT"].map(d => (
                      <div key={d} className="text-center text-[11px] font-bold tracking-widest text-gray-600 py-2">{d}</div>
                    ))}
                  </div>

                  <AnimatePresence mode="wait" custom={slideDirection}>
                    <motion.div key={`${year}-${month}`} custom={slideDirection} initial={{ x: slideDirection * 100, opacity: 0 }} animate={{ x: 0, opacity: 1 }} exit={{ x: slideDirection * -100, opacity: 0 }} transition={{ duration: 0.3, ease: [0.16, 1, 0.3, 1] }} className="grid grid-cols-7 gap-2">
                      {Array.from({ length: startDay }).map((_, i) => <div key={`empty-${i}`} className="h-24 rounded-xl bg-white/[0.01]" />)}
                      {Array.from({ length: daysInMonth }).map((_, i) => {
                        const day = i + 1;
                        const dayMeetings = getMeetingsForDay(day);
                        const isToday = new Date().getDate() === day && new Date().getMonth() === month && new Date().getFullYear() === year;
                        return (
                          <motion.div key={day} whileHover={{ scale: 1.05, backgroundColor: "rgba(168,85,247,0.08)" }} className={`relative h-24 rounded-xl border p-2 cursor-pointer transition-all duration-300 ${isToday ? "border-purple-500/50 bg-purple-500/[0.06] shadow-lg shadow-purple-500/10" : "border-white/[0.04] bg-white/[0.01] hover:border-white/10"}`}>
                            {isToday && <motion.div className="absolute inset-0 rounded-xl border-2 border-purple-500/30" animate={{ opacity: [0.3, 0.8, 0.3] }} transition={{ duration: 2, repeat: Infinity }} />}
                            <div className={`text-xs font-bold mb-1.5 ${isToday ? "text-purple-400" : "text-gray-500"}`}>{day}</div>
                            <div className="space-y-1">
                              {dayMeetings.slice(0, 2).map(m => (
                                <motion.div key={m.id} whileHover={{ x: 2 }} onClick={() => handleEdit(m)} className="text-[10px] px-1.5 py-1 rounded-lg bg-gradient-to-r from-purple-500/20 to-cyan-500/20 text-purple-300 truncate font-medium border border-purple-500/10 cursor-pointer">
                                  {m.title}
                                </motion.div>
                              ))}
                              {dayMeetings.length > 2 && <div className="text-[10px] text-cyan-400 font-semibold">+{dayMeetings.length - 2} more</div>}
                            </div>
                          </motion.div>
                        );
                      })}
                    </motion.div>
                  </AnimatePresence>
                </div>
              </motion.div>
            </motion.div>

            <motion.div className="space-y-6" initial={{ y: 30, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.4 }}>
              <div className="rounded-3xl bg-white/[0.02] border border-white/[0.06] backdrop-blur-xl overflow-hidden">
                <div className="p-6">
                  <div className="flex items-center gap-3 mb-6">
                    <div className="p-2 rounded-lg bg-gradient-to-br from-purple-500 to-cyan-500">
                      <Clock className="w-4 h-4 text-white" />
                    </div>
                    <h3 className="text-lg font-bold tracking-tight">Upcoming</h3>
                  </div>
                  {upcomingMeetings.length === 0 ? (
                    <div className="text-center py-8">
                      <div className="text-4xl mb-3">📭</div>
                      <p className="text-sm text-gray-500">No upcoming meetings</p>
                    </div>
                  ) : (
                    <div className="space-y-3">
                      {upcomingMeetings.map((m, i) => (
                        <motion.div key={m.id} initial={{ x: 20, opacity: 0 }} animate={{ x: 0, opacity: 1 }} transition={{ delay: 0.5 + i * 0.1 }} whileHover={{ x: 4 }} className="group relative">
                          <div className="absolute left-0 top-0 bottom-0 w-1 rounded-full bg-gradient-to-b from-purple-500 to-cyan-500 opacity-0 group-hover:opacity-100 transition-opacity" />
                          <div className="pl-4 pr-3 py-3 rounded-xl bg-white/[0.02] border border-white/[0.04] hover:border-purple-500/20 hover:bg-white/[0.04] transition-all cursor-pointer" onClick={() => handleEdit(m)}>
                            <div className="flex items-start justify-between mb-2">
                              <p className="font-semibold text-sm text-white/90 group-hover:text-white transition-colors">{m.title}</p>
                              <span className={`px-2 py-0.5 rounded-full text-[10px] font-bold bg-gradient-to-r ${STATUS_STYLES[m.status]?.bg || STATUS_STYLES.scheduled.bg} ${STATUS_STYLES[m.status]?.text || STATUS_STYLES.scheduled.text}`}>
                                {m.status}
                              </span>
                            </div>
                            <div className="flex items-center gap-3 text-[11px] text-gray-500">
                              <span className="flex items-center gap-1"><Clock className="w-3 h-3" />{new Date(m.startTime).toLocaleString([], { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" })}</span>
                              {m.roomId && <span className="flex items-center gap-1"><MapPin className="w-3 h-3" />{ROOMS.find(r => r.id === m.roomId)?.name || m.roomId}</span>}
                            </div>
                            {m.status === "scheduled" && (
                              <motion.button whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }} onClick={e => { e.stopPropagation(); handleJoin(m.id); }} disabled={actionLoading === m.id} className="w-full mt-3 py-2 rounded-lg bg-gradient-to-r from-purple-500 to-cyan-500 text-white text-xs font-bold flex items-center justify-center gap-2 hover:shadow-lg hover:shadow-purple-500/20 transition-shadow disabled:opacity-50">
                                <Play className="w-3 h-3" /> Join Meeting <ArrowRight className="w-3 h-3" />
                              </motion.button>
                            )}
                          </div>
                        </motion.div>
                      ))}
                    </div>
                  )}
                </div>
              </div>

              <div className="rounded-3xl bg-white/[0.02] border border-white/[0.06] backdrop-blur-xl overflow-hidden">
                <div className="p-6">
                  <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center gap-3">
                      <div className="p-2 rounded-lg bg-gradient-to-br from-pink-500 to-rose-500">
                        <Users className="w-4 h-4 text-white" />
                      </div>
                      <h3 className="text-lg font-bold tracking-tight">All Meetings</h3>
                    </div>
                    <span className="text-xs text-gray-500 bg-white/5 px-2.5 py-1 rounded-full font-medium">{meetings.length}</span>
                  </div>
                  <div className="space-y-2 max-h-72 overflow-y-auto custom-scrollbar">
                    {meetings.map((m, i) => (
                      <motion.div key={m.id} initial={{ y: 10, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.6 + i * 0.05 }} whileHover={{ x: 4 }} className="group flex items-center justify-between p-3 rounded-xl hover:bg-white/[0.03] transition-all cursor-pointer" onClick={() => handleEdit(m)}>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-white/80 group-hover:text-white truncate transition-colors">{m.title}</p>
                          <p className="text-[11px] text-gray-600 mt-0.5">{new Date(m.startTime).toLocaleDateString(undefined, { month: "short", day: "numeric", year: "numeric" })}</p>
                        </div>
                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                          {m.status === "scheduled" && (
                            <motion.button whileHover={{ scale: 1.2 }} whileTap={{ scale: 0.9 }} onClick={e => { e.stopPropagation(); handleJoin(m.id); }} className="p-1.5 rounded-lg bg-green-500/10 text-green-400 hover:bg-green-500/20">
                              <Play className="w-3 h-3" />
                            </motion.button>
                          )}
                          <motion.button whileHover={{ scale: 1.2 }} whileTap={{ scale: 0.9 }} onClick={e => { e.stopPropagation(); handleEdit(m); }} className="p-1.5 rounded-lg bg-blue-500/10 text-blue-400 hover:bg-blue-500/20">
                            <Edit3 className="w-3 h-3" />
                          </motion.button>
                          <motion.button whileHover={{ scale: 1.2 }} whileTap={{ scale: 0.9 }} onClick={e => { e.stopPropagation(); handleDelete(m.id); }} className="p-1.5 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20">
                            <Trash2 className="w-3 h-3" />
                          </motion.button>
                        </div>
                      </motion.div>
                    ))}
                  </div>
                </div>
              </div>
            </motion.div>
          </div>

          <motion.div className="flex justify-center" initial={{ y: 20, opacity: 0 }} animate={{ y: 0, opacity: 1 }} transition={{ delay: 0.7 }}>
            <motion.button whileHover={{ scale: 1.05, boxShadow: "0 0 40px rgba(168,85,247,0.3)" }} whileTap={{ scale: 0.95 }} onClick={() => { resetForm(); setEditingMeeting(null); setShowForm(true); }} className="group relative px-8 py-4 rounded-2xl bg-gradient-to-r from-purple-500 via-cyan-500 to-pink-500 text-white font-bold text-lg overflow-hidden">
              <motion.div className="absolute inset-0 bg-gradient-to-r from-pink-500 via-purple-500 to-cyan-500" initial={{ x: "100%" }} whileHover={{ x: 0 }} transition={{ duration: 0.4 }} />
              <span className="relative flex items-center gap-3"><Plus className="w-5 h-5" /> Create New Meeting</span>
            </motion.button>
          </motion.div>
        </motion.div>
      </div>

      <AnimatePresence>
        {showForm && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={() => setShowForm(false)}>
            <motion.div className="absolute inset-0 bg-black/80 backdrop-blur-xl" />
            <motion.div className="absolute inset-0 opacity-30" style={{ background: "radial-gradient(circle at 30% 50%, rgba(168,85,247,0.15), transparent 50%), radial-gradient(circle at 70% 50%, rgba(6,182,212,0.15), transparent 50%)" }} />
            
            <motion.div initial={{ y: 100, opacity: 0, scale: 0.9 }} animate={{ y: 0, opacity: 1, scale: 1 }} exit={{ y: 100, opacity: 0, scale: 0.9 }} transition={{ type: "spring", damping: 25, stiffness: 300 }} className="relative w-full max-w-lg" onClick={e => e.stopPropagation()}>
              <div className="rounded-3xl bg-[#0f0f1a] border border-white/[0.08] shadow-2xl shadow-purple-500/10 overflow-hidden">
                <div className="relative p-8">
                  <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-purple-500 via-cyan-500 to-pink-500" />
                  
                  <div className="flex items-center justify-between mb-8">
                    <div>
                      <h2 className="text-2xl font-black tracking-tight">{editingMeeting ? "Edit Meeting" : "New Meeting"}</h2>
                      <p className="text-sm text-gray-500 mt-1">{editingMeeting ? "Update meeting details" : "Schedule a new collaboration"}</p>
                    </div>
                    <motion.button whileHover={{ scale: 1.1, rotate: 90 }} whileTap={{ scale: 0.9 }} onClick={() => setShowForm(false)} className="p-2 rounded-xl bg-white/5 border border-white/10 hover:bg-white/10 transition-colors">
                      <X className="w-5 h-5" />
                    </motion.button>
                  </div>

                  <form onSubmit={handleSubmit} className="space-y-5">
                    <div className="space-y-2">
                      <label className="text-xs font-bold tracking-widest text-gray-500 uppercase">Title</label>
                      <Input value={formData.title} onChange={e => setFormData({ ...formData, title: e.target.value })} placeholder="Enter meeting title" required className="h-12 rounded-xl bg-white/[0.03] border-white/[0.06] focus:border-purple-500/50 focus:ring-purple-500/20 text-white placeholder:text-gray-600" />
                    </div>

                    <div className="space-y-2">
                      <label className="text-xs font-bold tracking-widest text-gray-500 uppercase">Description</label>
                      <textarea value={formData.description} onChange={e => setFormData({ ...formData, description: e.target.value })} placeholder="What's this meeting about?" className="w-full h-24 px-4 py-3 rounded-xl bg-white/[0.03] border border-white/[0.06] text-sm text-white placeholder:text-gray-600 resize-none focus:outline-none focus:border-purple-500/50 focus:ring-2 focus:ring-purple-500/20 transition-all" />
                    </div>

                    <div className="space-y-2">
                      <label className="text-xs font-bold tracking-widest text-gray-500 uppercase flex items-center gap-2"><MapPin className="w-3 h-3" /> Room</label>
                      <div className="grid grid-cols-2 gap-2">
                        {ROOMS.map(r => (
                          <motion.button key={r.id} type="button" whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }} onClick={() => setFormData({ ...formData, roomId: r.id })} className={`p-3 rounded-xl border text-left transition-all ${formData.roomId === r.id ? "border-purple-500/50 bg-purple-500/10 shadow-lg shadow-purple-500/10" : "border-white/[0.04] bg-white/[0.02] hover:border-white/10"}`}>
                            <div className="text-lg mb-1">{r.icon}</div>
                            <div className="text-xs font-semibold text-white/80">{r.name}</div>
                          </motion.button>
                        ))}
                      </div>
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-xs font-bold tracking-widest text-gray-500 uppercase flex items-center gap-2"><Calendar className="w-3 h-3" /> Start</label>
                        <Input type="datetime-local" value={formData.startTime} onChange={e => setFormData({ ...formData, startTime: e.target.value })} required className="h-12 rounded-xl bg-white/[0.03] border-white/[0.06] focus:border-cyan-500/50 focus:ring-cyan-500/20 text-white" />
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs font-bold tracking-widest text-gray-500 uppercase flex items-center gap-2"><Clock className="w-3 h-3" /> End</label>
                        <Input type="datetime-local" value={formData.endTime} onChange={e => setFormData({ ...formData, endTime: e.target.value })} required className="h-12 rounded-xl bg-white/[0.03] border-white/[0.06] focus:border-pink-500/50 focus:ring-pink-500/20 text-white" />
                      </div>
                    </div>

                    <div className="flex items-center gap-3 p-3 rounded-xl bg-white/[0.02] border border-white/[0.04]">
                      <div className="relative">
                        <input type="checkbox" id="recurring" checked={formData.recurring} onChange={e => setFormData({ ...formData, recurring: e.target.checked })} className="peer sr-only" />
                        <label htmlFor="recurring" className="block w-10 h-6 rounded-full bg-white/10 cursor-pointer transition-colors peer-checked:bg-purple-500 after:content-[''] after:block after:w-4 after:h-4 after:rounded-full after:bg-white after:mt-1 after:ml-1 after:transition-transform peer-checked:after:translate-x-4" />
                      </div>
                      <label htmlFor="recurring" className="text-sm font-medium text-white/80 cursor-pointer">Recurring meeting</label>
                    </div>

                    {formData.recurring && (
                      <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: "auto", opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="space-y-2">
                        <label className="text-xs font-bold tracking-widest text-gray-500 uppercase">Recurrence</label>
                        <div className="grid grid-cols-4 gap-2">
                          {["daily", "weekly", "biweekly", "monthly"].map(rule => (
                            <motion.button key={rule} type="button" whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }} onClick={() => setFormData({ ...formData, recurrenceRule: rule })} className={`py-2 rounded-lg text-xs font-semibold capitalize transition-all ${formData.recurrenceRule === rule ? "bg-gradient-to-r from-purple-500 to-cyan-500 text-white" : "bg-white/5 text-gray-400 hover:bg-white/10 border border-white/[0.04]"}`}>
                              {rule}
                            </motion.button>
                          ))}
                        </div>
                      </motion.div>
                    )}

                    <div className="flex gap-3 pt-4">
                      <motion.button type="button" whileHover={{ scale: 1.02 }} whileTap={{ scale: 0.98 }} onClick={() => setShowForm(false)} className="flex-1 h-12 rounded-xl bg-white/5 border border-white/10 text-sm font-semibold hover:bg-white/10 transition-colors">
                        Cancel
                      </motion.button>
                      <motion.button type="submit" whileHover={{ scale: 1.02, boxShadow: "0 0 30px rgba(168,85,247,0.3)" }} whileTap={{ scale: 0.98 }} disabled={actionLoading === "submit"} className="flex-1 h-12 rounded-xl bg-gradient-to-r from-purple-500 via-cyan-500 to-pink-500 text-white font-bold text-sm hover:shadow-lg transition-shadow disabled:opacity-50">
                        {actionLoading === "submit" ? (
                          <motion.div animate={{ rotate: 360 }} transition={{ duration: 1, repeat: Infinity, ease: "linear" }} className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full mx-auto" />
                        ) : (
                          editingMeeting ? "Update Meeting" : "Create Meeting"
                        )}
                      </motion.button>
                    </div>
                  </form>
                </div>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      <style jsx global>{`
        .custom-scrollbar::-webkit-scrollbar { width: 4px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(168,85,247,0.3); border-radius: 100px; }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(168,85,247,0.5); }
      `}</style>
    </div>
  );
}
