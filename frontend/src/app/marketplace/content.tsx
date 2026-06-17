"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { motion } from "framer-motion";
import { Search, SlidersHorizontal, Star, Clock } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";

export default function MarketplaceContent() {
  const searchParams = useSearchParams();
  const [gigs, setGigs] = useState<any[]>([]);
  const [categories, setCategories] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState(searchParams.get("q") || "");
  const [category, setCategory] = useState(searchParams.get("category") || "");
  const [sortBy, setSortBy] = useState("newest");

  useEffect(() => { api.getCategories().then(setCategories).catch(() => {}); }, []);

  const loadGigs = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = { page: String(page), limit: "12", sort: sortBy };
      if (query) params.q = query;
      if (category) params.category = category;
      const res = await api.getGigs(params);
      setGigs(res.gigs || []);
      setTotal(res.total || 0);
    } catch { setGigs([]); }
    setLoading(false);
  }, [page, query, category, sortBy]);

  useEffect(() => { loadGigs(); }, [loadGigs]);

  const gradients: Record<string, string> = {
    "ai-video": "from-purple-600 to-blue-600", "ai-image": "from-blue-600 to-cyan-600",
    "ai-audio": "from-green-600 to-emerald-600", "ai-animation": "from-pink-600 to-rose-600",
    "ai-chatbots": "from-indigo-600 to-violet-600", "ai-webdev": "from-orange-600 to-amber-600",
    "ai-writing": "from-teal-600 to-cyan-600", "ai-data": "from-red-600 to-orange-600",
    "ai-design": "from-violet-600 to-purple-600", "ai-marketing": "from-yellow-600 to-amber-600",
  };

  return (
    <div className="min-h-screen pt-20">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <h1 className="text-3xl md:text-4xl font-bold mb-2">Find AI <span className="text-gradient">Freelancers</span></h1>
          <p className="text-muted-foreground mb-6">{total} services available</p>
          <form onSubmit={(e) => { e.preventDefault(); setPage(1); loadGigs(); }} className="flex gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground" />
              <Input placeholder="Search AI services, tools, freelancers..." value={query} onChange={e => setQuery(e.target.value)} className="pl-12 h-12 text-base glass-card border-white/10" />
            </div>
            <Button type="submit" variant="glow" size="lg">Search</Button>
          </form>
        </motion.div>

        <div className="flex gap-8">
          <div className="hidden lg:block w-64 shrink-0">
            <div className="glass-card rounded-2xl p-6 sticky top-24">
              <h3 className="font-semibold mb-4">Categories</h3>
              <div className="space-y-1">
                <button onClick={() => { setCategory(""); setPage(1); }} className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${!category ? "bg-primary/10 text-primary" : "text-muted-foreground hover:text-foreground hover:bg-white/5"}`}>All Categories</button>
                {categories.map((cat: any) => (
                  <button key={cat.id} onClick={() => { setCategory(cat.slug); setPage(1); }} className={`w-full text-left px-3 py-2 rounded-lg text-sm transition-colors ${category === cat.slug ? "bg-primary/10 text-primary" : "text-muted-foreground hover:text-foreground hover:bg-white/5"}`}>{cat.name}</button>
                ))}
              </div>
              <div className="mt-6 pt-6 border-t border-white/5">
                <h3 className="font-semibold mb-4">Sort By</h3>
                {["newest", "popular", "rating", "price-low", "price-high"].map(s => (
                  <button key={s} onClick={() => { setSortBy(s); setPage(1); }} className={`w-full text-left px-3 py-2 rounded-lg text-sm capitalize transition-colors ${sortBy === s ? "bg-primary/10 text-primary" : "text-muted-foreground hover:text-foreground hover:bg-white/5"}`}>{s.replace("-", " ")}</button>
                ))}
              </div>
            </div>
          </div>

          <div className="flex-1">
            {loading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-5">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="glass-card rounded-2xl overflow-hidden animate-pulse">
                    <div className="aspect-video bg-white/5" />
                    <div className="p-5 space-y-3"><div className="h-4 bg-white/5 rounded w-3/4" /><div className="h-3 bg-white/5 rounded w-1/2" /></div>
                  </div>
                ))}
              </div>
            ) : gigs.length === 0 ? (
              <div className="text-center py-20"><p className="text-xl text-muted-foreground">No gigs found</p></div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-5">
                {gigs.map((gig: any) => (
                  <Link key={gig.id} href={`/gig/${gig.id}`}>
                    <motion.div whileHover={{ y: -6 }} className="glass-card rounded-2xl overflow-hidden cursor-pointer group h-full">
                      <div className="aspect-video relative overflow-hidden">
                        <div className={`absolute inset-0 bg-gradient-to-br ${gradients[gig.category] || "from-gray-600 to-slate-600"} opacity-60`} />
                        <div className="absolute inset-0 flex items-center justify-center text-lg font-bold opacity-80">{gig.aiTools?.[0] || gig.category}</div>
                      </div>
                      <div className="p-5">
                        <div className="flex items-center gap-3 mb-3">
                          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-xs font-bold">{gig.freelancer?.name?.[0] || "?"}</div>
                          <div>
                            <p className="text-sm font-medium">{gig.freelancer?.name}</p>
                            <p className="text-xs text-muted-foreground">{gig.freelancer?.isVerified ? "Verified Seller" : "New Seller"}</p>
                          </div>
                        </div>
                        <h3 className="font-semibold mb-2 line-clamp-2 group-hover:text-primary transition-colors">{gig.title}</h3>
                        <div className="flex items-center gap-4 text-sm">
                          <span className="flex items-center gap-1 text-amber-400"><Star className="w-3.5 h-3.5 fill-current" /> {gig.rating?.toFixed(1) || "New"}<span className="text-muted-foreground">({gig.reviewsCount || 0})</span></span>
                          <span className="flex items-center gap-1 text-muted-foreground"><Clock className="w-3.5 h-3.5" /> {gig.deliveryDays}d</span>
                        </div>
                        <div className="mt-3 pt-3 border-t border-white/5 flex justify-between items-center">
                          <span className="text-xs text-muted-foreground uppercase">{gig.ordersCount || 0} orders</span>
                          <span className="font-bold text-lg">${gig.price}</span>
                        </div>
                      </div>
                    </motion.div>
                  </Link>
                ))}
              </div>
            )}

            {total > 12 && (
              <div className="flex justify-center gap-2 mt-10">
                <Button variant="outline" size="sm" disabled={page === 1} onClick={() => setPage(p => p - 1)} className="border-white/10">Prev</Button>
                <span className="flex items-center px-4 text-sm text-muted-foreground">Page {page} of {Math.ceil(total / 12)}</span>
                <Button variant="outline" size="sm" disabled={page * 12 >= total} onClick={() => setPage(p => p + 1)} className="border-white/10">Next</Button>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
