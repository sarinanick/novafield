"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { Star, Clock, RefreshCcw, Heart, MessageSquare, Shield, ChevronRight, Check, ShoppingCart } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function GigDetailPage() {
  const params = useParams();
  const router = useRouter();
  const { user } = useAuth();
  const [gig, setGig] = useState<any>(null);
  const [reviews, setReviews] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedPackage, setSelectedPackage] = useState<any>(null);
  const [ordering, setOrdering] = useState(false);

  useEffect(() => {
    if (!params.id) return;
    Promise.all([
      api.getGig(params.id as string).catch(() => null),
      api.getGigReviews(params.id as string).catch(() => []),
    ]).then(([g, r]) => {
      setGig(g);
      setReviews(r);
      if (g?.packages?.length) setSelectedPackage(g.packages[0]);
      setLoading(false);
    });
  }, [params.id]);

  const handleOrder = async () => {
    if (!user) { router.push("/auth/login"); return; }
    if (!selectedPackage) return;
    setOrdering(true);
    try {
      const order = await api.createOrder({ gigId: gig.id, packageId: selectedPackage.id, requirements: "" });
      router.push(`/orders`);
    } catch (err: any) { alert(err.message); }
    setOrdering(false);
  };

  if (loading) return <div className="min-h-screen pt-20 flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" /></div>;
  if (!gig) return <div className="min-h-screen pt-20 text-center py-20"><p className="text-xl text-muted-foreground">Gig not found</p></div>;

  return (
    <div className="min-h-screen pt-20 pb-16">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Breadcrumb */}
        <div className="flex items-center gap-2 text-sm text-muted-foreground mb-6">
          <Link href="/marketplace" className="hover:text-foreground">Marketplace</Link>
          <ChevronRight className="w-3 h-3" />
          <Link href={`/marketplace?category=${gig.category}`} className="hover:text-foreground capitalize">{gig.category?.replace("ai-", "AI ")}</Link>
          <ChevronRight className="w-3 h-3" />
          <span className="text-foreground truncate">{gig.title}</span>
        </div>

        <div className="grid lg:grid-cols-3 gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-8">
            <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
              <h1 className="text-3xl font-bold mb-4">{gig.title}</h1>
              <div className="flex items-center gap-4 mb-6">
                <Link href={`/profile/${gig.freelancer?.id}`} className="flex items-center gap-3 hover:opacity-80 transition">
                  <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center font-bold">
                    {gig.freelancer?.name?.[0]}
                  </div>
                  <div>
                    <p className="font-medium">{gig.freelancer?.name}</p>
                    <p className="text-xs text-muted-foreground">{gig.freelancer?.isVerified ? "✓ Verified Seller" : "New Seller"}</p>
                  </div>
                </Link>
                <div className="flex items-center gap-1 text-amber-400">
                  <Star className="w-4 h-4 fill-current" />
                  <span className="font-medium">{gig.rating?.toFixed(1) || "New"}</span>
                  <span className="text-muted-foreground">({gig.reviewsCount || 0})</span>
                </div>
              </div>

              <div className="glass-card rounded-2xl overflow-hidden mb-8">
                <div className={`aspect-video bg-gradient-to-br ${getGradient(gig.category)} flex items-center justify-center text-4xl font-bold opacity-80`}>
                  {gig.title}
                </div>
              </div>

              <Tabs defaultValue="description">
                <TabsList className="glass-card border border-white/10">
                  <TabsTrigger value="description">Description</TabsTrigger>
                  <TabsTrigger value="about">About the Seller</TabsTrigger>
                  <TabsTrigger value="reviews">Reviews ({reviews.length})</TabsTrigger>
                </TabsList>

                <TabsContent value="description" className="mt-6">
                  <div className="prose prose-invert max-w-none">
                    <p className="text-muted-foreground leading-relaxed whitespace-pre-wrap">{gig.description}</p>
                    {gig.aiTools?.length > 0 && (
                      <div className="mt-6">
                        <h3 className="text-lg font-semibold mb-3">AI Tools Used</h3>
                        <div className="flex flex-wrap gap-2">
                          {gig.aiTools.map((tool: string) => (
                            <span key={tool} className="px-3 py-1 rounded-full glass-card text-sm">{tool}</span>
                          ))}
                        </div>
                      </div>
                    )}
                    {gig.tags?.length > 0 && (
                      <div className="mt-4">
                        <div className="flex flex-wrap gap-2">
                          {gig.tags.map((tag: string) => (
                            <span key={tag} className="px-3 py-1 rounded-full bg-primary/10 text-primary text-sm">#{tag}</span>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>
                </TabsContent>

                <TabsContent value="about" className="mt-6">
                  <Card className="glass-card border-white/5">
                    <CardContent className="p-6">
                      <div className="flex items-center gap-4 mb-4">
                        <div className="w-16 h-16 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-2xl font-bold">
                          {gig.freelancer?.name?.[0]}
                        </div>
                        <div>
                          <h3 className="text-xl font-semibold">{gig.freelancer?.name}</h3>
                          <p className="text-muted-foreground">{gig.freelancer?.isVerified ? "✓ Verified Seller" : "New Seller"}</p>
                        </div>
                      </div>
                      <div className="grid grid-cols-3 gap-4 mt-6">
                        <div className="text-center p-3 rounded-xl glass-card">
                          <div className="text-2xl font-bold">{gig.freelancer?.rating?.toFixed(1) || "New"}</div>
                          <div className="text-xs text-muted-foreground">Rating</div>
                        </div>
                        <div className="text-center p-3 rounded-xl glass-card">
                          <div className="text-2xl font-bold">{gig.freelancer?.reviewsCount || 0}</div>
                          <div className="text-xs text-muted-foreground">Reviews</div>
                        </div>
                        <div className="text-center p-3 rounded-xl glass-card">
                          <div className="text-2xl font-bold">{gig.ordersCount || 0}</div>
                          <div className="text-xs text-muted-foreground">Orders</div>
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </TabsContent>

                <TabsContent value="reviews" className="mt-6">
                  <div className="space-y-4">
                    {reviews.length === 0 ? (
                      <p className="text-muted-foreground text-center py-8">No reviews yet</p>
                    ) : reviews.map((r: any) => (
                      <Card key={r.id} className="glass-card border-white/5">
                        <CardContent className="p-5">
                          <div className="flex items-center gap-3 mb-3">
                            <div className="w-8 h-8 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-xs font-bold">
                              {r.reviewer?.name?.[0]}
                            </div>
                            <div>
                              <p className="text-sm font-medium">{r.reviewer?.name}</p>
                              <div className="flex items-center gap-1">
                                {[...Array(5)].map((_, i) => (
                                  <Star key={i} className={`w-3 h-3 ${i < r.rating ? "text-amber-400 fill-current" : "text-muted-foreground"}`} />
                                ))}
                              </div>
                            </div>
                          </div>
                          <p className="text-sm text-muted-foreground">{r.comment}</p>
                        </CardContent>
                      </Card>
                    ))}
                  </div>
                </TabsContent>
              </Tabs>
            </motion.div>
          </div>

          {/* Sidebar - Packages */}
          <div className="lg:col-span-1">
            <div className="sticky top-24 space-y-4">
              {gig.packages?.map((pkg: any) => (
                <motion.div
                  key={pkg.id}
                  whileHover={{ scale: 1.01 }}
                  className={`glass-card rounded-2xl p-6 cursor-pointer transition-all ${selectedPackage?.id === pkg.id ? "border-2 border-primary shadow-xl shadow-primary/10" : "border border-white/5"}`}
                  onClick={() => setSelectedPackage(pkg)}
                >
                  <div className="flex justify-between items-start mb-3">
                    <h3 className="font-semibold">{pkg.name}</h3>
                    <span className="text-2xl font-bold">${pkg.price}</span>
                  </div>
                  <p className="text-sm text-muted-foreground mb-4">{pkg.description}</p>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground mb-4">
                    <span className="flex items-center gap-1"><Clock className="w-3.5 h-3.5" /> {pkg.deliveryDays} day delivery</span>
                    <span className="flex items-center gap-1"><RefreshCcw className="w-3.5 h-3.5" /> {pkg.revisions} revisions</span>
                  </div>
                  {pkg.features?.length > 0 && (
                    <ul className="space-y-2">
                      {pkg.features.map((f: string) => (
                        <li key={f} className="flex items-center gap-2 text-sm text-muted-foreground">
                          <Check className="w-3.5 h-3.5 text-primary shrink-0" />
                          {f}
                        </li>
                      ))}
                    </ul>
                  )}
                </motion.div>
              ))}

              <Button variant="glow" className="w-full" size="lg" onClick={handleOrder} disabled={ordering || !selectedPackage}>
                <ShoppingCart className="w-4 h-4 mr-2" />
                {ordering ? "Placing Order..." : `Continue ($${selectedPackage?.price || 0})`}
              </Button>

              {user && user.id !== gig.freelancer?.id && (
                <Link href={`/messages?user=${gig.freelancer?.id}`}>
                  <Button variant="outline" className="w-full border-white/10" size="lg">
                    <MessageSquare className="w-4 h-4 mr-2" />
                    Contact Seller
                  </Button>
                </Link>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function getGradient(cat: string): string {
  const map: Record<string, string> = {
    "ai-video": "from-purple-600 to-blue-600", "ai-image": "from-blue-600 to-cyan-600",
    "ai-audio": "from-green-600 to-emerald-600", "ai-animation": "from-pink-600 to-rose-600",
  };
  return map[cat] || "from-gray-600 to-slate-600";
}
