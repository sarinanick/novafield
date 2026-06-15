"use client";

import { useEffect, useCallback, useState, createContext, useContext, ReactNode } from "react";
import { useRouter } from "next/navigation";

export interface Shortcut {
  id: string;
  key: string;
  ctrl?: boolean;
  shift?: boolean;
  alt?: boolean;
  description: string;
  category: string;
  action: () => void;
  enabled?: boolean;
}

interface ShortcutsContextType {
  shortcuts: Shortcut[];
  registerShortcut: (shortcut: Omit<Shortcut, "id">) => string;
  unregisterShortcut: (id: string) => void;
  showHelp: boolean;
  setShowHelp: (show: boolean) => void;
  searchOpen: boolean;
  setSearchOpen: (open: boolean) => void;
}

const ShortcutsContext = createContext<ShortcutsContextType | null>(null);

export function useShortcuts() {
  const ctx = useContext(ShortcutsContext);
  if (!ctx) throw new Error("useShortcuts must be used within ShortcutsProvider");
  return ctx;
}

let shortcutIdCounter = 0;

export function ShortcutsProvider({ children }: { children: ReactNode }) {
  const router = useRouter();
  const [shortcuts, setShortcuts] = useState<Shortcut[]>([]);
  const [showHelp, setShowHelp] = useState(false);
  const [searchOpen, setSearchOpen] = useState(false);

  const registerShortcut = useCallback((shortcut: Omit<Shortcut, "id">) => {
    const id = `shortcut-${++shortcutIdCounter}`;
    setShortcuts((prev) => [...prev, { ...shortcut, id, enabled: true }]);
    return id;
  }, []);

  const unregisterShortcut = useCallback((id: string) => {
    setShortcuts((prev) => prev.filter((s) => s.id !== id));
  }, []);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        setSearchOpen(false);
        setShowHelp(false);
        return;
      }

      if ((e.ctrlKey || e.metaKey) && e.key === "k") {
        e.preventDefault();
        setSearchOpen((prev) => !prev);
        return;
      }

      if ((e.ctrlKey || e.metaKey) && e.key === "/") {
        e.preventDefault();
        setShowHelp((prev) => !prev);
        return;
      }

      shortcuts.forEach((shortcut) => {
        if (shortcut.enabled === false) return;

        const ctrlMatch = shortcut.ctrl ? (e.ctrlKey || e.metaKey) : true;
        const shiftMatch = shortcut.shift ? e.shiftKey : true;
        const altMatch = shortcut.alt ? e.altKey : true;
        const keyMatch = e.key.toLowerCase() === shortcut.key.toLowerCase();

        if (ctrlMatch && shiftMatch && altMatch && keyMatch) {
          e.preventDefault();
          shortcut.action();
        }
      });
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [shortcuts, setSearchOpen, setShowHelp]);

  useEffect(() => {
    const defaultShortcuts: Omit<Shortcut, "id">[] = [
      {
        key: "d",
        ctrl: true,
        description: "Return to desk",
        category: "Navigation",
        action: () => router.push("/dashboard"),
      },
      {
        key: "m",
        ctrl: true,
        description: "Toggle Mini Mode",
        category: "World",
        action: () => {
          window.dispatchEvent(new CustomEvent("shortcut:toggle-mini"));
        },
      },
      {
        key: "d",
        ctrl: true,
        shift: true,
        description: "Toggle mic on/off",
        category: "World",
        action: () => {
          window.dispatchEvent(new CustomEvent("shortcut:toggle-mic"));
        },
      },
      {
        key: "v",
        ctrl: true,
        shift: true,
        description: "Toggle video on/off",
        category: "World",
        action: () => {
          window.dispatchEvent(new CustomEvent("shortcut:toggle-video"));
        },
      },
      {
        key: "1",
        description: "Jump to Work Area",
        category: "Zones",
        action: () => {
          window.dispatchEvent(new CustomEvent("shortcut:jump-zone", { detail: "work" }));
        },
      },
      {
        key: "2",
        description: "Jump to Social Lounge",
        category: "Zones",
        action: () => {
          window.dispatchEvent(new CustomEvent("shortcut:jump-zone", { detail: "social" }));
        },
      },
      {
        key: "3",
        description: "Jump to Meeting Room",
        category: "Zones",
        action: () => {
          window.dispatchEvent(new CustomEvent("shortcut:jump-zone", { detail: "meeting" }));
        },
      },
      {
        key: "4",
        description: "Jump to Chill Zone",
        category: "Zones",
        action: () => {
          window.dispatchEvent(new CustomEvent("shortcut:jump-zone", { detail: "lounge" }));
        },
      },
    ];

    const ids = defaultShortcuts.map((s) => registerShortcut(s));
    return () => ids.forEach((id) => unregisterShortcut(id));
  }, [registerShortcut, unregisterShortcut, router]);

  return (
    <ShortcutsContext.Provider
      value={{
        shortcuts,
        registerShortcut,
        unregisterShortcut,
        showHelp,
        setShowHelp,
        searchOpen,
        setSearchOpen,
      }}
    >
      {children}
    </ShortcutsContext.Provider>
  );
}

export function getShortcutDisplay(shortcut: { key: string; ctrl?: boolean; shift?: boolean; alt?: boolean }): string {
  const parts: string[] = [];
  if (shortcut.ctrl) parts.push("Ctrl");
  if (shortcut.shift) parts.push("Shift");
  if (shortcut.alt) parts.push("Alt");
  parts.push(shortcut.key.toUpperCase());
  return parts.join(" + ");
}

export function formatShortcutsForHelp(shortcuts: Shortcut[]) {
  const grouped: Record<string, Shortcut[]> = {};
  shortcuts.forEach((s) => {
    if (!grouped[s.category]) grouped[s.category] = [];
    grouped[s.category].push(s);
  });
  return grouped;
}