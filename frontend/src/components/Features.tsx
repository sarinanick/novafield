"use client";

import { Film, Image, Video, Users, Wand2, Layers, Camera, Palette } from "lucide-react";
import { AnimatedSection, StaggerContainer, StaggerItem } from "./AnimatedSection";
import { motion } from "framer-motion";

const tools = [
  { icon: Film, title: "Cinema Studio", desc: "Create cinematic scenes with full camera control and direction", color: "from-orange-500 to-red-500", badge: "New", badgeColor: "bg-purple-500/20 text-purple-400" },
  { icon: Image, title: "Image Generator", desc: "Generate high-quality images with 15+ specialized AI models", color: "from-blue-500 to-cyan-500", badge: "Popular", badgeColor: "bg-blue-500/20 text-blue-400" },
  { icon: Video, title: "Video Generator", desc: "Create videos from text prompts in seconds with AI", color: "from-green-500 to-emerald-500", badge: "Fast", badgeColor: "bg-green-500/20 text-green-400" },
  { icon: Users, title: "Character AI", desc: "Build consistent characters across multiple video scenes", color: "from-pink-500 to-rose-500", badge: "Trending", badgeColor: "bg-pink-500/20 text-pink-400" },
  { icon: Wand2, title: "Image Editor", desc: "AI-powered inpainting, outpainting, and style transfer", color: "from-violet-500 to-purple-500", badge: null, badgeColor: "" },
  { icon: Layers, title: "Marketing Studio", desc: "Launch full marketing campaigns from a single prompt", color: "from-amber-500 to-orange-500", badge: "Hot", badgeColor: "bg-amber-500/20 text-amber-400" },
  { icon: Camera, title: "Supercomputer", desc: "One superagent for your entire creative stack", color: "from-indigo-500 to-blue-500", badge: "Pro", badgeColor: "bg-indigo-500/20 text-indigo-400" },
  { icon: Palette, title: "Style Presets", desc: "50+ viral presets from CGI to cyberpunk to anime", color: "from-teal-500 to-cyan-500", badge: null, badgeColor: "" },
];

export default function Features() {
  return (
    <section id="features" className="py-24 lg:py-32 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-16 lg:mb-20">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass-card mb-6"
          >
            <span className="text-sm text-muted-foreground">AI-Powered Creative Tools</span>
          </motion.div>
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
            Everything you need to
            <br />
            <span className="text-gradient">bring your vision to life</span>
          </h2>
          <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
            From concept to final render, NovaField gives you complete creative control
            with the most advanced AI models available.
          </p>
        </AnimatedSection>

        <StaggerContainer className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5" staggerDelay={0.08}>
          {tools.map((tool) => (
            <StaggerItem key={tool.title}>
              <motion.div
                whileHover={{ y: -8, scale: 1.02 }}
                transition={{ duration: 0.3, ease: "easeOut" }}
                className="group glass-card rounded-2xl p-6 cursor-pointer h-full relative overflow-hidden"
              >
                <div className="absolute inset-0 bg-gradient-to-br from-white/[0.02] to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

                <div className={`w-12 h-12 bg-gradient-to-br ${tool.color} rounded-xl flex items-center justify-center mb-4 shadow-lg group-hover:scale-110 transition-transform duration-300`}>
                  <tool.icon className="w-6 h-6 text-white" />
                </div>

                <h3 className="text-lg font-semibold mb-2 flex items-center gap-2">
                  {tool.title}
                  {tool.badge && (
                    <span className={`text-[10px] px-2 py-0.5 rounded-full ${tool.badgeColor} font-medium`}>
                      {tool.badge}
                    </span>
                  )}
                </h3>
                <p className="text-sm text-muted-foreground leading-relaxed">{tool.desc}</p>

                <div className="absolute bottom-0 left-0 right-0 h-1 bg-gradient-to-r from-transparent via-primary/50 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-500" />
              </motion.div>
            </StaggerItem>
          ))}
        </StaggerContainer>
      </div>
    </section>
  );
}
