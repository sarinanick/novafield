"use client";

import { motion } from "framer-motion";
import { Cpu, Bot, Link2, Cloud, ArrowRight, Sparkles } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AnimatedSection } from "./AnimatedSection";

const capabilities = [
  { icon: Bot, label: "Agents", desc: "AI assistants" },
  { icon: Cpu, label: "Skills", desc: "Custom abilities" },
  { icon: Link2, label: "Connect", desc: "APIs & tools" },
  { icon: Cloud, label: "Drive", desc: "Cloud storage" },
];

export default function Supercomputer() {
  return (
    <section className="py-24 lg:py-32 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection>
          <motion.div
            whileHover={{ scale: 1.005 }}
            transition={{ duration: 0.4 }}
            className="glass-card rounded-3xl p-8 md:p-12 lg:p-16 relative overflow-hidden"
          >
            <div className="absolute top-0 right-0 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px]" />
            <div className="absolute bottom-0 left-0 w-[400px] h-[400px] bg-blue-500/5 rounded-full blur-[100px]" />

            <div className="relative grid lg:grid-cols-2 gap-12 items-center">
              <div>
                <motion.div
                  initial={{ opacity: 0, x: -30 }}
                  whileInView={{ opacity: 1, x: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.6 }}
                >
                  <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full glass-card text-sm text-primary mb-6">
                    <Sparkles className="w-3.5 h-3.5" />
                    New Feature
                  </span>
                  <h2 className="text-4xl md:text-5xl font-bold mb-6">
                    <span className="text-gradient">Supercomputer</span>
                  </h2>
                  <p className="text-lg text-muted-foreground mb-8 max-w-lg leading-relaxed">
                    One superagent for your entire creative stack. Build workflows, deploy agents,
                    connect APIs, and automate your creative pipeline.
                  </p>
                  <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
                    <Button variant="glow" size="lg">
                      Try Supercomputer
                      <ArrowRight className="w-4 h-4 ml-2" />
                    </Button>
                  </motion.div>
                </motion.div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                {capabilities.map((cap, i) => (
                  <motion.div
                    key={cap.label}
                    initial={{ opacity: 0, y: 20 }}
                    whileInView={{ opacity: 1, y: 0 }}
                    viewport={{ once: true }}
                    transition={{ delay: 0.2 + i * 0.1, duration: 0.5 }}
                    whileHover={{ y: -4, scale: 1.05 }}
                    className="glass-card rounded-xl p-5 text-center cursor-pointer group"
                  >
                    <cap.icon className="w-8 h-8 mx-auto mb-3 text-primary group-hover:scale-110 transition-transform duration-300" />
                    <div className="font-semibold mb-1">{cap.label}</div>
                    <div className="text-sm text-muted-foreground">{cap.desc}</div>
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.div>
        </AnimatedSection>
      </div>
    </section>
  );
}
