"use client";

import { motion } from "framer-motion";
import { Film, Image, Sparkles, Zap, Brain, Palette } from "lucide-react";
import { AnimatedSection, StaggerContainer, StaggerItem } from "./AnimatedSection";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";

const videoModels = [
  { id: "sora-2", name: "Sora 2", desc: "State-of-the-art video generation by OpenAI", icon: Film, color: "from-purple-500 to-blue-500" },
  { id: "kling-3", name: "Kling 3.0", desc: "High-quality cinematic videos with camera control", icon: Film, color: "from-blue-500 to-cyan-500" },
  { id: "veo-3", name: "Google Veo 3", desc: "Advanced video synthesis with physics understanding", icon: Sparkles, color: "from-green-500 to-emerald-500" },
  { id: "seedance-2", name: "Seedance 2.0", desc: "Create videos in seconds with motion control", icon: Zap, color: "from-orange-500 to-red-500" },
  { id: "minimax", name: "MiniMax Hailuo", desc: "Realistic video generation with lip sync", icon: Film, color: "from-pink-500 to-rose-500" },
];

const imageModels = [
  { id: "soul-v2", name: "Higgsfield Soul", desc: "Consistent character generation across scenes", icon: Brain, color: "from-violet-500 to-purple-500" },
  { id: "nano-banana", name: "Nano Banana", desc: "High-quality visual generation with style control", icon: Palette, color: "from-amber-500 to-orange-500" },
  { id: "gpt-image", name: "GPT Image", desc: "OpenAI's latest image generation model", icon: Image, color: "from-teal-500 to-cyan-500" },
  { id: "recraft-v4", name: "Recraft 4.1", desc: "Crisp vectors, refined aesthetics, total control", icon: Palette, color: "from-indigo-500 to-blue-500" },
  { id: "flux", name: "Flux Kontext", desc: "Context-aware image generation", icon: Sparkles, color: "from-emerald-500 to-green-500" },
];

export default function Models() {
  return (
    <section id="tools" className="py-24 lg:py-32 relative">
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-primary/[0.02] to-transparent" />
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
        <AnimatedSection className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
            30+ <span className="text-gradient">AI Models</span>
          </h2>
          <p className="text-lg text-muted-foreground max-w-xl mx-auto">
            Access the most powerful AI models in one platform. From video to image generation.
          </p>
        </AnimatedSection>

        <Tabs defaultValue="video" className="w-full">
          <AnimatedSection delay={0.2} className="flex justify-center mb-12">
            <TabsList className="glass-card border border-white/10 p-1 h-12">
              <TabsTrigger value="video" className="px-6 data-[state=active]:bg-primary data-[state=active]:text-white rounded-lg transition-all duration-300">
                <Film className="w-4 h-4 mr-2" />
                Video Models
              </TabsTrigger>
              <TabsTrigger value="image" className="px-6 data-[state=active]:bg-primary data-[state=active]:text-white rounded-lg transition-all duration-300">
                <Image className="w-4 h-4 mr-2" />
                Image Models
              </TabsTrigger>
            </TabsList>
          </AnimatedSection>

          <TabsContent value="video">
            <StaggerContainer className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4" staggerDelay={0.08}>
              {videoModels.map((model) => (
                <StaggerItem key={model.id}>
                  <motion.div
                    whileHover={{ y: -6, scale: 1.02 }}
                    className="glass-card rounded-2xl p-6 cursor-pointer group relative overflow-hidden"
                  >
                    <div className="absolute inset-0 bg-gradient-to-br from-white/[0.01] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                    <div className="flex items-start gap-4">
                      <div className={`w-12 h-12 bg-gradient-to-br ${model.color} rounded-xl flex items-center justify-center shrink-0 shadow-lg group-hover:scale-110 transition-transform duration-300`}>
                        <model.icon className="w-6 h-6 text-white" />
                      </div>
                      <div>
                        <h3 className="font-semibold mb-1">{model.name}</h3>
                        <p className="text-sm text-muted-foreground">{model.desc}</p>
                      </div>
                    </div>
                  </motion.div>
                </StaggerItem>
              ))}
            </StaggerContainer>
          </TabsContent>

          <TabsContent value="image">
            <StaggerContainer className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4" staggerDelay={0.08}>
              {imageModels.map((model) => (
                <StaggerItem key={model.id}>
                  <motion.div
                    whileHover={{ y: -6, scale: 1.02 }}
                    className="glass-card rounded-2xl p-6 cursor-pointer group relative overflow-hidden"
                  >
                    <div className="absolute inset-0 bg-gradient-to-br from-white/[0.01] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
                    <div className="flex items-start gap-4">
                      <div className={`w-12 h-12 bg-gradient-to-br ${model.color} rounded-xl flex items-center justify-center shrink-0 shadow-lg group-hover:scale-110 transition-transform duration-300`}>
                        <model.icon className="w-6 h-6 text-white" />
                      </div>
                      <div>
                        <h3 className="font-semibold mb-1">{model.name}</h3>
                        <p className="text-sm text-muted-foreground">{model.desc}</p>
                      </div>
                    </div>
                  </motion.div>
                </StaggerItem>
              ))}
            </StaggerContainer>
          </TabsContent>
        </Tabs>
      </div>
    </section>
  );
}
