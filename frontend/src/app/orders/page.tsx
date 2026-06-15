"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { motion } from "framer-motion";
import { Package, Clock, CheckCircle, AlertCircle, Star, MessageSquare, Send, FileUp } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function OrdersPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [orders, setOrders] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState("active");
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    if (!authLoading && !user) { router.push("/auth/login"); return; }
    loadOrders();
  }, [user, authLoading]);

  const loadOrders = async () => {
    setLoading(true);
    try {
      const o = await api.getOrders();
      setOrders(o);
    } catch {}
    setLoading(false);
  };

  const handleDeliver = async (orderId: string) => {
    setActionLoading(orderId);
    try {
      await api.deliverOrder(orderId, { file: "delivery.zip", notes: "Here is your delivery" });
      loadOrders();
    } catch (err: any) { alert(err.message); }
    setActionLoading(null);
  };

  const handleApprove = async (orderId: string) => {
    setActionLoading(orderId);
    try {
      await api.approveOrder(orderId);
      loadOrders();
    } catch (err: any) { alert(err.message); }
    setActionLoading(null);
  };

  const handleRevision = async (orderId: string) => {
    setActionLoading(orderId);
    try {
      await api.requestRevision(orderId, { message: "Please make revisions" });
      loadOrders();
    } catch (err: any) { alert(err.message); }
    setActionLoading(null);
  };

  if (authLoading || loading) return <div className="min-h-screen pt-20 flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" /></div>;

  const statusColors: Record<string, string> = {
    pending: "bg-yellow-500/10 text-yellow-400 border-yellow-500/20",
    active: "bg-blue-500/10 text-blue-400 border-blue-500/20",
    delivered: "bg-purple-500/10 text-purple-400 border-purple-500/20",
    revision: "bg-orange-500/10 text-orange-400 border-orange-500/20",
    completed: "bg-green-500/10 text-green-400 border-green-500/20",
  };

  const filtered = orders.filter(o => {
    if (activeTab === "active") return ["active", "pending", "revision"].includes(o.status);
    if (activeTab === "delivered") return o.status === "delivered";
    if (activeTab === "completed") return o.status === "completed";
    return true;
  });

  return (
    <div className="min-h-screen pt-20 pb-16">
      <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
          <h1 className="text-3xl font-bold mb-2">Orders</h1>
          <p className="text-muted-foreground mb-8">{orders.length} total orders</p>

          <Tabs value={activeTab} onValueChange={setActiveTab}>
            <TabsList className="glass-card border border-white/10 mb-6">
              <TabsTrigger value="active">Active ({orders.filter(o => ["active", "pending", "revision"].includes(o.status)).length})</TabsTrigger>
              <TabsTrigger value="delivered">Delivered ({orders.filter(o => o.status === "delivered").length})</TabsTrigger>
              <TabsTrigger value="completed">Completed ({orders.filter(o => o.status === "completed").length})</TabsTrigger>
              <TabsTrigger value="all">All</TabsTrigger>
            </TabsList>

            <TabsContent value={activeTab}>
              {filtered.length === 0 ? (
                <div className="text-center py-16">
                  <Package className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-lg text-muted-foreground">No orders in this category</p>
                  <Link href="/marketplace"><Button variant="outline" className="mt-4 border-white/10">Browse Marketplace</Button></Link>
                </div>
              ) : (
                <div className="space-y-4">
                  {filtered.map((order: any, i: number) => (
                    <motion.div key={order.id} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.05 }}>
                      <Card className="glass-card border-white/5">
                        <CardContent className="p-6">
                          <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                            <div className="flex items-start gap-4">
                              <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-primary/20 to-blue-500/20 flex items-center justify-center shrink-0">
                                <Package className="w-6 h-6 text-primary" />
                              </div>
                              <div>
                                <Link href={`/gig/${order.gigId}`} className="font-semibold hover:text-primary transition-colors">
                                  {order.gig?.title || "Order"}
                                </Link>
                                <p className="text-sm text-muted-foreground mt-1">
                                  {user?.role === "freelancer" ? `Client: ${order.buyer?.name}` : `Freelancer: ${order.seller?.name}`}
                                </p>
                                <p className="text-xs text-muted-foreground mt-1">Created: {order.createdAt?.split("T")[0]}</p>
                              </div>
                            </div>

                            <div className="flex items-center gap-3">
                              <span className={`px-3 py-1.5 rounded-full text-xs font-medium border ${statusColors[order.status] || ""}`}>
                                {order.status}
                              </span>
                              <span className="text-lg font-bold">${order.price}</span>
                            </div>
                          </div>

                          {order.status === "delivered" && user?.role === "client" && (
                            <div className="flex gap-3 mt-4 pt-4 border-t border-white/5">
                              <Button size="sm" variant="glow" onClick={() => handleApprove(order.id)} disabled={actionLoading === order.id}>
                                <CheckCircle className="w-4 h-4 mr-1" /> Approve & Pay
                              </Button>
                              <Button size="sm" variant="outline" className="border-white/10" onClick={() => handleRevision(order.id)} disabled={actionLoading === order.id}>
                                <AlertCircle className="w-4 h-4 mr-1" /> Request Revision
                              </Button>
                              <Link href={`/messages?user=${order.sellerId}`}>
                                <Button size="sm" variant="ghost"><MessageSquare className="w-4 h-4 mr-1" /> Message</Button>
                              </Link>
                            </div>
                          )}

                          {order.status === "active" && user?.role === "freelancer" && (
                            <div className="flex gap-3 mt-4 pt-4 border-t border-white/5">
                              <Button size="sm" variant="glow" onClick={() => handleDeliver(order.id)} disabled={actionLoading === order.id}>
                                <Send className="w-4 h-4 mr-1" /> Deliver Order
                              </Button>
                              <Link href={`/messages?user=${order.buyerId}`}>
                                <Button size="sm" variant="ghost"><MessageSquare className="w-4 h-4 mr-1" /> Message</Button>
                              </Link>
                            </div>
                          )}

                          {order.status === "completed" && (
                            <div className="mt-4 pt-4 border-t border-white/5">
                              <Link href={`/orders`}>
                                <Button size="sm" variant="outline" className="border-white/10">
                                  <Star className="w-4 h-4 mr-1" /> Leave Review
                                </Button>
                              </Link>
                            </div>
                          )}
                        </CardContent>
                      </Card>
                    </motion.div>
                  ))}
                </div>
              )}
            </TabsContent>
          </Tabs>
        </motion.div>
      </div>
    </div>
  );
}
