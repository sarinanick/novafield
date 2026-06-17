"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { BarChart3, DollarSign, ShoppingCart, Star, Eye, TrendingUp, Package, Clock, CheckCircle, AlertCircle, MessageSquare, Plus } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function DashboardPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [stats, setStats] = useState<any>(null);
  const [orders, setOrders] = useState<any[]>([]);
  const [myGigs, setMyGigs] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!authLoading && !user) { router.push("/auth/login"); return; }
    if (!user) return;
    Promise.all([
      api.getDashboard().catch(() => null),
      api.getOrders().catch(() => []),
      user.role === "freelancer" ? api.getMyGigs().catch(() => []) : Promise.resolve([]),
    ]).then(([s, o, g]) => {
      setStats(s);
      setOrders(o);
      setMyGigs(g);
      setLoading(false);
    });
  }, [user, authLoading]);

  if (authLoading || loading) return <div className="min-h-screen pt-20 flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" /></div>;
  if (!user) return null;

  const statusColors: Record<string, string> = {
    pending: "bg-yellow-500/10 text-yellow-400",
    active: "bg-blue-500/10 text-blue-400",
    delivered: "bg-purple-500/10 text-purple-400",
    revision: "bg-orange-500/10 text-orange-400",
    completed: "bg-green-500/10 text-green-400",
  };

  return (
    <div className="min-h-screen pt-20 pb-16">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
          <div className="flex justify-between items-center mb-8">
            <div>
              <h1 className="text-3xl font-bold">Welcome back, {user.name}</h1>
              <p className="text-muted-foreground mt-1">Here&apos;s your {user.role === "freelancer" ? "seller" : "buyer"} dashboard</p>
            </div>
            {user.role === "freelancer" && (
              <Link href="/create-gig">
                <Button variant="glow"><Plus className="w-4 h-4 mr-2" /> Create New Gig</Button>
              </Link>
            )}
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
            {[
              { label: "Total Orders", value: stats?.totalOrders || 0, icon: ShoppingCart, color: "from-blue-500 to-cyan-500" },
              { label: "Active Orders", value: stats?.activeOrders || 0, icon: Clock, color: "from-orange-500 to-amber-500" },
              { label: user.role === "freelancer" ? "Total Earnings" : "Total Spent", value: `$${(user.role === "freelancer" ? stats?.totalEarnings : stats?.totalSpent)?.toFixed(0) || 0}`, icon: DollarSign, color: "from-green-500 to-emerald-500" },
              { label: "Avg Rating", value: stats?.avgRating?.toFixed(1) || "New", icon: Star, color: "from-amber-500 to-yellow-500" },
            ].map((stat, i) => (
              <motion.div key={stat.label} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.1 }}>
                <Card className="glass-card border-white/5">
                  <CardContent className="p-5">
                    <div className="flex items-center gap-3">
                      <div className={`w-10 h-10 bg-gradient-to-br ${stat.color} rounded-xl flex items-center justify-center`}>
                        <stat.icon className="w-5 h-5 text-white" />
                      </div>
                      <div>
                        <p className="text-2xl font-bold">{stat.value}</p>
                        <p className="text-xs text-muted-foreground">{stat.label}</p>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>

          <Tabs defaultValue="orders">
            <TabsList className="glass-card border border-white/10 mb-6">
              <TabsTrigger value="orders">Orders</TabsTrigger>
              {user.role === "freelancer" && <TabsTrigger value="gigs">My Gigs</TabsTrigger>}
              <TabsTrigger value="stats">Statistics</TabsTrigger>
            </TabsList>

            <TabsContent value="orders">
              <Card className="glass-card border-white/5">
                <CardContent className="p-6">
                  <h3 className="font-semibold mb-4">Recent Orders</h3>
                  {orders.length === 0 ? (
                    <p className="text-muted-foreground text-center py-8">No orders yet</p>
                  ) : (
                    <div className="space-y-3">
                      {orders.map((o: any) => (
                        <Link key={o.id} href={`/orders`}>
                          <motion.div whileHover={{ x: 4 }} className="flex items-center justify-between p-4 rounded-xl hover:bg-white/5 cursor-pointer transition-colors">
                            <div className="flex items-center gap-4">
                              <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-primary/20 to-blue-500/20 flex items-center justify-center">
                                <Package className="w-5 h-5 text-primary" />
                              </div>
                              <div>
                                <p className="font-medium text-sm">{o.gig?.title || "Order"}</p>
                                <p className="text-xs text-muted-foreground">
                                  {user.role === "freelancer" ? `From ${o.buyer?.name}` : `To ${o.seller?.name}`}
                                </p>
                              </div>
                            </div>
                            <div className="flex items-center gap-4">
                              <span className={`px-2.5 py-1 rounded-full text-xs font-medium ${statusColors[o.status] || ""}`}>
                                {o.status}
                              </span>
                              <span className="font-semibold">${o.price}</span>
                            </div>
                          </motion.div>
                        </Link>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            {user.role === "freelancer" && (
              <TabsContent value="gigs">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {myGigs.map((gig: any) => (
                    <Link key={gig.id} href={`/gig/${gig.id}`}>
                      <Card className="glass-card border-white/5 cursor-pointer hover:border-primary/30 transition-colors h-full">
                        <CardContent className="p-5">
                          <h3 className="font-semibold mb-2">{gig.title}</h3>
                          <div className="flex justify-between text-sm text-muted-foreground">
                            <span>{gig.ordersCount || 0} orders</span>
                            <span>{gig.views || 0} views</span>
                          </div>
                          <div className="flex justify-between items-center mt-3">
                            <span className={`text-xs px-2 py-1 rounded-full ${gig.status === "active" ? "bg-green-500/10 text-green-400" : "bg-gray-500/10 text-gray-400"}`}>
                              {gig.status}
                            </span>
                            <span className="font-bold">${gig.price}</span>
                          </div>
                        </CardContent>
                      </Card>
                    </Link>
                  ))}
                  <Link href="/create-gig">
                    <Card className="glass-card border-dashed border-white/10 cursor-pointer hover:border-primary/30 transition-colors h-full flex items-center justify-center min-h-[200px]">
                      <div className="text-center">
                        <Plus className="w-8 h-8 text-muted-foreground mx-auto mb-2" />
                        <p className="text-sm text-muted-foreground">Create New Gig</p>
                      </div>
                    </Card>
                  </Link>
                </div>
              </TabsContent>
            )}

            <TabsContent value="stats">
              <Card className="glass-card border-white/5">
                <CardContent className="p-6">
                  <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
                    <div className="text-center">
                      <div className="text-3xl font-bold">{stats?.completedOrders || 0}</div>
                      <div className="text-sm text-muted-foreground">Completed</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold">{stats?.totalGigs || 0}</div>
                      <div className="text-sm text-muted-foreground">Active Gigs</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold">{stats?.totalViews || 0}</div>
                      <div className="text-sm text-muted-foreground">Total Views</div>
                    </div>
                    <div className="text-center">
                      <div className="text-3xl font-bold">{((stats?.conversionRate || 0) * 100).toFixed(0)}%</div>
                      <div className="text-sm text-muted-foreground">Conversion</div>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </motion.div>
      </div>
    </div>
  );
}
