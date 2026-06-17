"use client";

import { motion } from "framer-motion";
import { Zap, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AnimatedSection } from "./AnimatedSection";

export default function CTA() {
  return (
    <section className="py-24 lg:py-32 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection>
          <motion.div
            whileHover={{ scale: 1.01 }}
            className="relative overflow-hidden rounded-3xl"
          >
            <div className="absolute inset-0 aurora opacity-60" />
            <div className="absolute inset-0 grid-bg" />
            <div className="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-blue-500/10" />

            <div className="relative text-center py-20 px-8">
              <motion.div
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ duration: 0.7 }}
              >
                <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
                  Ready to create?
                </h2>
                <p className="text-xl text-muted-foreground mb-10 max-w-lg mx-auto">
                  Join 500,000+ creators already using NovaField to bring their ideas to life.
                </p>
                <motion.div whileHover={{ scale: 1.05 }} whileTap={{ scale: 0.95 }}>
                  <Button variant="glow" size="xl">
                    <Zap className="w-5 h-5 mr-2" />
                    Start Creating Free
                    <ArrowRight className="w-5 h-5 ml-2" />
                  </Button>
                </motion.div>
                <p className="text-sm text-muted-foreground mt-6">No credit card required</p>
              </motion.div>
            </div>
          </motion.div>
        </AnimatedSection>
      </div>
    </section>
  );
}
