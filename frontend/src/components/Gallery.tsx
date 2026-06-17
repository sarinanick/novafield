"use client";

import { motion } from "framer-motion";
import { Play, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AnimatedSection, StaggerContainer, StaggerItem } from "./AnimatedSection";

const presets = [
  { name: "CGI BREAKDOWN", gradient: "from-purple-600 to-blue-600" },
  { name: "KUNG FU HIT", gradient: "from-red-600 to-orange-600" },
  { name: "DRIFT RACING", gradient: "from-cyan-600 to-blue-600" },
  { name: "ZOMBIE DANCE", gradient: "from-green-600 to-emerald-600" },
  { name: "NEON CITY", gradient: "from-pink-600 to-purple-600" },
  { name: "DRAGON FANTASY", gradient: "from-yellow-600 to-amber-600" },
  { name: "RED CARPET", gradient: "from-red-600 to-rose-600" },
  { name: "OFFICE CCTV", gradient: "from-gray-600 to-slate-600" },
  { name: "ORBITAL PRESENCE", gradient: "from-indigo-600 to-violet-600" },
  { name: "SOUL FIGHTER", gradient: "from-orange-600 to-red-600" },
  { name: "NIGHT VISION", gradient: "from-green-600 to-lime-600" },
  { name: "3D RENDER", gradient: "from-blue-600 to-cyan-600" },
];

export default function Gallery() {
  return (
    <section id="gallery" className="py-24 lg:py-32 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
            Viral <span className="text-gradient">Presets</span>
          </h2>
          <p className="text-lg text-muted-foreground max-w-xl mx-auto">
            Big-budget visual effects, one click away. From explosions to surreal transformations.
          </p>
        </AnimatedSection>

        <StaggerContainer className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-3 lg:gap-4" staggerDelay={0.05}>
          {presets.map((preset) => (
            <StaggerItem key={preset.name}>
              <motion.div
                whileHover={{ scale: 1.05, y: -5 }}
                whileTap={{ scale: 0.98 }}
                transition={{ duration: 0.25 }}
                className="relative group rounded-2xl overflow-hidden aspect-[3/4] cursor-pointer"
              >
                <div className={`absolute inset-0 bg-gradient-to-br ${preset.gradient} opacity-80`} />
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-xs sm:text-sm font-bold text-center px-2">{preset.name}</span>
                </div>
                <motion.div
                  className="absolute inset-0 bg-black/60 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity duration-300"
                >
                  <div className="w-12 h-12 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center border border-white/30">
                    <Play className="w-5 h-5 text-white ml-0.5" />
                  </div>
                </motion.div>
                <div className="absolute bottom-0 left-0 right-0 h-16 bg-gradient-to-t from-black/40 to-transparent" />
              </motion.div>
            </StaggerItem>
          ))}
        </StaggerContainer>

        <AnimatedSection delay={0.3} className="text-center mt-12">
          <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
            <Button variant="outline" size="lg" className="border-white/10 hover:border-white/20">
              View All Presets
              <ArrowRight className="w-4 h-4 ml-2" />
            </Button>
          </motion.div>
        </AnimatedSection>
      </div>
    </section>
  );
}
