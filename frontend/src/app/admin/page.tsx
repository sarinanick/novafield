"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { motion, AnimatePresence } from "framer-motion";
import { Shield, Users, Settings, UserPlus, Search, Trash2, ArrowDown, ChevronDown, Eye, EyeOff, Mail, Bell, Layout, Crown, AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogClose } from "@/components/ui/dialog";
import { useAuth } from "@/lib/auth-context";
import { api } from "@/lib/api";

export default function AdminPage() {
  const { user, loading } = useAuth();
  const router = useRouter();
  const [members, setMembers] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState("");
  const [roleFilter, setRoleFilter] = useState("all");
  const [loadingMembers, setLoadingMembers] = useState(true);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("client");
  const [inviting, setInviting] = useState(false);
  const [actionMenuId, setActionMenuId] = useState<string | null>(null);
  const [confirmAction, setConfirmAction] = useState<{ type: string; id: string; name: string } | null>(null);
  const [processing, setProcessing] = useState(false);

  const [settings, setSettings] = useState({
    name: "NovaField AI",
    defaultLayout: "open",
    maxMembers: 100,
    guestAccess: true,
    emailNotify: true,
  });
  const [savingSettings, setSavingSettings] = useState(false);

  useEffect(() => {
    if (!loading && (!user || user.role !== "admin")) {
      router.push("/dashboard");
    }
  }, [user, loading, router]);

  useEffect(() => {
    if (user?.role === "admin") {
      loadMembers();
      loadSettings();
    }
  }, [user]);

  const loadMembers = async () => {
    try {
      setLoadingMembers(true);
      const res = await api.getMembers();
      setMembers(res.members || []);
      setTotal(res.total || 0);
    } catch {
    } finally {
      setLoadingMembers(false);
    }
  };

  const loadSettings = async () => {
    try {
      const res = await api.getAdminSettings();
      setSettings(res);
    } catch {
    }
  };

  const handleInvite = async () => {
    if (!inviteEmail) return;
    try {
      setInviting(true);
      await api.inviteMember({ email: inviteEmail, role: inviteRole });
      setShowInviteModal(false);
      setInviteEmail("");
      setInviteRole("client");
      loadMembers();
    } catch (err: any) {
      alert(err.message || "Failed to invite member");
    } finally {
      setInviting(false);
    }
  };

  const handleRoleChange = async (id: string, role: string) => {
    try {
      setProcessing(true);
      await api.changeMemberRole(id, role);
      setActionMenuId(null);
      loadMembers();
    } catch (err: any) {
      alert(err.message || "Failed to change role");
    } finally {
      setProcessing(false);
    }
  };

  const handleRemove = async () => {
    if (!confirmAction) return;
    try {
      setProcessing(true);
      await api.removeMember(confirmAction.id);
      setConfirmAction(null);
      loadMembers();
    } catch (err: any) {
      alert(err.message || "Failed to remove member");
    } finally {
      setProcessing(false);
    }
  };

  const handleDemote = async () => {
    if (!confirmAction) return;
    try {
      setProcessing(true);
      await api.demoteMember(confirmAction.id);
      setConfirmAction(null);
      loadMembers();
    } catch (err: any) {
      alert(err.message || "Failed to demote member");
    } finally {
      setProcessing(false);
    }
  };

  const handleSaveSettings = async () => {
    try {
      setSavingSettings(true);
      await api.updateAdminSettings(settings);
    } catch (err: any) {
      alert(err.message || "Failed to save settings");
    } finally {
      setSavingSettings(false);
    }
  };

  const filtered = members.filter((m) => {
    const matchSearch = !search || m.name?.toLowerCase().includes(search.toLowerCase()) || m.email?.toLowerCase().includes(search.toLowerCase());
    const matchRole = roleFilter === "all" || m.role === roleFilter;
    return matchSearch && matchRole;
  });

  const roleCounts = {
    all: members.length,
    admin: members.filter((m) => m.role === "admin").length,
    client: members.filter((m) => m.role === "client").length,
    freelancer: members.filter((m) => m.role === "freelancer").length,
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case "admin": return "bg-purple-500/10 text-purple-400 border-purple-500/20";
      case "freelancer": return "bg-blue-500/10 text-blue-400 border-blue-500/20";
      case "client": return "bg-emerald-500/10 text-emerald-400 border-emerald-500/20";
      default: return "bg-gray-500/10 text-gray-400 border-gray-500/20";
    }
  };

  if (loading || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" />
      </div>
    );
  }

  if (user.role !== "admin") {
    return null;
  }

  return (
    <div className="min-h-screen pt-20 pb-10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center">
              <Shield className="w-5 h-5 text-white" />
            </div>
            <h1 className="text-3xl font-bold">Admin Panel</h1>
          </div>
          <p className="text-muted-foreground">Manage members, roles, and workspace settings</p>
        </motion.div>

        <div className="grid grid-cols-1 sm:grid-cols-4 gap-4 mb-8">
          {[
            { label: "Total Members", value: total, icon: Users, color: "from-blue-500 to-cyan-500" },
            { label: "Admins", value: roleCounts.admin, icon: Crown, color: "from-purple-500 to-pink-500" },
            { label: "Freelancers", value: roleCounts.freelancer, icon: Users, color: "from-emerald-500 to-teal-500" },
            { label: "Clients", value: roleCounts.client, icon: Users, color: "from-orange-500 to-amber-500" },
          ].map((stat) => (
            <motion.div key={stat.label} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
              <Card>
                <CardContent className="p-4">
                  <div className="flex items-center gap-3">
                    <div className={`w-10 h-10 rounded-lg bg-gradient-to-br ${stat.color} flex items-center justify-center`}>
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

        <Tabs defaultValue="members">
          <TabsList className="mb-6">
            <TabsTrigger value="members" className="gap-2">
              <Users className="w-4 h-4" /> Members
            </TabsTrigger>
            <TabsTrigger value="settings" className="gap-2">
              <Settings className="w-4 h-4" /> Settings
            </TabsTrigger>
          </TabsList>

          <TabsContent value="members">
            <Card>
              <CardHeader>
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <CardTitle className="text-lg">Member Management</CardTitle>
                  <Button onClick={() => setShowInviteModal(true)} size="sm" className="gap-2">
                    <UserPlus className="w-4 h-4" /> Invite Member
                  </Button>
                </div>
                <div className="flex flex-col sm:flex-row gap-3 mt-4">
                  <div className="relative flex-1">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                    <Input
                      placeholder="Search by name or email..."
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      className="pl-9"
                    />
                  </div>
                  <select
                    value={roleFilter}
                    onChange={(e) => setRoleFilter(e.target.value)}
                    className="h-10 rounded-lg border border-input bg-background px-3 text-sm"
                  >
                    <option value="all">All Roles ({roleCounts.all})</option>
                    <option value="admin">Admins ({roleCounts.admin})</option>
                    <option value="freelancer">Freelancers ({roleCounts.freelancer})</option>
                    <option value="client">Clients ({roleCounts.client})</option>
                  </select>
                </div>
              </CardHeader>
              <CardContent>
                {loadingMembers ? (
                  <div className="flex justify-center py-8">
                    <div className="animate-spin w-6 h-6 border-2 border-primary border-t-transparent rounded-full" />
                  </div>
                ) : filtered.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                    <Users className="w-10 h-10 mx-auto mb-3 opacity-50" />
                    <p>No members found</p>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {filtered.map((member) => (
                      <div key={member.id} className="flex items-center gap-4 p-3 rounded-xl hover:bg-white/5 transition-colors group">
                        <div className="w-10 h-10 rounded-full bg-gradient-to-br from-primary to-blue-500 flex items-center justify-center text-xs font-bold shrink-0">
                          {member.name?.[0] || "?"}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="font-medium truncate">{member.name}</p>
                          <p className="text-xs text-muted-foreground truncate">{member.email}</p>
                        </div>
                        <span className={`px-2.5 py-1 rounded-full text-xs font-medium border ${getRoleColor(member.role)}`}>
                          {member.role}
                        </span>
                        <div className="relative">
                          <button
                            onClick={() => setActionMenuId(actionMenuId === member.id ? null : member.id)}
                            className="p-2 rounded-lg hover:bg-white/5 transition-colors opacity-0 group-hover:opacity-100"
                          >
                            <ChevronDown className="w-4 h-4" />
                          </button>
                          <AnimatePresence>
                            {actionMenuId === member.id && (
                              <motion.div
                                initial={{ opacity: 0, y: -5 }}
                                animate={{ opacity: 1, y: 0 }}
                                exit={{ opacity: 0, y: -5 }}
                                className="absolute right-0 top-full mt-1 w-48 glass-card rounded-xl border border-white/10 shadow-2xl p-1 z-10"
                              >
                                <button
                                  onClick={() => { handleRoleChange(member.id, "admin"); }}
                                  disabled={member.role === "admin" || processing}
                                  className="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors disabled:opacity-40"
                                >
                                  Make Admin
                                </button>
                                <button
                                  onClick={() => { handleRoleChange(member.id, "freelancer"); }}
                                  disabled={member.role === "freelancer" || processing}
                                  className="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors disabled:opacity-40"
                                >
                                  Make Freelancer
                                </button>
                                <button
                                  onClick={() => { handleRoleChange(member.id, "client"); }}
                                  disabled={member.role === "client" || processing}
                                  className="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-white/5 transition-colors disabled:opacity-40"
                                >
                                  Make Client
                                </button>
                                <div className="border-t border-white/5 my-1" />
                                {member.role !== "client" && (
                                  <button
                                    onClick={() => setConfirmAction({ type: "demote", id: member.id, name: member.name })}
                                    disabled={processing}
                                    className="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-white/5 text-amber-400 transition-colors disabled:opacity-40"
                                  >
                                    Demote to Guest
                                  </button>
                                )}
                                <button
                                  onClick={() => setConfirmAction({ type: "remove", id: member.id, name: member.name })}
                                  disabled={processing}
                                  className="w-full text-left px-3 py-2 text-sm rounded-lg hover:bg-white/5 text-red-400 transition-colors disabled:opacity-40"
                                >
                                  Remove Member
                                </button>
                              </motion.div>
                            )}
                          </AnimatePresence>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="settings">
            <Card>
              <CardHeader>
                <CardTitle className="text-lg">Workspace Settings</CardTitle>
              </CardHeader>
              <CardContent className="space-y-6">
                <div className="space-y-2">
                  <label className="text-sm font-medium">Workspace Name</label>
                  <Input
                    value={settings.name}
                    onChange={(e) => setSettings({ ...settings, name: e.target.value })}
                  />
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">Default Layout</label>
                  <select
                    value={settings.defaultLayout}
                    onChange={(e) => setSettings({ ...settings, defaultLayout: e.target.value })}
                    className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm"
                  >
                    <option value="open">Open</option>
                    <option value="private">Private</option>
                    <option value="hybrid">Hybrid</option>
                  </select>
                </div>

                <div className="space-y-2">
                  <label className="text-sm font-medium">Max Members</label>
                  <Input
                    type="number"
                    value={settings.maxMembers}
                    onChange={(e) => setSettings({ ...settings, maxMembers: parseInt(e.target.value) || 100 })}
                  />
                </div>

                <div className="flex items-center justify-between p-4 rounded-xl bg-white/5">
                  <div className="flex items-center gap-3">
                    {settings.guestAccess ? <Eye className="w-5 h-5 text-emerald-400" /> : <EyeOff className="w-5 h-5 text-muted-foreground" />}
                    <div>
                      <p className="text-sm font-medium">Guest Access</p>
                      <p className="text-xs text-muted-foreground">Allow guest users to view the workspace</p>
                    </div>
                  </div>
                  <button
                    onClick={() => setSettings({ ...settings, guestAccess: !settings.guestAccess })}
                    className={`w-12 h-6 rounded-full transition-colors ${settings.guestAccess ? "bg-primary" : "bg-muted"}`}
                  >
                    <div className={`w-5 h-5 rounded-full bg-white shadow transition-transform ${settings.guestAccess ? "translate-x-6" : "translate-x-0.5"}`} />
                  </button>
                </div>

                <div className="flex items-center justify-between p-4 rounded-xl bg-white/5">
                  <div className="flex items-center gap-3">
                    <Bell className={`w-5 h-5 ${settings.emailNotify ? "text-blue-400" : "text-muted-foreground"}`} />
                    <div>
                      <p className="text-sm font-medium">Email Notifications</p>
                      <p className="text-xs text-muted-foreground">Send email notifications for important events</p>
                    </div>
                  </div>
                  <button
                    onClick={() => setSettings({ ...settings, emailNotify: !settings.emailNotify })}
                    className={`w-12 h-6 rounded-full transition-colors ${settings.emailNotify ? "bg-primary" : "bg-muted"}`}
                  >
                    <div className={`w-5 h-5 rounded-full bg-white shadow transition-transform ${settings.emailNotify ? "translate-x-6" : "translate-x-0.5"}`} />
                  </button>
                </div>

                <Button onClick={handleSaveSettings} disabled={savingSettings} className="w-full sm:w-auto">
                  {savingSettings ? "Saving..." : "Save Settings"}
                </Button>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>

      <Dialog open={showInviteModal} onOpenChange={setShowInviteModal}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Invite Member</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 mt-2">
            <div className="space-y-2">
              <label className="text-sm font-medium">Email Address</label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  type="email"
                  placeholder="member@example.com"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  className="pl-9"
                />
              </div>
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Role</label>
              <select
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value)}
                className="h-10 w-full rounded-lg border border-input bg-background px-3 text-sm"
              >
                <option value="client">Client</option>
                <option value="freelancer">Freelancer</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <div className="flex gap-3 justify-end">
              <Button variant="outline" onClick={() => setShowInviteModal(false)}>Cancel</Button>
              <Button onClick={handleInvite} disabled={!inviteEmail || inviting}>
                {inviting ? "Inviting..." : "Send Invite"}
              </Button>
            </div>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={!!confirmAction} onOpenChange={() => setConfirmAction(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="w-5 h-5 text-amber-400" />
              {confirmAction?.type === "remove" ? "Remove Member" : "Demote Member"}
            </DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground mt-2">
            {confirmAction?.type === "remove"
              ? `Are you sure you want to remove ${confirmAction?.name}? This action cannot be undone.`
              : `Are you sure you want to demote ${confirmAction?.name} to guest?`}
          </p>
          <div className="flex gap-3 justify-end mt-4">
            <Button variant="outline" onClick={() => setConfirmAction(null)}>Cancel</Button>
            <Button
              variant="destructive"
              onClick={confirmAction?.type === "remove" ? handleRemove : handleDemote}
              disabled={processing}
            >
              {processing ? "Processing..." : confirmAction?.type === "remove" ? "Remove" : "Demote"}
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
