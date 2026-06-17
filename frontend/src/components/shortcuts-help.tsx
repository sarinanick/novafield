"use client";

import { motion, AnimatePresence } from "framer-motion";
import { X, Keyboard } from "lucide-react";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useShortcuts, getShortcutDisplay, formatShortcutsForHelp } from "@/lib/shortcuts";

export default function ShortcutsHelp() {
  const { shortcuts, showHelp, setShowHelp } = useShortcuts();
  const grouped = formatShortcutsForHelp(shortcuts);

  return (
    <Dialog open={showHelp} onOpenChange={setShowHelp}>
      <DialogContent className="max-w-lg glass-card border-white/10">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Keyboard className="w-5 h-5" />
            Keyboard Shortcuts
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 max-h-[60vh] overflow-y-auto pr-2">
          {Object.entries(grouped).map(([category, items]) => (
            <div key={category}>
              <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-2">
                {category}
              </h3>
              <div className="space-y-1">
                {items.map((shortcut) => (
                  <div
                    key={shortcut.id}
                    className="flex items-center justify-between py-2 px-3 rounded-lg hover:bg-white/5 transition-colors"
                  >
                    <span className="text-sm">{shortcut.description}</span>
                    <kbd className="inline-flex items-center gap-1 px-2 py-1 text-xs font-mono text-muted-foreground bg-muted rounded-md border border-white/10">
                      {getShortcutDisplay(shortcut)}
                    </kbd>
                  </div>
                ))}
              </div>
            </div>
          ))}

          {Object.keys(grouped).length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              <Keyboard className="w-8 h-8 mx-auto mb-2 opacity-50" />
              <p className="text-sm">No shortcuts registered</p>
            </div>
          )}
        </div>

        <div className="pt-2 border-t border-white/5 text-center">
          <p className="text-xs text-muted-foreground">
            Press <kbd className="px-1.5 py-0.5 bg-muted rounded border border-white/5 font-mono">Ctrl + /</kbd> to toggle this help
          </p>
        </div>
      </DialogContent>
    </Dialog>
  );
}