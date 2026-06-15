"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { Menu, X, Sparkles, MessageSquare, Bell, LayoutDashboard, LogOut, ChevronDown, Globe, Calendar, Sun, Moon, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api";
import { useTheme } from "@/lib/theme-context";

const navLinks = [
  { href: "/marketplace", label: "Marketplace" },
  { href: "/world", label: "World" },
  { href: "/meetings", label: "Meetings" },
  { href: "/dashboard", label: "Dashboard" },
];

export default function Navbar() {
  const { user, logout, loading } = useAuth();
  const pathname = usePathname();
  const { theme, toggleTheme } = useTheme();
  const [scrolled, setScrolled] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [unread, setUnread] = useState(0);
  const [showMenu, setShowMenu] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 20);
    window.addEventListener("scroll", onScroll);
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  useEffect(() => {
    if (user) {
      api.getUnreadCount().then(r => setUnread(r.count)).catch(() => {});
      const interval = setInterval(() => {
        api.getUnreadCount().then(r => setUnread(r.count)).catch(() => {});
      }, 30000);
      return () => clearInterval(interval);
    }
  }, [user]);

  const isHome = pathname === "/";

  return (
    <motion.header
      initial={{ y: -100, opacity: 0 }}
      animate={{ y: 0, opacity: 1 }}
      transition={{ duration: 0.6, ease: [0.25, 0.46, 0.45, 0.94] }}
      className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ${
        scrolled || !isHome ? "bg-background/80 backdrop-blur-xl border-b border-white/5 shadow-2xl shadow-black/20" : "bg-transparent"
      }`}
    >
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 lg:h-20">
          <Link href="/" className="flex items-center gap-2.5">
            <div className="relative w-9 h-9">
              <div className="absolute inset-0 bg-gradient-to-br from-primary via-blue-500 to-emerald-500 rounded-xl rotate-6 opacity-70" />
              <div className="absolute inset-0 bg-gradient-to-br from-primary via-blue-500 to-emerald-500 rounded-xl flex items-center justify-center">
                <Sparkles className="w-5 h-5 text-white" />
              </div>
            </div>
            <span className="text-xl font-bold tracking-tight">
              Nova<span className="text-primary">Field</span>
            </span>
          </Link>

          <nav className="hidden md:flex items-center gap-1">
            {navLinks.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                className={`relative px-4 py-2 text-sm font-medium rounded-lg transition-colors ${
                  pathname.startsWith(link.href) ? "text-foreground bg-white/5" : "text-muted-foreground hover:text-foreground hover:bg-white/5"
                }`}
              >
                {link.label}
              </Link>
            ))}
          </nav>

          <div className="hidden md:flex items-center gap-3">
            <button
              onClick={toggleTheme}
              className="p-2 rounded-lg hover:bg-white/5 transition-colors"
              title={theme === "dark" ? "Switch to light mode" : "Switch to dark mode"}
            >
              {theme === "dark" ? (
                <Sun className="w-5 h-5 text-muted-foreground" />
              ) : (
                <Moon className="w-5 h-5 text-muted-foreground" />
              )}
            </button>
            {!loading && (
              <>
                {user ? (
                  <>
                    <Link href="/messages" className="relative p-2 rounded-lg hover:bg-white/5 transition-colors">
                      <MessageSquare className="w-5 h-5 text-muted-foreground" />
                      {unread > 0 && (
                        <span className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-primary text-[9px] flex items-center justify-center font-bold">
                          {unread > 9 ? "9+" : unread}
                        </span>
                      )}
                    </Link>
                    <div className="relative">
                      <button onClick={() => setShowMenu(!showMenu)} className="flex items-center gap-2 p-1.5 rounded-lg hover:bg-white/5 transition-colors">
                        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-xs font-bold">
                          {user.name?.[0]}
                        </div>
                        <ChevronDown className="w-3 h-3 text-muted-foreground" />
                      </button>
                      <AnimatePresence>
                        {showMenu && (
                          <motion.div
                            initial={{ opacity: 0, y: 10, scale: 0.95 }}
                            animate={{ opacity: 1, y: 0, scale: 1 }}
                            exit={{ opacity: 0, y: 10, scale: 0.95 }}
                            className="absolute right-0 top-full mt-2 w-56 glass-card rounded-xl border border-white/10 shadow-2xl p-2"
                          >
                            <div className="px-3 py-2 border-b border-white/5 mb-1">
                              <p className="text-sm font-medium">{user.name}</p>
                              <p className="text-xs text-muted-foreground">{user.email}</p>
                            </div>
                            <Link href="/dashboard" onClick={() => setShowMenu(false)} className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors">
                              <LayoutDashboard className="w-4 h-4" /> Dashboard
                            </Link>
                            <Link href="/world" onClick={() => setShowMenu(false)} className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors">
                              <Globe className="w-4 h-4" /> World
                            </Link>
                            <Link href="/meetings" onClick={() => setShowMenu(false)} className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors">
                              <Calendar className="w-4 h-4" /> Meetings
                            </Link>
                            <Link href="/messages" onClick={() => setShowMenu(false)} className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors">
                              <MessageSquare className="w-4 h-4" /> Messages
                            </Link>
                            <Link href={`/profile/${user.id}`} onClick={() => setShowMenu(false)} className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors">
                              <Bell className="w-4 h-4" /> Profile
                            </Link>
                            {user.role === "admin" && (
                              <Link href="/admin" onClick={() => setShowMenu(false)} className="flex items-center gap-2 px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors">
                                <Shield className="w-4 h-4" /> Admin
                              </Link>
                            )}
                            <button onClick={() => { logout(); setShowMenu(false); }} className="w-full flex items-center gap-2 px-3 py-2 text-sm rounded-lg hover:bg-white/5 text-red-400 transition-colors">
                              <LogOut className="w-4 h-4" /> Sign Out
                            </button>
                          </motion.div>
                        )}
                      </AnimatePresence>
                    </div>
                  </>
                ) : (
                  <>
                    <Link href="/auth/login"><Button variant="ghost" size="sm">Sign In</Button></Link>
                    <Link href="/auth/register"><Button variant="glow" size="sm">Get Started Free</Button></Link>
                  </>
                )}
              </>
            )}
          </div>

          <button className="md:hidden p-2 rounded-lg hover:bg-white/10" onClick={() => setMobileOpen(!mobileOpen)}>
            {mobileOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
          </button>
        </div>
      </div>

      <AnimatePresence>
        {mobileOpen && (
          <motion.div initial={{ opacity: 0, height: 0 }} animate={{ opacity: 1, height: "auto" }} exit={{ opacity: 0, height: 0 }} className="md:hidden bg-background/95 backdrop-blur-xl border-t border-white/5">
            <div className="px-4 py-4 space-y-2">
              {navLinks.map(l => (
                <Link key={l.href} href={l.href} className="block px-4 py-3 text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg" onClick={() => setMobileOpen(false)}>
                  {l.label}
                </Link>
              ))}
              {user ? (
                <>
                  <Link href="/messages" className="block px-4 py-3 text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg" onClick={() => setMobileOpen(false)}>Messages {unread > 0 && `(${unread})`}</Link>
                  <Link href="/meetings" className="block px-4 py-3 text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg" onClick={() => setMobileOpen(false)}>Meetings</Link>
                  <Link href="/orders" className="block px-4 py-3 text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg" onClick={() => setMobileOpen(false)}>Orders</Link>
                  {user.role === "admin" && (
                    <Link href="/admin" className="block px-4 py-3 text-sm font-medium text-muted-foreground hover:text-foreground hover:bg-white/5 rounded-lg" onClick={() => setMobileOpen(false)}>Admin</Link>
                  )}
                  <button onClick={() => { logout(); setMobileOpen(false); }} className="w-full text-left px-4 py-3 text-sm font-medium text-red-400 hover:bg-white/5 rounded-lg">Sign Out</button>
                </>
              ) : (
                <div className="pt-2 space-y-2">
                  <Link href="/auth/login"><Button variant="outline" className="w-full border-white/10">Sign In</Button></Link>
                  <Link href="/auth/register"><Button variant="glow" className="w-full">Get Started Free</Button></Link>
                </div>
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </motion.header>
  );
}
