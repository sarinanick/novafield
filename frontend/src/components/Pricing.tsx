"use client";

import { motion } from "framer-motion";
import { Check, ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { AnimatedSection, StaggerContainer, StaggerItem } from "./AnimatedSection";

const plans = [
  {
    name: "Free",
    price: "$0",
    period: "per month",
    features: ["10 generations/month", "Basic AI models", "720p output", "Community support", "1 project"],
    popular: false,
    cta: "Get Started",
    variant: "outline" as const,
  },
  {
    name: "Pro",
    price: "$20",
    period: "per month",
    features: ["500 generations/month", "All 30+ AI models", "4K output", "Priority support", "Unlimited projects", "Custom characters", "API access", "Commercial license"],
    popular: true,
    cta: "Upgrade to Pro",
    variant: "glow" as const,
  },
  {
    name: "Enterprise",
    price: "Custom",
    period: "volume pricing",
    features: ["Unlimited generations", "Custom model training", "Dedicated account manager", "SLA guarantee", "On-premise deployment", "White-label options", "SSO & SAML", "Priority queue"],
    popular: false,
    cta: "Contact Sales",
    variant: "outline" as const,
  },
];

export default function Pricing() {
  return (
    <section id="pricing" className="py-24 lg:py-32 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <AnimatedSection className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl lg:text-6xl font-bold mb-6">
            Simple, transparent <span className="text-gradient">pricing</span>
          </h2>
          <p className="text-lg text-muted-foreground max-w-xl mx-auto">
            Start free, upgrade when you need more. No hidden fees.
          </p>
        </AnimatedSection>

        <StaggerContainer className="grid grid-cols-1 md:grid-cols-3 gap-6 lg:gap-8 max-w-5xl mx-auto" staggerDelay={0.15}>
          {plans.map((plan) => (
            <StaggerItem key={plan.name}>
              <motion.div
                whileHover={{ y: -8 }}
                transition={{ duration: 0.3 }}
                className={`relative rounded-2xl p-8 h-full flex flex-col ${
                  plan.popular
                    ? "glass-card border-2 border-primary/50 shadow-xl shadow-primary/10"
                    : "glass-card"
                }`}
              >
                {plan.popular && (
                  <motion.div
                    initial={{ opacity: 0, scale: 0.8 }}
                    animate={{ opacity: 1, scale: 1 }}
                    className="absolute -top-4 left-1/2 -translate-x-1/2"
                  >
                    <span className="bg-gradient-to-r from-primary to-blue-500 px-4 py-1.5 rounded-full text-sm font-medium shadow-lg shadow-primary/25">
                      Most Popular
                    </span>
                  </motion.div>
                )}

                <div className="mb-6">
                  <h3 className="text-xl font-semibold mb-2">{plan.name}</h3>
                  <div className="flex items-baseline gap-1">
                    <span className="text-4xl font-bold">{plan.price}</span>
                    {plan.price !== "Custom" && (
                      <span className="text-muted-foreground text-sm">/month</span>
                    )}
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">{plan.period}</p>
                </div>

                <ul className="space-y-3 mb-8 flex-1">
                  {plan.features.map((feature) => (
                    <li key={feature} className="flex items-center gap-3 text-sm">
                      <div className="w-5 h-5 rounded-full bg-primary/10 flex items-center justify-center shrink-0">
                        <Check className="w-3 h-3 text-primary" />
                      </div>
                      <span className="text-muted-foreground">{feature}</span>
                    </li>
                  ))}
                </ul>

                <motion.div whileHover={{ scale: 1.03 }} whileTap={{ scale: 0.97 }}>
                  <Button variant={plan.variant} className="w-full" size="lg">
                    {plan.cta}
                    <ArrowRight className="w-4 h-4 ml-2" />
                  </Button>
                </motion.div>
              </motion.div>
            </StaggerItem>
          ))}
        </StaggerContainer>
      </div>
    </section>
  );
}
