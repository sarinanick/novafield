"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { Search, User, Globe, Calendar, MessageSquare, ArrowRight, X } from "lucide-react";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { useShortcuts } from "@/lib/shortcuts";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

interface SearchResult {
  id: string;
  type: "user" | "zone" | "meeting";
  title: string;
  subtitle?: string;
  action: () => void;
}

const ZONE_RESULTS: SearchResult[] = [
  { id: "work", type: "zone", title: "Work Area", subtitle: "Productivity zone", action: () => {} },
  { id: "social", type: "zone", title: "Social Lounge", subtitle: "Chat & hang out", action: () => {} },
  { id: "meeting", type: "zone", title: "Meeting Room", subtitle: "Group discussions", action: () => {} },
  { id: "lounge", type: "zone", title: "Chill Zone", subtitle: "Relax & unwind", action: () => {} },
];

export default function SearchModal() {
  const { searchOpen, setSearchOpen } = useShortcuts();
  const { user } = useAuth();
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const [recentSearches, setRecentSearches] = useState<string[]>([]);
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (searchOpen) {
      setQuery("");
      setResults([]);
      setSelectedIndex(0);
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  }, [searchOpen]);

  useEffect(() => {
    const saved = localStorage.getItem("novafield-recent-searches");
    if (saved) {
      try {
        setRecentSearches(JSON.parse(saved));
      } catch {}
    }
  }, []);

  const saveRecentSearch = useCallback((q: string) => {
    if (!q.trim()) return;
    setRecentSearches((prev) => {
      const updated = [q, ...prev.filter((s) => s !== q)].slice(0, 5);
      localStorage.setItem("novafield-recent-searches", JSON.stringify(updated));
      return updated;
    });
  }, []);

  const search = useCallback(async (q: string) => {
    if (!q.trim()) {
      setResults([]);
      return;
    }

    setLoading(true);
    const lower = q.toLowerCase();
    const found: SearchResult[] = [];

    const zoneMatches = ZONE_RESULTS.filter(
      (z) => z.title.toLowerCase().includes(lower) || z.subtitle?.toLowerCase().includes(lower)
    ).map((z) => ({
      ...z,
      action: () => {
        window.dispatchEvent(new CustomEvent("shortcut:jump-zone", { detail: z.id }));
        setSearchOpen(false);
      },
    }));
    found.push(...zoneMatches);

    try {
      const gigs = await api.getGigs({ search: q }).catch(() => ({ gigs: [] }));
      if (gigs.gigs) {
        gigs.gigs.slice(0, 3).forEach((gig: any) => {
          found.push({
            id: gig.id,
            type: "meeting",
            title: gig.title,
            subtitle: `$${gig.price} · ${gig.ordersCount || 0} orders`,
            action: () => {
              router.push(`/gig/${gig.id}`);
              setSearchOpen(false);
            },
          });
        });
      }
    } catch {}

    try {
      const freelancers = await api.getFreelancers().catch(() => []);
      const userMatches = Array.isArray(freelancers)
        ? freelancers
            .filter((f: any) => f.name?.toLowerCase().includes(lower) || f.skills?.some((s: string) => s.toLowerCase().includes(lower)))
            .slice(0, 3)
        : [];
      userMatches.forEach((u: any) => {
        found.push({
          id: u.id,
          type: "user",
          title: u.name,
          subtitle: u.skills?.join(", ") || u.role,
          action: () => {
            router.push(`/profile/${u.id}`);
            setSearchOpen(false);
          },
        });
      });
    } catch {}

    try {
      const meetings = await api.getMeetings().catch(() => []);
      const meetingMatches = Array.isArray(meetings)
        ? meetings
            .filter((m: any) => m.title?.toLowerCase().includes(lower) || m.description?.toLowerCase().includes(lower))
            .slice(0, 3)
        : [];
      meetingMatches.forEach((m: any) => {
        found.push({
          id: m.id,
          type: "meeting",
          title: m.title,
          subtitle: new Date(m.scheduledAt).toLocaleString(),
          action: () => {
            router.push(`/meetings`);
            setSearchOpen(false);
          },
        });
      });
    } catch {}

    setResults(found);
    setSelectedIndex(0);
    setLoading(false);
  }, [router, setSearchOpen]);

  useEffect(() => {
    const timer = setTimeout(() => search(query), 200);
    return () => clearTimeout(timer);
  }, [query, search]);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    const allItems = results.length > 0 ? results : recentSearches.map((s) => ({ id: s, type: "recent" as const, title: s }));

    if (e.key === "ArrowDown") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev + 1) % allItems.length);
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setSelectedIndex((prev) => (prev - 1 + allItems.length) % allItems.length);
    } else if (e.key === "Enter") {
      e.preventDefault();
      if (results.length > 0 && results[selectedIndex]) {
        saveRecentSearch(query);
        results[selectedIndex].action();
      } else if (recentSearches.length > 0 && selectedIndex < recentSearches.length) {
        setQuery(recentSearches[selectedIndex]);
      }
    }
  };

  const groupedResults = results.reduce(
    (acc, r) => {
      const typeLabel = r.type === "user" ? "Users" : r.type === "zone" ? "Zones" : "Gigs & Meetings";
      if (!acc[typeLabel]) acc[typeLabel] = [];
      acc[typeLabel].push(r);
      return acc;
    },
    {} as Record<string, SearchResult[]>
  );

  const typeIcons = {
    user: User,
    zone: Globe,
    meeting: Calendar,
  };

  return (
    <Dialog open={searchOpen} onOpenChange={setSearchOpen}>
      <DialogContent className="max-w-xl p-0 gap-0 overflow-hidden glass-card border-white/10">
        <div className="flex items-center gap-3 px-4 py-3 border-b border-white/5">
          <Search className="w-5 h-5 text-muted-foreground shrink-0" />
          <Input
            ref={inputRef}
            placeholder="Search users, zones, gigs, meetings..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            className="border-0 bg-transparent focus-visible:ring-0 focus-visible:ring-offset-0 h-auto text-base"
          />
          <kbd className="hidden sm:inline-flex items-center gap-1 px-2 py-0.5 text-[10px] text-muted-foreground bg-muted rounded border border-white/5">
            ESC
          </kbd>
        </div>

        <div className="max-h-[400px] overflow-y-auto p-2">
          {loading && (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin w-6 h-6 border-2 border-primary border-t-transparent rounded-full" />
            </div>
          )}

          {!loading && query && results.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              <p className="text-sm">No results found for &quot;{query}&quot;</p>
            </div>
          )}

          {!loading && !query && recentSearches.length > 0 && (
            <div className="mb-2">
              <p className="text-xs font-medium text-muted-foreground px-2 py-1">Recent</p>
              {recentSearches.map((s, i) => (
                <button
                  key={s}
                  onClick={() => {
                    setQuery(s);
                    saveRecentSearch(s);
                  }}
                  className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                    i === selectedIndex ? "bg-white/10" : "hover:bg-white/5"
                  }`}
                >
                  <Search className="w-4 h-4 text-muted-foreground" />
                  <span>{s}</span>
                  <ArrowRight className="w-3 h-3 ml-auto text-muted-foreground" />
                </button>
              ))}
            </div>
          )}

          {!loading && Object.entries(groupedResults).map(([type, items]) => (
            <div key={type} className="mb-2">
              <p className="text-xs font-medium text-muted-foreground px-2 py-1">{type}</p>
              {items.map((result) => {
                const Icon = typeIcons[result.type];
                const idx = results.indexOf(result);
                return (
                  <button
                    key={result.id}
                    onClick={() => {
                      saveRecentSearch(query);
                      result.action();
                    }}
                    className={`w-full flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${
                      idx === selectedIndex ? "bg-white/10" : "hover:bg-white/5"
                    }`}
                  >
                    <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary/20 to-blue-500/20 flex items-center justify-center shrink-0">
                      <Icon className="w-4 h-4 text-primary" />
                    </div>
                    <div className="flex-1 text-left min-w-0">
                      <p className="font-medium truncate">{result.title}</p>
                      {result.subtitle && (
                        <p className="text-xs text-muted-foreground truncate">{result.subtitle}</p>
                      )}
                    </div>
                    <ArrowRight className="w-3 h-3 text-muted-foreground shrink-0" />
                  </button>
                );
              })}
            </div>
          ))}

          {!loading && !query && recentSearches.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              <Search className="w-8 h-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">Type to search across NovaField</p>
              <p className="text-xs mt-1">Search users, zones, gigs, and meetings</p>
            </div>
          )}
        </div>

        <div className="flex items-center justify-between px-4 py-2 border-t border-white/5 text-[10px] text-muted-foreground">
          <div className="flex items-center gap-3">
            <span className="flex items-center gap-1">
              <kbd className="px-1 py-0.5 bg-muted rounded border border-white/5">↑↓</kbd>
              navigate
            </span>
            <span className="flex items-center gap-1">
              <kbd className="px-1 py-0.5 bg-muted rounded border border-white/5">↵</kbd>
              select
            </span>
          </div>
          <span className="flex items-center gap-1">
            <kbd className="px-1 py-0.5 bg-muted rounded border border-white/5">esc</kbd>
            close
          </span>
        </div>
      </DialogContent>
    </Dialog>
  );
}