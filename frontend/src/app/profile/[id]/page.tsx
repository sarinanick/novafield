"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { Star, MapPin, Calendar, Shield, ExternalLink } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";

export default function ProfilePage() {
  const params = useParams();
  const [profile, setProfile] = useState<any>(null);
  const [gigs, setGigs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!params.id) return;
    Promise.all([
      api.getProfile(params.id as string).catch(() => null),
      api.getGigs({ q: "", category: "" }).then(res => (res.gigs || []).filter((g: any) => g.freelancerId === params.id)).catch(() => []),
    ]).then(([p, g]) => {
      setProfile(p);
      setGigs(g);
      setLoading(false);
    });
  }, [params.id]);

  if (loading) return <div className="min-h-screen pt-20 flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" /></div>;
  if (!profile) return <div className="min-h-screen pt-20 text-center py-20"><p className="text-xl text-muted-foreground">Profile not found</p></div>;

  return (
    <div className="min-h-screen pt-20 pb-16">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
          <Card className="glass-card border-white/5 mb-8">
            <CardContent className="p-8">
              <div className="flex flex-col md:flex-row items-start gap-6">
                <div className="w-24 h-24 rounded-2xl bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-4xl font-bold shrink-0 shadow-xl shadow-primary/20">
                  {profile.name?.[0]}
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-3 mb-2">
                    <h1 className="text-3xl font-bold">{profile.name}</h1>
                    {profile.isVerified && <Shield className="w-6 h-6 text-primary" />}
                  </div>
                  <p className="text-muted-foreground mb-4">{profile.bio || "No bio yet"}</p>
                  <div className="flex flex-wrap gap-4 text-sm text-muted-foreground">
                    {profile.location && <span className="flex items-center gap-1"><MapPin className="w-4 h-4" />{profile.location}</span>}
                    <span className="flex items-center gap-1"><Calendar className="w-4 h-4" />Joined {profile.joinedAt?.split("T")[0] || "Recently"}</span>
                    {profile.hourlyRate > 0 && <span className="font-semibold text-foreground">${profile.hourlyRate}/hr</span>}
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-8">
                <div className="text-center p-4 rounded-xl glass-card">
                  <div className="text-2xl font-bold text-amber-400">{profile.rating?.toFixed(1) || "New"}</div>
                  <div className="text-xs text-muted-foreground">Rating</div>
                </div>
                <div className="text-center p-4 rounded-xl glass-card">
                  <div className="text-2xl font-bold">{profile.reviewsCount || 0}</div>
                  <div className="text-xs text-muted-foreground">Reviews</div>
                </div>
                <div className="text-center p-4 rounded-xl glass-card">
                  <div className="text-2xl font-bold">{gigs.length}</div>
                  <div className="text-xs text-muted-foreground">Active Gigs</div>
                </div>
                <div className="text-center p-4 rounded-xl glass-card">
                  <div className="text-2xl font-bold">{profile.isVerified ? "✓" : "—"}</div>
                  <div className="text-xs text-muted-foreground">Verified</div>
                </div>
              </div>

              {profile.skills?.length > 0 && (
                <div className="mt-6">
                  <h3 className="text-sm font-medium mb-3">Skills & AI Tools</h3>
                  <div className="flex flex-wrap gap-2">
                    {profile.skills.map((s: string) => (
                      <span key={s} className="px-3 py-1 rounded-full glass-card text-sm">{s}</span>
                    ))}
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {gigs.length > 0 && (
            <div>
              <h2 className="text-2xl font-bold mb-6">Active Gigs</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-5">
                {gigs.map((gig: any) => (
                  <Link key={gig.id} href={`/gig/${gig.id}`}>
                    <motion.div whileHover={{ y: -6 }} className="glass-card rounded-2xl overflow-hidden cursor-pointer group h-full">
                      <div className="aspect-video bg-gradient-to-br from-primary/20 to-blue-500/20 flex items-center justify-center text-lg font-bold opacity-80">
                        {gig.title}
                      </div>
                      <div className="p-5">
                        <h3 className="font-semibold mb-2 group-hover:text-primary transition-colors">{gig.title}</h3>
                        <div className="flex justify-between items-center">
                          <span className="flex items-center gap-1 text-amber-400 text-sm">
                            <Star className="w-3.5 h-3.5 fill-current" /> {gig.rating?.toFixed(1) || "New"}
                          </span>
                          <span className="font-bold">${gig.price}</span>
                        </div>
                      </div>
                    </motion.div>
                  </Link>
                ))}
              </div>
            </div>
          )}
        </motion.div>
      </div>
    </div>
  );
}
