"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { motion } from "framer-motion";
import { Plus, Trash2, ArrowLeft } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function CreateGigPage() {
  const { user, loading: authLoading } = useAuth();
  const router = useRouter();
  const [categories, setCategories] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [step, setStep] = useState(1);

  const [form, setForm] = useState({
    title: "", description: "", category: "", subcategory: "",
    tags: [""], aiTools: [""], priceType: "fixed",
    deliveryDays: 7, revisions: 3, images: [], videoUrl: "",
  });

  const [packages, setPackages] = useState([
    { name: "Basic", description: "Simple delivery", price: 25, deliveryDays: 3, revisions: 1, features: ["High-res delivery", "Source file"] },
    { name: "Standard", description: "Professional delivery", price: 75, deliveryDays: 5, revisions: 2, features: ["High-res delivery", "Source file", "Commercial license", "Priority support"] },
    { name: "Premium", description: "Full service", price: 200, deliveryDays: 7, revisions: 5, features: ["High-res delivery", "Source file", "Commercial license", "Priority support", "Custom revisions", "Rush delivery"] },
  ]);

  useEffect(() => {
    if (!authLoading && (!user || user.role !== "freelancer")) { router.push("/auth/login"); return; }
    api.getCategories().then(setCategories).catch(() => {});
  }, [user, authLoading]);

  const updateForm = (field: string, value: any) => setForm(prev => ({ ...prev, [field]: value }));

  const addTag = () => setForm(prev => ({ ...prev, tags: [...prev.tags, ""] }));
  const removeTag = (i: number) => setForm(prev => ({ ...prev, tags: prev.tags.filter((_, idx) => idx !== i) }));
  const updateTag = (i: number, v: string) => setForm(prev => ({ ...prev, tags: prev.tags.map((t, idx) => idx === i ? v : t) }));

  const addTool = () => setForm(prev => ({ ...prev, aiTools: [...prev.aiTools, ""] }));
  const removeTool = (i: number) => setForm(prev => ({ ...prev, aiTools: prev.aiTools.filter((_, idx) => idx !== i) }));
  const updateTool = (i: number, v: string) => setForm(prev => ({ ...prev, aiTools: prev.aiTools.map((t, idx) => idx === i ? v : t) }));

  const updatePackage = (i: number, field: string, value: any) => {
    setPackages(prev => prev.map((p, idx) => idx === i ? { ...p, [field]: value } : p));
  };

  const addFeature = (pkgIdx: number) => {
    setPackages(prev => prev.map((p, idx) => idx === pkgIdx ? { ...p, features: [...p.features, ""] } : p));
  };
  const updateFeature = (pkgIdx: number, featIdx: number, value: string) => {
    setPackages(prev => prev.map((p, idx) => idx === pkgIdx ? { ...p, features: p.features.map((f, fi) => fi === featIdx ? value : f) } : p));
  };
  const removeFeature = (pkgIdx: number, featIdx: number) => {
    setPackages(prev => prev.map((p, idx) => idx === pkgIdx ? { ...p, features: p.features.filter((_, fi) => fi !== featIdx) } : p));
  };

  const handleSubmit = async () => {
    setLoading(true);
    try {
      const payload = {
        ...form,
        tags: form.tags.filter(t => t),
        aiTools: form.aiTools.filter(t => t),
        packages: packages.map(p => ({ ...p, features: p.features.filter(f => f) })),
      };
      const res = await api.createGig(payload);
      router.push(`/gig/${res.id}`);
    } catch (err: any) { alert(err.message); }
    setLoading(false);
  };

  if (authLoading) return <div className="min-h-screen pt-20 flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" /></div>;

  return (
    <div className="min-h-screen pt-20 pb-16">
      <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }}>
          <div className="flex items-center gap-3 mb-8">
            <button onClick={() => router.back()} className="p-2 hover:bg-white/10 rounded-lg"><ArrowLeft className="w-5 h-5" /></button>
            <div>
              <h1 className="text-3xl font-bold">Create a Gig</h1>
              <p className="text-muted-foreground">Step {step} of 3</p>
            </div>
          </div>

          {/* Progress */}
          <div className="flex gap-2 mb-8">
            {[1, 2, 3].map(s => (
              <div key={s} className={`h-1.5 flex-1 rounded-full transition-colors ${s <= step ? "bg-primary" : "bg-white/10"}`} />
            ))}
          </div>

          {step === 1 && (
            <Card className="glass-card border-white/5">
              <CardContent className="p-6 space-y-6">
                <h2 className="text-xl font-semibold">Gig Details</h2>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Gig Title</label>
                  <Input placeholder="I will create cinematic AI videos..." value={form.title} onChange={e => updateForm("title", e.target.value)} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Description</label>
                  <textarea className="w-full h-32 rounded-lg border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring" placeholder="Describe your service in detail..." value={form.description} onChange={e => updateForm("description", e.target.value)} />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Category</label>
                  <select className="w-full h-10 rounded-lg border border-input bg-background px-3 text-sm" value={form.category} onChange={e => updateForm("category", e.target.value)}>
                    <option value="">Select category</option>
                    {categories.map((c: any) => <option key={c.id} value={c.slug}>{c.name}</option>)}
                  </select>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Tags</label>
                  {form.tags.map((tag, i) => (
                    <div key={i} className="flex gap-2">
                      <Input placeholder="e.g. sora, cinematic" value={tag} onChange={e => updateTag(i, e.target.value)} />
                      {form.tags.length > 1 && <Button variant="ghost" size="icon" onClick={() => removeTag(i)}><Trash2 className="w-4 h-4" /></Button>}
                    </div>
                  ))}
                  <Button variant="outline" size="sm" onClick={addTag} className="border-white/10"><Plus className="w-3 h-3 mr-1" /> Add Tag</Button>
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">AI Tools Used</label>
                  {form.aiTools.map((tool, i) => (
                    <div key={i} className="flex gap-2">
                      <Input placeholder="e.g. Sora 2, Kling 3.0" value={tool} onChange={e => updateTool(i, e.target.value)} />
                      {form.aiTools.length > 1 && <Button variant="ghost" size="icon" onClick={() => removeTool(i)}><Trash2 className="w-4 h-4" /></Button>}
                    </div>
                  ))}
                  <Button variant="outline" size="sm" onClick={addTool} className="border-white/10"><Plus className="w-3 h-3 mr-1" /> Add Tool</Button>
                </div>
                <Button variant="glow" className="w-full" onClick={() => setStep(2)}>Continue</Button>
              </CardContent>
            </Card>
          )}

          {step === 2 && (
            <Card className="glass-card border-white/5">
              <CardContent className="p-6 space-y-6">
                <h2 className="text-xl font-semibold">Pricing Packages</h2>
                <div className="grid gap-6">
                  {packages.map((pkg, i) => (
                    <div key={i} className="glass-card rounded-xl p-5 space-y-4">
                      <div className="flex justify-between items-center">
                        <h3 className="font-semibold">{pkg.name}</h3>
                        <span className="text-2xl font-bold">${pkg.price}</span>
                      </div>
                      <div className="grid grid-cols-2 gap-3">
                        <div className="space-y-1">
                          <label className="text-xs text-muted-foreground">Price ($)</label>
                          <Input type="number" value={pkg.price} onChange={e => updatePackage(i, "price", Number(e.target.value))} />
                        </div>
                        <div className="space-y-1">
                          <label className="text-xs text-muted-foreground">Delivery (days)</label>
                          <Input type="number" value={pkg.deliveryDays} onChange={e => updatePackage(i, "deliveryDays", Number(e.target.value))} />
                        </div>
                      </div>
                      <div className="space-y-1">
                        <label className="text-xs text-muted-foreground">Description</label>
                        <Input value={pkg.description} onChange={e => updatePackage(i, "description", e.target.value)} />
                      </div>
                      <div className="space-y-2">
                        <label className="text-xs text-muted-foreground">Features</label>
                        {pkg.features.map((f, fi) => (
                          <div key={fi} className="flex gap-2">
                            <Input value={f} onChange={e => updateFeature(i, fi, e.target.value)} placeholder="Feature..." className="text-sm" />
                            <Button variant="ghost" size="icon" className="h-10 w-10" onClick={() => removeFeature(i, fi)}><Trash2 className="w-3 h-3" /></Button>
                          </div>
                        ))}
                        <Button variant="outline" size="sm" onClick={() => addFeature(i)} className="border-white/10"><Plus className="w-3 h-3 mr-1" /> Feature</Button>
                      </div>
                    </div>
                  ))}
                </div>
                <div className="flex gap-3">
                  <Button variant="outline" className="flex-1 border-white/10" onClick={() => setStep(1)}>Back</Button>
                  <Button variant="glow" className="flex-1" onClick={() => setStep(3)}>Continue</Button>
                </div>
              </CardContent>
            </Card>
          )}

          {step === 3 && (
            <Card className="glass-card border-white/5">
              <CardContent className="p-6 space-y-6">
                <h2 className="text-xl font-semibold">Review & Publish</h2>
                <div className="glass-card rounded-xl p-5 space-y-3">
                  <h3 className="font-semibold">{form.title || "Untitled Gig"}</h3>
                  <p className="text-sm text-muted-foreground line-clamp-2">{form.description || "No description"}</p>
                  <div className="flex flex-wrap gap-2">
                    {form.tags.filter(t => t).map(t => <span key={t} className="px-2 py-1 rounded-full bg-primary/10 text-primary text-xs">#{t}</span>)}
                    {form.aiTools.filter(t => t).map(t => <span key={t} className="px-2 py-1 rounded-full glass-card text-xs">{t}</span>)}
                  </div>
                  <div className="grid grid-cols-3 gap-3 mt-4">
                    {packages.map(p => (
                      <div key={p.name} className="text-center p-3 rounded-xl glass-card">
                        <p className="text-xs text-muted-foreground">{p.name}</p>
                        <p className="text-lg font-bold">${p.price}</p>
                      </div>
                    ))}
                  </div>
                </div>
                <div className="flex gap-3">
                  <Button variant="outline" className="flex-1 border-white/10" onClick={() => setStep(2)}>Back</Button>
                  <Button variant="glow" className="flex-1" onClick={handleSubmit} disabled={loading}>
                    {loading ? "Publishing..." : "Publish Gig"}
                  </Button>
                </div>
              </CardContent>
            </Card>
          )}
        </motion.div>
      </div>
    </div>
  );
}
