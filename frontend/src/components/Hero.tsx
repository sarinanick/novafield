"use client";

import { motion } from "framer-motion";
import { Play, ArrowRight, Zap } from "lucide-react";
import { Button } from "@/components/ui/button";
import { FloatingElement } from "./AnimatedSection";

const stats = [
  { value: "4.5M+", label: "Videos Generated" },
  { value: "30+", label: "AI Models" },
  { value: "500K+", label: "Active Creators" },
  { value: "50M+", label: "Images Created" },
];

export default function Hero() {
  return (
    <section className="relative min-h-screen flex items-center justify-center overflow-hidden pt-20">
      <div className="absolute inset-0 aurora opacity-40" />
      <div className="absolute inset-0 grid-bg" />

      <FloatingElement className="absolute top-1/4 left-[10%] opacity-20" duration={8} delay={0}>
        <div className="w-72 h-72 bg-primary/30 rounded-full blur-[100px]" />
      </FloatingElement>
      <FloatingElement className="absolute bottom-1/4 right-[10%] opacity-20" duration={10} delay={2}>
        <div className="w-96 h-96 bg-blue-500/20 rounded-full blur-[120px]" />
      </FloatingElement>
      <FloatingElement className="absolute top-1/3 right-[30%] opacity-10" duration={12} delay={4}>
        <div className="w-48 h-48 bg-emerald-500/20 rounded-full blur-[80px]" />
      </FloatingElement>

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        <motion.div
          initial={{ opacity: 0, scale: 0.9, y: 30 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          transition={{ duration: 0.8, ease: [0.25, 0.46, 0.45, 0.94] }}
        >
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.2 }}
            className="inline-flex items-center gap-2 px-4 py-2 rounded-full glass-card mb-8 cursor-pointer group"
            whileHover={{ scale: 1.05 }}
          >
            <span className="relative flex h-2 w-2">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
            </span>
            <span className="text-sm text-muted-foreground">New: Cinema Studio 4.0 is live</span>
            <ArrowRight className="w-3 h-3 text-muted-foreground group-hover:translate-x-1 transition-transform" />
          </motion.div>
        </motion.div>

        <motion.h1
          className="text-5xl sm:text-6xl md:text-7xl lg:text-8xl font-bold tracking-tight mb-6 leading-[1.05]"
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3, duration: 0.8, ease: [0.25, 0.46, 0.45, 0.94] }}
        >
          Turn ideas into
          <br />
          <span className="text-gradient">cinematic AI videos</span>
        </motion.h1>

        <motion.p
          className="text-lg sm:text-xl text-muted-foreground max-w-2xl mx-auto mb-10"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5, duration: 0.7 }}
        >
          Generate stunning videos and images with 30+ AI models. No limits, full creative control.
          From concept to cinema in seconds.
        </motion.p>

        <motion.div
          className="flex flex-col sm:flex-row gap-4 justify-center mb-20"
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6, duration: 0.7 }}
        >
          <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
            <Button variant="glow" size="xl" className="min-w-[220px]">
              <Zap className="w-5 h-5 mr-2" />
              Start Creating Free
            </Button>
          </motion.div>
          <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
            <Button variant="outline" size="xl" className="min-w-[220px] border-white/10 hover:border-white/20">
              <Play className="w-5 h-5 mr-2" />
              Watch Demo
            </Button>
          </motion.div>
        </motion.div>

        <motion.div
          className="grid grid-cols-2 md:grid-cols-4 gap-6 md:gap-8 max-w-4xl mx-auto"
          initial={{ opacity: 0, y: 40 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.8, duration: 0.7 }}
        >
          {stats.map((stat, i) => (
            <motion.div
              key={stat.label}
              className="text-center"
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.9 + i * 0.1 }}
            >
              <motion.div
                className="text-3xl md:text-4xl font-bold mb-1"
                initial={{ scale: 0.5 }}
                animate={{ scale: 1 }}
                transition={{ delay: 1 + i * 0.1, type: "spring", stiffness: 200 }}
              >
                {stat.value}
              </motion.div>
              <div className="text-muted-foreground text-sm">{stat.label}</div>
            </motion.div>
          ))}
        </motion.div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 h-32 bg-gradient-to-t from-background to-transparent" />
    </section>
  );
}
