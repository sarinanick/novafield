"use client";

import { Suspense } from "react";
import MarketplaceContent from "./content";

export default function MarketplacePage() {
  return (
    <Suspense fallback={<div className="min-h-screen pt-20 flex items-center justify-center"><div className="animate-spin w-8 h-8 border-2 border-primary border-t-transparent rounded-full" /></div>}>
      <MarketplaceContent />
    </Suspense>
  );
}
