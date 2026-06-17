"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import {
  Undo2, Redo2, ZoomIn, ZoomOut, Save, Trash2, Plus,
  MousePointer2, Square, Circle, Home, Sofa, TreePine,
  Gamepad2, Music, Lock, Unlock, GripVertical, ChevronDown,
  ChevronRight, Eye, EyeOff, Copy, Move
} from "lucide-react";
import Link from "next/link";

interface StudioZone {
  id: string;
  name: string;
  type: string;
  x: number;
  y: number;
  w: number;
  h: number;
  color: string;
  capacity: number;
}

interface StudioObject {
  id: string;
  type: string;
  name: string;
  x: number;
  y: number;
  w: number;
  h: number;
  icon: string;
  color: string;
}

interface HistoryEntry {
  zones: StudioZone[];
  objects: StudioObject[];
}

const CANVAS_W = 1000;
const CANVAS_H = 700;

const ZONE_COLORS: Record<string, string> = {
  work: "#3b82f6",
  meeting: "#10b981",
  social: "#f59e0b",
  lounge: "#8b5cf6",
  hallway: "#6b7280",
};

const ZONE_TYPES = ["work", "meeting", "social", "lounge", "hallway"];

const CATALOG = [
  { category: "Rooms", items: [
    { type: "room", name: "Small Meeting Room", w: 120, h: 90, icon: "door", color: "#10b981" },
    { type: "room", name: "Medium Meeting Room", w: 180, h: 120, icon: "door", color: "#10b981" },
    { type: "room", name: "Large Conference Room", w: 250, h: 150, icon: "door", color: "#10b981" },
    { type: "room", name: "Private Office", w: 100, h: 80, icon: "door", color: "#6366f1" },
  ]},
  { category: "Team Areas", items: [
    { type: "area", name: "Open Workspace", w: 200, h: 150, icon: "grid", color: "#3b82f6" },
    { type: "area", name: "Team Pod", w: 140, h: 100, icon: "users", color: "#06b6d4" },
    { type: "area", name: "Focus Zone", w: 120, h: 80, icon: "target", color: "#8b5cf6" },
  ]},
  { category: "Furniture", items: [
    { type: "desk", name: "Desk Cluster", w: 80, h: 60, icon: "square", color: "#94a3b8" },
    { type: "desk", name: "Long Table", w: 160, h: 50, icon: "rectangle", color: "#94a3b8" },
    { type: "furniture", name: "Sofa", w: 100, h: 40, icon: "sofa", color: "#a78bfa" },
    { type: "furniture", name: "Coffee Table", w: 60, h: 40, icon: "circle", color: "#d4a574" },
  ]},
  { category: "Decorative", items: [
    { type: "deco", name: "Plant", w: 30, h: 30, icon: "plant", color: "#22c55e" },
    { type: "deco", name: "Bookshelf", w: 50, h: 30, icon: "book", color: "#78716c" },
    { type: "deco", name: "Lamp", w: 20, h: 20, icon: "lamp", color: "#fbbf24" },
  ]},
  { category: "Interactive", items: [
    { type: "interactive", name: "Gong", w: 40, h: 40, icon: "gong", color: "#f59e0b" },
    { type: "interactive", name: "Game Table", w: 80, h: 80, icon: "game", color: "#ef4444" },
    { type: "interactive", name: "Whiteboard", w: 120, h: 10, icon: "board", color: "#e5e7eb" },
  ]},
];

const TEMPLATES = [
  { id: "tpl-startup", name: "Startup", size: "small", layout: "open" },
  { id: "tpl-corporate", name: "Corporate", size: "medium", layout: "hybrid" },
  { id: "tpl-enterprise", name: "Enterprise", size: "large", layout: "private" },
  { id: "tpl-creative", name: "Creative Studio", size: "medium", layout: "open" },
  { id: "tpl-remote", name: "Remote Team", size: "medium", layout: "hybrid" },
  { id: "tpl-event", name: "Event Space", size: "large", layout: "open" },
];

const DEFAULT_ZONES: StudioZone[] = [
  { id: "zone-1", name: "Work Area", type: "work", x: 50, y: 50, w: 250, h: 200, color: "#3b82f6", capacity: 20 },
  { id: "zone-2", name: "Social Lounge", type: "social", x: 350, y: 50, w: 200, h: 200, color: "#f59e0b", capacity: 10 },
  { id: "zone-3", name: "Meeting Room", type: "meeting", x: 50, y: 300, w: 300, h: 150, color: "#10b981", capacity: 8 },
  { id: "zone-4", name: "Chill Zone", type: "lounge", x: 400, y: 300, w: 200, h: 150, color: "#8b5cf6", capacity: 6 },
];

export default function StudioPage() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [zones, setZones] = useState<StudioZone[]>(DEFAULT_ZONES);
  const [objects, setObjects] = useState<StudioObject[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [selectedType, setSelectedType] = useState<"zone" | "object" | null>(null);
  const [dragging, setDragging] = useState(false);
  const [resizing, setResizing] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [resizeHandle, setResizeHandle] = useState<string | null>(null);
  const [zoom, setZoom] = useState(1);
  const [history, setHistory] = useState<HistoryEntry[]>([{ zones: DEFAULT_ZONES, objects: [] }]);
  const [historyIndex, setHistoryIndex] = useState(0);
  const [expandedCatalog, setExpandedCatalog] = useState<string | null>("Rooms");
  const [published, setPublished] = useState(true);
  const [showProperties, setShowProperties] = useState(true);

  const pushHistory = useCallback((newZones: StudioZone[], newObjects: StudioObject[]) => {
    const entry: HistoryEntry = { zones: newZones, objects: newObjects };
    setHistory((prev) => [...prev.slice(0, historyIndex + 1), entry]);
    setHistoryIndex((prev) => prev + 1);
  }, [historyIndex]);

  const undo = useCallback(() => {
    if (historyIndex > 0) {
      setHistoryIndex((prev) => prev - 1);
      const entry = history[historyIndex - 1];
      setZones(entry.zones);
      setObjects(entry.objects);
    }
  }, [historyIndex, history]);

  const redo = useCallback(() => {
    if (historyIndex < history.length - 1) {
      setHistoryIndex((prev) => prev + 1);
      const entry = history[historyIndex + 1];
      setZones(entry.zones);
      setObjects(entry.objects);
    }
  }, [historyIndex, history]);

  const getCanvasPos = useCallback((e: React.MouseEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return { x: 0, y: 0 };
    const rect = canvas.getBoundingClientRect();
    return {
      x: (e.clientX - rect.left) / zoom,
      y: (e.clientY - rect.top) / zoom,
    };
  }, [zoom]);

  const findItemAt = useCallback((x: number, y: number): { id: string; type: "zone" | "object" } | null => {
    for (let i = objects.length - 1; i >= 0; i--) {
      const o = objects[i];
      if (x >= o.x && x <= o.x + o.w && y >= o.y && y <= o.y + o.h) {
        return { id: o.id, type: "object" };
      }
    }
    for (let i = zones.length - 1; i >= 0; i--) {
      const z = zones[i];
      if (x >= z.x && x <= z.x + z.w && y >= z.y && y <= z.y + z.h) {
        return { id: z.id, type: "zone" };
      }
    }
    return null;
  }, [zones, objects]);

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    const pos = getCanvasPos(e);
    const item = findItemAt(pos.x, pos.y);

    if (item) {
      setSelectedId(item.id);
      setSelectedType(item.type);
      setDragging(true);

      if (item.type === "zone") {
        const zone = zones.find((z) => z.id === item.id)!;
        setDragOffset({ x: pos.x - zone.x, y: pos.y - zone.y });
      } else {
        const obj = objects.find((o) => o.id === item.id)!;
        setDragOffset({ x: pos.x - obj.x, y: pos.y - obj.y });
      }
    } else {
      setSelectedId(null);
      setSelectedType(null);
    }
  }, [getCanvasPos, findItemAt, zones, objects]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!dragging || !selectedId || !selectedType) return;
    const pos = getCanvasPos(e);
    const newX = Math.max(0, Math.min(CANVAS_W, pos.x - dragOffset.x));
    const newY = Math.max(0, Math.min(CANVAS_H, pos.y - dragOffset.y));

    if (selectedType === "zone") {
      setZones((prev) => prev.map((z) => z.id === selectedId ? { ...z, x: newX, y: newY } : z));
    } else {
      setObjects((prev) => prev.map((o) => o.id === selectedId ? { ...o, x: newX, y: newY } : o));
    }
  }, [dragging, selectedId, selectedType, getCanvasPos, dragOffset]);

  const handleMouseUp = useCallback(() => {
    if (dragging && selectedId) {
      if (selectedType === "zone") {
        pushHistory(zones, objects);
      } else {
        pushHistory(zones, objects);
      }
    }
    setDragging(false);
    setResizing(false);
    setResizeHandle(null);
  }, [dragging, selectedId, selectedType, zones, objects, pushHistory]);

  const handleResizeStart = useCallback((e: React.MouseEvent, handle: string) => {
    e.stopPropagation();
    if (!selectedId || !selectedType) return;
    setResizing(true);
    setResizeHandle(handle);
  }, [selectedId, selectedType]);

  const addZone = useCallback((type: string) => {
    const color = ZONE_COLORS[type] || "#6b7280";
    const newZone: StudioZone = {
      id: `zone-${Date.now()}`,
      name: `${type.charAt(0).toUpperCase() + type.slice(1)} Zone`,
      type,
      x: 100 + Math.random() * 200,
      y: 100 + Math.random() * 200,
      w: 150,
      h: 120,
      color,
      capacity: 6,
    };
    const newZones = [...zones, newZone];
    setZones(newZones);
    pushHistory(newZones, objects);
    setSelectedId(newZone.id);
    setSelectedType("zone");
  }, [zones, objects, pushHistory]);

  const addObject = useCallback((item: typeof CATALOG[0]["items"][0]) => {
    const newObj: StudioObject = {
      id: `obj-${Date.now()}`,
      type: item.type,
      name: item.name,
      x: 100 + Math.random() * 300,
      y: 100 + Math.random() * 300,
      w: item.w,
      h: item.h,
      icon: item.icon,
      color: item.color,
    };
    const newObjects = [...objects, newObj];
    setObjects(newObjects);
    pushHistory(zones, newObjects);
    setSelectedId(newObj.id);
    setSelectedType("object");
  }, [zones, objects, pushHistory]);

  const deleteSelected = useCallback(() => {
    if (!selectedId || !selectedType) return;
    if (selectedType === "zone") {
      const newZones = zones.filter((z) => z.id !== selectedId);
      setZones(newZones);
      pushHistory(newZones, objects);
    } else {
      const newObjects = objects.filter((o) => o.id !== selectedId);
      setObjects(newObjects);
      pushHistory(zones, newObjects);
    }
    setSelectedId(null);
    setSelectedType(null);
  }, [selectedId, selectedType, zones, objects, pushHistory]);

  const duplicateSelected = useCallback(() => {
    if (!selectedId || !selectedType) return;
    if (selectedType === "zone") {
      const zone = zones.find((z) => z.id === selectedId);
      if (zone) {
        const newZone = { ...zone, id: `zone-${Date.now()}`, x: zone.x + 20, y: zone.y + 20 };
        const newZones = [...zones, newZone];
        setZones(newZones);
        pushHistory(newZones, objects);
      }
    } else {
      const obj = objects.find((o) => o.id === selectedId);
      if (obj) {
        const newObj = { ...obj, id: `obj-${Date.now()}`, x: obj.x + 20, y: obj.y + 20 };
        const newObjects = [...objects, newObj];
        setObjects(newObjects);
        pushHistory(zones, newObjects);
      }
    }
  }, [selectedId, selectedType, zones, objects, pushHistory]);

  const updateSelected = useCallback((field: string, value: any) => {
    if (!selectedId || !selectedType) return;
    if (selectedType === "zone") {
      const newZones = zones.map((z) => z.id === selectedId ? { ...z, [field]: value } : z);
      setZones(newZones);
      pushHistory(newZones, objects);
    } else {
      const newObjects = objects.map((o) => o.id === selectedId ? { ...o, [field]: value } : o);
      setObjects(newObjects);
      pushHistory(zones, newObjects);
    }
  }, [selectedId, selectedType, zones, objects, pushHistory]);

  const loadTemplate = useCallback((templateId: string) => {
    const template = TEMPLATES.find((t) => t.id === templateId);
    if (!template) return;
    const newZones: StudioZone[] = [
      { id: "zone-1", name: "Work Area", type: "work", x: 50, y: 50, w: template.size === "large" ? 300 : 200, h: 180, color: "#3b82f6", capacity: template.size === "large" ? 30 : 15 },
      { id: "zone-2", name: "Meeting Room", type: "meeting", x: template.size === "large" ? 400 : 300, y: 50, w: 180, h: 120, color: "#10b981", capacity: 8 },
      { id: "zone-3", name: "Social Area", type: "social", x: 50, y: 280, w: 220, h: 140, color: "#f59e0b", capacity: 10 },
      { id: "zone-4", name: "Chill Zone", type: "lounge", x: template.size === "large" ? 500 : 320, y: 280, w: template.size === "large" ? 180 : 160, h: 140, color: "#8b5cf6", capacity: 6 },
    ];
    setZones(newZones);
    setObjects([]);
    pushHistory(newZones, []);
    setSelectedId(null);
    setSelectedType(null);
  }, [pushHistory]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Delete" || e.key === "Backspace") {
        if (selectedId && !(e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement)) {
          e.preventDefault();
          deleteSelected();
        }
      }
      if (e.ctrlKey && e.key === "z") { e.preventDefault(); undo(); }
      if (e.ctrlKey && e.key === "y") { e.preventDefault(); redo(); }
      if (e.ctrlKey && e.key === "d") { e.preventDefault(); duplicateSelected(); }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [selectedId, deleteSelected, undo, redo, duplicateSelected]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    ctx.clearRect(0, 0, CANVAS_W, CANVAS_H);

    ctx.fillStyle = "#0f172a";
    ctx.fillRect(0, 0, CANVAS_W, CANVAS_H);

    ctx.strokeStyle = "#1e293b";
    ctx.lineWidth = 1;
    for (let x = 0; x < CANVAS_W; x += 32) {
      ctx.beginPath(); ctx.moveTo(x, 0); ctx.lineTo(x, CANVAS_H); ctx.stroke();
    }
    for (let y = 0; y < CANVAS_H; y += 32) {
      ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(CANVAS_W, y); ctx.stroke();
    }

    zones.forEach((zone) => {
      const isSelected = selectedId === zone.id && selectedType === "zone";

      ctx.fillStyle = zone.color + "25";
      ctx.fillRect(zone.x, zone.y, zone.w, zone.h);

      ctx.strokeStyle = isSelected ? "#ffffff" : zone.color;
      ctx.lineWidth = isSelected ? 3 : 2;
      ctx.strokeRect(zone.x, zone.y, zone.w, zone.h);

      ctx.fillStyle = zone.color;
      ctx.font = "bold 13px system-ui";
      ctx.textAlign = "center";
      ctx.fillText(zone.name, zone.x + zone.w / 2, zone.y + 22);

      ctx.fillStyle = zone.color + "80";
      ctx.font = "11px system-ui";
      ctx.fillText(`${zone.capacity} seats`, zone.x + zone.w / 2, zone.y + 38);

      if (isSelected) {
        const handles = [
          { x: zone.x, y: zone.y, cursor: "nw" },
          { x: zone.x + zone.w, y: zone.y, cursor: "ne" },
          { x: zone.x, y: zone.y + zone.h, cursor: "sw" },
          { x: zone.x + zone.w, y: zone.y + zone.h, cursor: "se" },
        ];
        handles.forEach((h) => {
          ctx.fillStyle = "#ffffff";
          ctx.fillRect(h.x - 4, h.y - 4, 8, 8);
          ctx.strokeStyle = zone.color;
          ctx.lineWidth = 1;
          ctx.strokeRect(h.x - 4, h.y - 4, 8, 8);
        });
      }
    });

    objects.forEach((obj) => {
      const isSelected = selectedId === obj.id && selectedType === "object";

      ctx.fillStyle = obj.color + "30";
      ctx.fillRect(obj.x, obj.y, obj.w, obj.h);

      ctx.strokeStyle = isSelected ? "#ffffff" : obj.color;
      ctx.lineWidth = isSelected ? 3 : 2;
      ctx.strokeRect(obj.x, obj.y, obj.w, obj.h);

      ctx.fillStyle = obj.color;
      ctx.font = "11px system-ui";
      ctx.textAlign = "center";
      ctx.fillText(obj.name, obj.x + obj.w / 2, obj.y + obj.h / 2 + 4);

      if (isSelected) {
        const handles = [
          { x: obj.x, y: obj.y },
          { x: obj.x + obj.w, y: obj.y },
          { x: obj.x, y: obj.y + obj.h },
          { x: obj.x + obj.w, y: obj.y + obj.h },
        ];
        handles.forEach((h) => {
          ctx.fillStyle = "#ffffff";
          ctx.fillRect(h.x - 4, h.y - 4, 8, 8);
        });
      }
    });
  }, [zones, objects, selectedId, selectedType, zoom]);

  const selectedZone = selectedType === "zone" ? zones.find((z) => z.id === selectedId) : null;
  const selectedObject = selectedType === "object" ? objects.find((o) => o.id === selectedId) : null;

  return (
    <div className="h-screen flex flex-col bg-gray-950 text-white">
      <header className="flex items-center justify-between px-4 py-2 bg-gray-900 border-b border-gray-800">
        <div className="flex items-center gap-3">
          <Link href="/world" className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors">
            <Home className="w-4 h-4" />
            <span className="text-sm font-medium">Back to World</span>
          </Link>
          <div className="w-px h-6 bg-gray-700" />
          <h1 className="text-lg font-bold">Gather Studio</h1>
          <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${published ? "bg-green-500/20 text-green-400" : "bg-yellow-500/20 text-yellow-400"}`}>
            {published ? "Published" : "Draft"}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={undo} disabled={historyIndex === 0} className="p-2 rounded-lg hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors" title="Undo (Ctrl+Z)">
            <Undo2 className="w-4 h-4" />
          </button>
          <button onClick={redo} disabled={historyIndex === history.length - 1} className="p-2 rounded-lg hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition-colors" title="Redo (Ctrl+Y)">
            <Redo2 className="w-4 h-4" />
          </button>
          <div className="w-px h-6 bg-gray-700" />
          <button onClick={() => setZoom((z) => Math.min(2, z + 0.1))} className="p-2 rounded-lg hover:bg-gray-800 transition-colors" title="Zoom In">
            <ZoomIn className="w-4 h-4" />
          </button>
          <span className="text-xs text-gray-400 w-12 text-center">{Math.round(zoom * 100)}%</span>
          <button onClick={() => setZoom((z) => Math.max(0.3, z - 0.1))} className="p-2 rounded-lg hover:bg-gray-800 transition-colors" title="Zoom Out">
            <ZoomOut className="w-4 h-4" />
          </button>
          <div className="w-px h-6 bg-gray-700" />
          <button onClick={() => setPublished(!published)} className={`px-3 py-1.5 rounded-lg text-sm font-medium transition-colors ${published ? "bg-green-600 hover:bg-green-500" : "bg-yellow-600 hover:bg-yellow-500"}`}>
            {published ? <><Eye className="w-4 h-4 inline mr-1" /> Published</> : <><EyeOff className="w-4 h-4 inline mr-1" /> Draft</>}
          </button>
          <button className="px-3 py-1.5 rounded-lg text-sm font-medium bg-blue-600 hover:bg-blue-500 transition-colors">
            <Save className="w-4 h-4 inline mr-1" /> Save
          </button>
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        <aside className="w-64 bg-gray-900 border-r border-gray-800 flex flex-col overflow-hidden">
          <div className="p-3 border-b border-gray-800">
            <h3 className="text-sm font-semibold text-gray-300 mb-2">Templates</h3>
            <div className="grid grid-cols-2 gap-1.5">
              {TEMPLATES.map((t) => (
                <button key={t.id} onClick={() => loadTemplate(t.id)} className="px-2 py-1.5 rounded text-xs font-medium bg-gray-800 hover:bg-gray-700 text-gray-300 hover:text-white transition-colors text-left">
                  <div>{t.name}</div>
                  <div className="text-[10px] text-gray-500">{t.size} · {t.layout}</div>
                </button>
              ))}
            </div>
          </div>
          <div className="p-3 border-b border-gray-800">
            <h3 className="text-sm font-semibold text-gray-300 mb-2">Quick Add Zone</h3>
            <div className="flex flex-wrap gap-1.5">
              {ZONE_TYPES.map((type) => (
                <button key={type} onClick={() => addZone(type)} className="flex items-center gap-1.5 px-2.5 py-1 rounded text-xs font-medium bg-gray-800 hover:bg-gray-700 text-gray-300 transition-colors">
                  <div className="w-2.5 h-2.5 rounded-sm" style={{ backgroundColor: ZONE_COLORS[type] }} />
                  {type}
                </button>
              ))}
            </div>
          </div>
          <div className="flex-1 overflow-y-auto">
            {CATALOG.map((cat) => (
              <div key={cat.category}>
                <button onClick={() => setExpandedCatalog(expandedCatalog === cat.category ? null : cat.category)} className="w-full flex items-center justify-between px-3 py-2 text-sm font-medium text-gray-300 hover:bg-gray-800/50 transition-colors">
                  <span>{cat.category}</span>
                  {expandedCatalog === cat.category ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                </button>
                <AnimatePresence>
                  {expandedCatalog === cat.category && (
                    <motion.div initial={{ height: 0, opacity: 0 }} animate={{ height: "auto", opacity: 1 }} exit={{ height: 0, opacity: 0 }} className="overflow-hidden">
                      <div className="px-2 pb-2 space-y-1">
                        {cat.items.map((item, i) => (
                          <button key={i} onClick={() => addObject(item)} className="w-full flex items-center gap-2 px-2.5 py-1.5 rounded text-xs text-gray-300 hover:bg-gray-800 hover:text-white transition-colors">
                            <div className="w-3 h-3 rounded-sm flex-shrink-0" style={{ backgroundColor: item.color }} />
                            <span className="truncate">{item.name}</span>
                            <Plus className="w-3 h-3 ml-auto opacity-0 group-hover:opacity-100" />
                          </button>
                        ))}
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            ))}
          </div>
        </aside>

        <main className="flex-1 relative overflow-hidden bg-gray-950 flex items-center justify-center">
          <div style={{ transform: `scale(${zoom})`, transformOrigin: "center" }}>
            <canvas
              ref={canvasRef}
              width={CANVAS_W}
              height={CANVAS_H}
              className="cursor-crosshair rounded-lg border border-gray-800"
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={handleMouseUp}
            />
          </div>
          <div className="absolute bottom-4 left-4 flex items-center gap-2 text-xs text-gray-500">
            <MousePointer2 className="w-3 h-3" />
            Click to select · Drag to move · Del to delete · Ctrl+D duplicate
          </div>
        </main>

        <AnimatePresence>
          {showProperties && (
            <motion.aside initial={{ width: 0, opacity: 0 }} animate={{ width: 280, opacity: 1 }} exit={{ width: 0, opacity: 0 }} className="bg-gray-900 border-l border-gray-800 flex flex-col overflow-hidden">
              <div className="p-3 border-b border-gray-800 flex items-center justify-between">
                <h3 className="text-sm font-semibold text-gray-300">Properties</h3>
                <button onClick={() => setShowProperties(false)} className="p-1 rounded hover:bg-gray-800 transition-colors">
                  <ChevronRight className="w-4 h-4 text-gray-400" />
                </button>
              </div>
              <div className="flex-1 overflow-y-auto p-3">
                {selectedZone ? (
                  <div className="space-y-3">
                    <div>
                      <label className="text-xs text-gray-400 block mb-1">Name</label>
                      <input type="text" value={selectedZone.name} onChange={(e) => updateSelected("name", e.target.value)} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                    </div>
                    <div>
                      <label className="text-xs text-gray-400 block mb-1">Type</label>
                      <select value={selectedZone.type} onChange={(e) => { updateSelected("type", e.target.value); updateSelected("color", ZONE_COLORS[e.target.value] || "#6b7280"); }} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500">
                        {ZONE_TYPES.map((t) => <option key={t} value={t}>{t}</option>)}
                      </select>
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">X</label>
                        <input type="number" value={Math.round(selectedZone.x)} onChange={(e) => updateSelected("x", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">Y</label>
                        <input type="number" value={Math.round(selectedZone.y)} onChange={(e) => updateSelected("y", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">Width</label>
                        <input type="number" value={Math.round(selectedZone.w)} onChange={(e) => updateSelected("w", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">Height</label>
                        <input type="number" value={Math.round(selectedZone.h)} onChange={(e) => updateSelected("h", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                    </div>
                    <div>
                      <label className="text-xs text-gray-400 block mb-1">Capacity</label>
                      <input type="number" value={selectedZone.capacity} onChange={(e) => updateSelected("capacity", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                    </div>
                    <div className="flex gap-2 pt-2">
                      <button onClick={duplicateSelected} className="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 rounded text-xs font-medium bg-gray-800 hover:bg-gray-700 transition-colors">
                        <Copy className="w-3 h-3" /> Duplicate
                      </button>
                      <button onClick={deleteSelected} className="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 rounded text-xs font-medium bg-red-600/20 text-red-400 hover:bg-red-600/30 transition-colors">
                        <Trash2 className="w-3 h-3" /> Delete
                      </button>
                    </div>
                  </div>
                ) : selectedObject ? (
                  <div className="space-y-3">
                    <div>
                      <label className="text-xs text-gray-400 block mb-1">Name</label>
                      <input type="text" value={selectedObject.name} onChange={(e) => updateSelected("name", e.target.value)} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">X</label>
                        <input type="number" value={Math.round(selectedObject.x)} onChange={(e) => updateSelected("x", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">Y</label>
                        <input type="number" value={Math.round(selectedObject.y)} onChange={(e) => updateSelected("y", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">Width</label>
                        <input type="number" value={Math.round(selectedObject.w)} onChange={(e) => updateSelected("w", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                      <div>
                        <label className="text-xs text-gray-400 block mb-1">Height</label>
                        <input type="number" value={Math.round(selectedObject.h)} onChange={(e) => updateSelected("h", Number(e.target.value))} className="w-full px-2.5 py-1.5 rounded bg-gray-800 border border-gray-700 text-sm text-white focus:outline-none focus:border-blue-500" />
                      </div>
                    </div>
                    <div className="flex gap-2 pt-2">
                      <button onClick={duplicateSelected} className="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 rounded text-xs font-medium bg-gray-800 hover:bg-gray-700 transition-colors">
                        <Copy className="w-3 h-3" /> Duplicate
                      </button>
                      <button onClick={deleteSelected} className="flex-1 flex items-center justify-center gap-1 px-3 py-1.5 rounded text-xs font-medium bg-red-600/20 text-red-400 hover:bg-red-600/30 transition-colors">
                        <Trash2 className="w-3 h-3" /> Delete
                      </button>
                    </div>
                  </div>
                ) : (
                  <div className="text-center text-gray-500 text-sm py-8">
                    <MousePointer2 className="w-8 h-8 mx-auto mb-2 opacity-50" />
                    <p>Select a zone or object to edit properties</p>
                  </div>
                )}
              </div>
            </motion.aside>
          )}
        </AnimatePresence>

        {!showProperties && (
          <button onClick={() => setShowProperties(true)} className="absolute right-2 top-1/2 -translate-y-1/2 p-1.5 rounded bg-gray-800 hover:bg-gray-700 transition-colors" title="Show Properties">
            <ChevronRight className="w-4 h-4 rotate-180" />
          </button>
        )}
      </div>
    </div>
  );
}
