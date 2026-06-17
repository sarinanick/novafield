"use client";

import { motion } from "framer-motion";
import { Sparkles } from "lucide-react";
import { AnimatedSection } from "./AnimatedSection";

const footerLinks = {
  Product: ["Features", "Pricing", "API Docs", "Enterprise", "Changelog"],
  Resources: ["Documentation", "Tutorials", "Blog", "Community", "Status"],
  Company: ["About", "Careers", "Privacy", "Terms", "Contact"],
};

const socials = [
  { name: "X", href: "#" },
  { name: "GitHub", href: "#" },
  { name: "Discord", href: "#" },
  { name: "YouTube", href: "#" },
];

export default function Footer() {
  return (
    <footer className="border-t border-white/5 relative">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <AnimatedSection>
          <div className="grid grid-cols-2 md:grid-cols-5 gap-12 mb-16">
            <div className="col-span-2">
              <div className="flex items-center gap-2.5 mb-4">
                <div className="relative w-8 h-8">
                  <div className="absolute inset-0 bg-gradient-to-br from-primary via-blue-500 to-emerald-500 rounded-lg rotate-6 opacity-70" />
                  <div className="absolute inset-0 bg-gradient-to-br from-primary via-blue-500 to-emerald-500 rounded-lg flex items-center justify-center">
                    <Sparkles className="w-4 h-4 text-white" />
                  </div>
                </div>
                <span className="text-lg font-bold">
                  Nova<span className="text-primary">Field</span>
                </span>
              </div>
              <p className="text-sm text-muted-foreground max-w-xs leading-relaxed mb-6">
                AI video and image generation platform with 30+ models. From concept to cinema in seconds.
              </p>
              <div className="flex gap-3">
                {socials.map((social) => (
                  <motion.a
                    key={social.name}
                    href={social.href}
                    whileHover={{ y: -2 }}
                    className="w-9 h-9 rounded-lg glass-card flex items-center justify-center text-xs font-medium text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {social.name[0]}
                  </motion.a>
                ))}
              </div>
            </div>

            {Object.entries(footerLinks).map(([category, links]) => (
              <div key={category}>
                <h4 className="font-semibold mb-4 text-sm">{category}</h4>
                <ul className="space-y-3">
                  {links.map((link) => (
                    <li key={link}>
                      <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                        {link}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>

          <div className="flex flex-col md:flex-row justify-between items-center pt-8 border-t border-white/5">
            <p className="text-sm text-muted-foreground">© 2024 NovaField. All rights reserved.</p>
            <div className="flex gap-6 mt-4 md:mt-0">
              <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Privacy</a>
              <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Terms</a>
              <a href="#" className="text-sm text-muted-foreground hover:text-foreground transition-colors">Cookies</a>
            </div>
          </div>
        </AnimatedSection>
      </div>
    </footer>
  );
}
