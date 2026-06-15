"use client";

import dynamic from "next/dynamic";

const Navbar = dynamic(() => import("@/components/Navbar"), { ssr: false });
const Hero = dynamic(() => import("@/components/Hero"), { ssr: false });
const Features = dynamic(() => import("@/components/Features"), { ssr: false });
const Models = dynamic(() => import("@/components/Models"), { ssr: false });
const Gallery = dynamic(() => import("@/components/Gallery"), { ssr: false });
const Supercomputer = dynamic(() => import("@/components/Supercomputer"), { ssr: false });
const Pricing = dynamic(() => import("@/components/Pricing"), { ssr: false });
const CTA = dynamic(() => import("@/components/CTA"), { ssr: false });
const Footer = dynamic(() => import("@/components/Footer"), { ssr: false });

export default function Home() {
  return (
    <main className="min-h-screen bg-background noise-bg">
      <Navbar />
      <Hero />
      <Features />
      <Models />
      <Gallery />
      <Supercomputer />
      <Pricing />
      <CTA />
      <Footer />
    </main>
  );
}
