"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import {
  Mic, MicOff, Video, VideoOff, Monitor, MonitorOff,
  Lock, Unlock, Layers, Maximize2, Minimize2, Layout,
  Music, Share2, Users, Wifi, WifiOff, AlertCircle,
} from "lucide-react";
import { useTheme } from "@/lib/theme-context";
import { motion, AnimatePresence } from "framer-motion";
import Link from "next/link";
import { worldEngine } from "@/lib/world-engine";
import { api } from "@/lib/api";

interface WorldUser {
  id: string;
  name: string;
  avatar: string;
  x: number;
  y: number;
  zone: string;
  floorId: string;
  isMoving: boolean;
  emoji?: string;
  spotifySharing?: boolean;
  spotifyTrackName?: string;
  spotifyTrackArtist?: string;
}

interface Floor {
  id: string;
  name: string;
  level: number;
  zones: { id: string; name: string; type: string; x: number; y: number; w: number; h: number; color?: string; capacity?: number }[];
  width: number;
  height: number;
  isDefault: boolean;
  createdAt: string;
}

interface CoworkingSession {
  id: string;
  hostId: string;
  host?: { id: string; name: string; avatar: string };
  type: string;
  title: string;
  zoneId: string;
  startTime: string;
  duration: number;
  participantIds: string[];
  maxParticipants: number;
  status: string;
  timerState?: { remaining: number; isPaused: boolean; phase: string };
  createdAt: string;
}

const ICE_SERVERS: RTCConfiguration = {
  iceServers: [
    { urls: "stun:stun.l.google.com:19302" },
    { urls: "stun:stun1.l.google.com:19302" },
  ],
};

const EMOJIS = ["👋", "👍", "❤️", "😂", "🎉", "🤔", "✅", "🔥"];

export default function WorldPage() {
  const { theme } = useTheme();
  const isDark = theme === "dark";
  const containerRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const realtimeWsRef = useRef<WebSocket | null>(null);
  const localStreamRef = useRef<MediaStream | null>(null);
  const videoStreamRef = useRef<MediaStream | null>(null);
  const peersRef = useRef<Map<string, RTCPeerConnection>>(new Map());
  const audioElementsRef = useRef<Map<string, HTMLAudioElement>>(new Map());
  const videoElementsRef = useRef<Map<string, HTMLVideoElement>>(new Map());
  const screenStreamRef = useRef<MediaStream | null>(null);
  const usersRef = useRef<WorldUser[]>([]);

  const [users, setUsers] = useState<WorldUser[]>([]);
  const [myId, setMyId] = useState("");
  const [myName, setMyName] = useState("");
  const [connected, setConnected] = useState(false);
  const [currentZone, setCurrentZone] = useState<string | null>(null);
  const [voiceEnabled, setVoiceEnabled] = useState(false);
  const [videoEnabled, setVideoEnabled] = useState(false);
  const [screenSharing, setScreenSharing] = useState(false);
  const [muted, setMuted] = useState(false);
  const [activeCalls, setActiveCalls] = useState<Set<string>>(new Set());
  const [lockedZones, setLockedZones] = useState<Set<string>>(new Set());
  const lockedZonesRef = useRef<Set<string>>(new Set());
  const [miniMode, setMiniMode] = useState(false);
  const [desks, setDesks] = useState<any[]>([]);
  const [showDeskModal, setShowDeskModal] = useState(false);
  const [selectedDesk, setSelectedDesk] = useState<any>(null);
  const [deskColor, setDeskColor] = useState("#6366f1");
  const [deskObjects, setDeskObjects] = useState<any[]>([]);
  const [coworkingSessions, setCoworkingSessions] = useState<CoworkingSession[]>([]);
  const [showCoworkingModal, setShowCoworkingModal] = useState(false);
  const [coworkingTitle, setCoworkingTitle] = useState("");
  const [coworkingType, setCoworkingType] = useState("focused");
  const [coworkingDuration, setCoworkingDuration] = useState(25);
  const [coworkingZone, setCoworkingZone] = useState("work");
  const [myCoworkingSession, setMyCoworkingSession] = useState<CoworkingSession | null>(null);
  const [coworkingError, setCoworkingError] = useState<string | null>(null);
  const [spotifyConnected, setSpotifyConnected] = useState(false);
  const [spotifyTrack, setSpotifyTrack] = useState<{ name: string; artist: string; albumArt: string; trackUrl: string; isPlaying: boolean } | null>(null);
  const [spotifySharing, setSpotifySharing] = useState(false);
  const [floors, setFloors] = useState<Floor[]>([]);
  const [currentFloorId, setCurrentFloorId] = useState("floor-ground");
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [engineReady, setEngineReady] = useState(false);

  useEffect(() => { usersRef.current = users; }, [users]);
  useEffect(() => { lockedZonesRef.current = lockedZones; }, [lockedZones]);

  useEffect(() => {
    if (!containerRef.current) return;
    const id = "phaser-container";
    let el = document.getElementById(id);
    if (!el) {
      el = document.createElement("div");
      el.id = id;
      el.style.width = "100%";
      el.style.height = "100%";
      containerRef.current.appendChild(el);
    }

    worldEngine.mount(id, {
      onReady: () => setEngineReady(true),
      onMove: (x: number, y: number) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          wsRef.current.send(JSON.stringify({ type: "move", payload: { x, y } }));
        }
      },
      onZoneChange: (zoneId: string | null) => setCurrentZone(zoneId),
      onDeskClick: (deskId: string) => {
        const desk = desks.find(d => d.id === deskId);
        if (desk) {
          setSelectedDesk(desk);
          setDeskColor(desk.color || "#6366f1");
          setDeskObjects(desk.objects || []);
          setShowDeskModal(true);
        }
      },
      onEmote: (emoji: string) => {
        if (wsRef.current?.readyState === WebSocket.OPEN) {
          wsRef.current.send(JSON.stringify({ type: "emoji", payload: emoji }));
        }
      },
      onProximityEnter: (playerId: string) => {
        if (voiceEnabled) createPeerConnection(playerId, true);
      },
      onProximityLeave: (playerId: string) => {
        hangupPeer(playerId);
      },
    });

    return () => { worldEngine.unmount(); };
  }, []);

  const sendSignal = useCallback((type: string, to: string, data: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, payload: { to, data } }));
    }
  }, []);

  const createPeerConnection = useCallback((peerId: string, isInitiator: boolean) => {
    if (peersRef.current.has(peerId)) return;
    const pc = new RTCPeerConnection(ICE_SERVERS);
    peersRef.current.set(peerId, pc);

    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach(t => pc.addTrack(t, localStreamRef.current!));
    }
    if (videoStreamRef.current) {
      videoStreamRef.current.getTracks().forEach(t => pc.addTrack(t, videoStreamRef.current!));
    }

    pc.onicecandidate = (e) => { if (e.candidate) sendSignal("voice-candidate", peerId, e.candidate); };
    pc.ontrack = (e) => {
      const track = e.track;
      const stream = e.streams[0];
      if (track.kind === "audio") {
        let audio = audioElementsRef.current.get(peerId);
        if (!audio) { audio = document.createElement("audio"); audio.autoplay = true; document.body.appendChild(audio); audioElementsRef.current.set(peerId, audio); }
        audio.srcObject = stream;
      } else if (track.kind === "video") {
        let video = videoElementsRef.current.get(peerId);
        if (!video) {
          video = document.createElement("video"); video.autoplay = true; video.playsInline = true;
          video.style.cssText = "width:160px;height:120px;object-fit:cover;border-radius:8px;border:1px solid #374151;";
          document.getElementById("video-overlay")?.appendChild(video);
          videoElementsRef.current.set(peerId, video);
        }
        video.srcObject = stream;
      }
    };
    pc.onconnectionstatechange = () => { if (pc.connectionState === "disconnected" || pc.connectionState === "failed") hangupPeer(peerId); };
    if (isInitiator) pc.createOffer().then(o => pc.setLocalDescription(o)).then(() => sendSignal("voice-offer", peerId, pc.localDescription)).catch(console.error);
    return pc;
  }, [sendSignal]);

  const handleOffer = useCallback(async (fromId: string, offer: RTCSessionDescriptionInit) => {
    const pc = createPeerConnection(fromId, false);
    if (!pc) return;
    await pc.setRemoteDescription(offer);
    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    sendSignal("voice-answer", fromId, pc.localDescription);
  }, [createPeerConnection, sendSignal]);

  const handleAnswer = useCallback(async (fromId: string, answer: RTCSessionDescriptionInit) => {
    const pc = peersRef.current.get(fromId);
    if (pc) await pc.setRemoteDescription(answer);
  }, []);

  const handleCandidate = useCallback(async (fromId: string, candidate: RTCIceCandidateInit) => {
    const pc = peersRef.current.get(fromId);
    if (pc) try { await pc.addIceCandidate(candidate); } catch {}
  }, []);

  const hangupPeer = useCallback((peerId: string) => {
    const pc = peersRef.current.get(peerId);
    if (pc) { pc.close(); peersRef.current.delete(peerId); }
    const audio = audioElementsRef.current.get(peerId);
    if (audio) { audio.remove(); audioElementsRef.current.delete(peerId); }
    const video = videoElementsRef.current.get(peerId);
    if (video) { video.remove(); videoElementsRef.current.delete(peerId); }
    setActiveCalls(prev => { const n = new Set(prev); n.delete(peerId); return n; });
  }, []);

  const hangupAll = useCallback(() => {
    peersRef.current.forEach((_, id) => hangupPeer(id));
  }, [hangupPeer]);

  const enableVoice = useCallback(async () => {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true } });
      localStreamRef.current = stream;
      setVoiceEnabled(true);
      setMuted(false);
    } catch (err) { console.error("Mic error:", err); }
  }, []);

  const toggleMute = useCallback(() => {
    if (localStreamRef.current) {
      localStreamRef.current.getAudioTracks().forEach(t => { t.enabled = !t.enabled; });
      setMuted(p => !p);
    }
  }, []);

  const toggleVideo = useCallback(async () => {
    if (videoEnabled) {
      videoStreamRef.current?.getTracks().forEach(t => t.stop());
      videoStreamRef.current = null;
      setVideoEnabled(false);
    } else {
      try {
        const stream = await navigator.mediaDevices.getUserMedia({ video: true });
        videoStreamRef.current = stream;
        setVideoEnabled(true);
        peersRef.current.forEach((pc, id) => { stream.getTracks().forEach(t => pc.addTrack(t, stream)); sendSignal("video-offer", id, {}); });
      } catch (err) { console.error("Video error:", err); }
    }
  }, [videoEnabled, sendSignal]);

  const toggleScreenShare = useCallback(async () => {
    if (screenSharing) {
      screenStreamRef.current?.getTracks().forEach(t => t.stop());
      screenStreamRef.current = null;
      setScreenSharing(false);
    } else {
      try {
        const stream = await navigator.mediaDevices.getDisplayMedia({ video: true, audio: true });
        screenStreamRef.current = stream;
        setScreenSharing(true);
        peersRef.current.forEach(pc => { stream.getVideoTracks().forEach(t => pc.addTrack(t, stream)); });
        stream.getVideoTracks()[0]?.addEventListener("ended", () => { setScreenSharing(false); screenStreamRef.current = null; });
      } catch {}
    }
  }, [screenSharing]);

  const fetchFloors = useCallback(async () => {
    try {
      const token = localStorage.getItem("token");
      const res = await fetch("http://localhost:3001/api/v1/floors", { headers: token ? { Authorization: `Bearer ${token}` } : {} });
      if (res.ok) { const data = await res.json(); setFloors(data); worldEngine.setFloors(data); }
    } catch {}
  }, []);

  const fetchDesks = useCallback(async () => {
    try {
      const token = localStorage.getItem("token");
      const res = await fetch("http://localhost:3001/api/v1/desks", { headers: token ? { Authorization: `Bearer ${token}` } : {} });
      if (res.ok) { const data = await res.json(); setDesks(data); }
    } catch {}
  }, []);

  const fetchCoworkingSessions = useCallback(async () => {
    try {
      const data = await api.getCoworkingSessions();
      setCoworkingSessions(data);
      const myId2 = localStorage.getItem("userId") || "";
      const mine = data.find((s: CoworkingSession) => s.hostId === myId2 || s.participantIds.includes(myId2));
      setMyCoworkingSession(mine || null);
    } catch (e: any) {
      console.error("Failed to fetch coworking sessions:", e.message);
    }
  }, []);

  const startCoworking = useCallback(async () => {
    if (!coworkingTitle.trim()) { setCoworkingError("Title is required"); return; }
    if (coworkingDuration < 1) { setCoworkingError("Duration must be at least 1 minute"); return; }
    try {
      setCoworkingError(null);
      await api.createCoworkingSession({ type: coworkingType, title: coworkingTitle.trim(), zoneId: coworkingZone, duration: coworkingDuration });
      setShowCoworkingModal(false);
      setCoworkingTitle("");
      fetchCoworkingSessions();
    } catch (e: any) {
      setCoworkingError(e.message || "Failed to create session");
    }
  }, [coworkingType, coworkingTitle, coworkingZone, coworkingDuration, fetchCoworkingSessions]);

  const joinCoworking = useCallback(async (id: string) => {
    try {
      await api.joinCoworkingSession(id);
      fetchCoworkingSessions();
    } catch (e: any) {
      console.error("Failed to join session:", e.message);
    }
  }, [fetchCoworkingSessions]);

  const leaveCoworking = useCallback(async (id: string) => {
    try {
      await api.leaveCoworkingSession(id);
      fetchCoworkingSessions();
    } catch (e: any) {
      console.error("Failed to leave session:", e.message);
    }
  }, [fetchCoworkingSessions]);

  const endCoworking = useCallback(async (id: string) => {
    try {
      await api.endCoworkingSession(id);
      fetchCoworkingSessions();
    } catch (e: any) {
      console.error("Failed to end session:", e.message);
    }
  }, [fetchCoworkingSessions]);

  const toggleCoworkingTimer = useCallback(async (id: string, isPaused: boolean) => {
    try {
      await api.updateCoworkingTimer(id, { isPaused });
      fetchCoworkingSessions();
    } catch (e: any) {
      console.error("Failed to update timer:", e.message);
    }
  }, [fetchCoworkingSessions]);

  const connectSpotify = useCallback(() => { window.location.href = `http://localhost:3001/api/v1/spotify/auth`; }, []);

  const fetchSpotifyStatus = useCallback(async () => {
    try {
      const token = localStorage.getItem("token");
      if (!token) return;
      const res = await fetch("http://localhost:3001/api/v1/spotify/status", { headers: { Authorization: `Bearer ${token}` } });
      if (res.ok) { const data = await res.json(); setSpotifyConnected(data.connected); setSpotifyTrack(data.track || null); }
    } catch {}
  }, []);

  const toggleSpotifySharing = useCallback(async (share: boolean) => {
    try { const token = localStorage.getItem("token"); if (!token) return; await fetch("http://localhost:3001/api/v1/spotify/share", { method: "POST", headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }, body: JSON.stringify({ share }) }); setSpotifySharing(share); if (share) fetchSpotifyStatus(); } catch {}
  }, [fetchSpotifyStatus]);

  const spotifyPlaybackControl = useCallback(async (action: string) => {
    try { const token = localStorage.getItem("token"); if (!token) return; await fetch("http://localhost:3001/api/v1/spotify/playback", { method: "POST", headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" }, body: JSON.stringify({ action }) }); setTimeout(fetchSpotifyStatus, 500); } catch {}
  }, [fetchSpotifyStatus]);

  const toggleLockZone = useCallback((zone: string) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      const isLocked = lockedZones.has(zone);
      wsRef.current.send(JSON.stringify({ type: isLocked ? "unlock-conversation" : "lock-conversation", payload: { zone } }));
    }
  }, [lockedZones]);

  const connect = useCallback(() => {
    const token = localStorage.getItem("token");
    const userId = localStorage.getItem("userId") || `user-${Math.random().toString(36).slice(2, 8)}`;
    const userName = localStorage.getItem("userName") || "Anonymous";
    localStorage.setItem("userId", userId);
    setMyId(userId);
    setMyName(userName);

    const ws = new WebSocket(`ws://localhost:3001/api/v1/world/ws?userId=${userId}&userName=${encodeURIComponent(userName)}`);
    wsRef.current = ws;

    ws.onopen = () => { setConnected(true); fetchFloors(); fetchDesks(); fetchCoworkingSessions(); };

    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      switch (msg.type) {
        case "init":
          setMyId(msg.payload.id);
          localStorage.setItem("userId", msg.payload.id);
          break;
        case "user-joined":
          setUsers(prev => [...prev, msg.payload]);
          worldEngine.addRemotePlayer(msg.payload);
          break;
        case "user-moved":
          setUsers(prev => prev.map(u => u.id === msg.payload.id ? { ...u, x: msg.payload.x, y: msg.payload.y } : u));
          worldEngine.updateRemotePlayer(msg.payload);
          break;
        case "user-emoji":
          setUsers(prev => prev.map(u => u.id === msg.payload.id ? { ...u, emoji: msg.payload.emoji } : u));
          worldEngine.showRemoteEmote(msg.payload.id, msg.payload.emoji);
          setTimeout(() => { setUsers(prev => prev.map(u => u.id === msg.payload.id ? { ...u, emoji: undefined } : u)); }, 3000);
          break;
        case "user-left":
          setUsers(prev => prev.filter(u => u.id !== msg.payload.id));
          worldEngine.removeRemotePlayer(msg.payload.id);
          hangupPeer(msg.payload.id);
          break;
        case "voice-offer": handleOffer(msg.payload.from, msg.payload.data); setActiveCalls(prev => new Set(prev).add(msg.payload.from)); break;
        case "voice-answer": handleAnswer(msg.payload.from, msg.payload.data); break;
        case "voice-candidate": handleCandidate(msg.payload.from, msg.payload.data); break;
        case "voice-hangup": hangupPeer(msg.payload.from); break;
        case "conversation-locked": setLockedZones(prev => new Set(prev).add(msg.payload.zone)); break;
        case "conversation-unlocked": setLockedZones(prev => { const n = new Set(prev); n.delete(msg.payload.zone); return n; }); break;
        case "user-switched-floor": setUsers(prev => prev.map(u => u.id === msg.payload.id ? { ...u, floorId: msg.payload.floorId, x: msg.payload.x, y: msg.payload.y } : u)); break;
        case "coworking-session-created":
        case "coworking-participant-joined":
        case "coworking-participant-left":
        case "coworking-timer-updated":
        case "coworking-session-ended":
          fetchCoworkingSessions();
          break;
      }
    };

    ws.onclose = () => { setConnected(false); setTimeout(connect, 2000); };
    ws.onerror = () => ws.close();
  }, [voiceEnabled, fetchFloors, fetchDesks, fetchCoworkingSessions, handleOffer, handleAnswer, handleCandidate, hangupPeer]);

  useEffect(() => { connect(); return () => { wsRef.current?.close(); realtimeWsRef.current?.close(); hangupAll(); }; }, [connect, hangupAll]);

  useEffect(() => {
    const rtWs = api.connectRealtime(myId, (msg) => {
      if (msg.type?.startsWith("coworking-")) {
        fetchCoworkingSessions();
      }
    });
    realtimeWsRef.current = rtWs;
    return () => { rtWs?.close(); };
  }, [myId, fetchCoworkingSessions]);
  useEffect(() => { fetchSpotifyStatus(); const i = setInterval(fetchSpotifyStatus, 30000); return () => clearInterval(i); }, [fetchSpotifyStatus]);
  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    if (params.get("spotify") === "connected") { fetchSpotifyStatus(); window.history.replaceState({}, "", "/world"); }
  }, [fetchSpotifyStatus]);
  useEffect(() => {
    if (!myCoworkingSession?.timerState || myCoworkingSession.timerState.isPaused) return;
    const i = setInterval(() => {
      setCoworkingSessions(prev => prev.map(s => {
        if (s.id !== myCoworkingSession.id || !s.timerState) return s;
        const r = s.timerState.remaining - 1;
        if (r <= 0) { endCoworking(s.id); return { ...s, timerState: { ...s.timerState, remaining: 0 } }; }
        return { ...s, timerState: { ...s.timerState, remaining: r } };
      }));
    }, 1000);
    return () => clearInterval(i);
  }, [myCoworkingSession?.id, myCoworkingSession?.timerState?.isPaused, endCoworking]);

  useEffect(() => {
    const h = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.key === "m") setMiniMode(p => !p);
      if (e.ctrlKey && e.shiftKey && e.key === "D") { if (voiceEnabled) toggleMute(); else enableVoice(); }
      if (e.ctrlKey && e.shiftKey && e.key === "V") toggleVideo();
    };
    window.addEventListener("keydown", h);
    return () => window.removeEventListener("keydown", h);
  }, [voiceEnabled, toggleMute, enableVoice, toggleVideo]);

  const currentFloor = floors.find(f => f.id === currentFloorId);
  const zoneLabels: Record<string, string> = {};
  if (currentFloor) currentFloor.zones.forEach(z => { zoneLabels[z.id] = z.name; });
  else { zoneLabels["work-1"] = "Work Area"; zoneLabels["meeting-1"] = "Conference Room"; zoneLabels["social-1"] = "Social Lounge"; zoneLabels["lounge-1"] = "Chill Zone"; zoneLabels["meeting-2"] = "Virtual Meeting Room"; }

  const nearbyCount = users.filter(u => u.id !== myId).length;

  return (
    <div className={`h-screen flex flex-col ${isDark ? "bg-gray-950 text-white" : "bg-gray-50 text-gray-900"}`}>
      <header className={`flex items-center justify-between px-4 py-2 border-b ${isDark ? "bg-gray-900 border-gray-800" : "bg-white border-gray-200"} z-20`}>
        <div className="flex items-center gap-3">
          <Link href="/" className="text-lg font-bold bg-gradient-to-r from-purple-400 to-cyan-400 bg-clip-text text-transparent">NovaField</Link>
          <div className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs ${connected ? "bg-green-500/10 text-green-400" : "bg-red-500/10 text-red-400"}`}>
            {connected ? <Wifi className="w-3 h-3" /> : <WifiOff className="w-3 h-3" />}
            {connected ? "Connected" : "Reconnecting..."}
          </div>
          <span className={`text-xs px-2 py-1 rounded-full ${isDark ? "bg-gray-800 text-gray-400" : "bg-gray-100 text-gray-500"}`}>
            {currentZone ? zoneLabels[currentZone] || currentZone : "Hallway"}
          </span>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setMiniMode(p => !p)} className={`p-2 rounded-lg transition-colors ${isDark ? "hover:bg-gray-800" : "hover:bg-gray-100"}`}>
            {miniMode ? <Maximize2 className="w-4 h-4" /> : <Minimize2 className="w-4 h-4" />}
          </button>
          <button onClick={() => setSidebarOpen(p => !p)} className={`p-2 rounded-lg transition-colors ${isDark ? "hover:bg-gray-800" : "hover:bg-gray-100"}`}>
            <Layers className="w-4 h-4" />
          </button>
        </div>
      </header>

      <div className="flex flex-1 overflow-hidden">
        <div ref={containerRef} className="flex-1 relative" id="game-wrapper">
          <div id="video-overlay" className="absolute bottom-4 left-4 flex gap-2 z-10" />

          <div className="absolute top-4 left-4 flex gap-2 z-10">
            <button onClick={() => voiceEnabled ? toggleMute() : enableVoice()} className={`p-2.5 rounded-xl transition-all ${voiceEnabled && !muted ? "bg-green-500/20 text-green-400 border border-green-500/30" : isDark ? "bg-gray-800/80 text-gray-400 border border-gray-700" : "bg-white/80 text-gray-600 border border-gray-300"}`}>
              {voiceEnabled && !muted ? <Mic className="w-4 h-4" /> : <MicOff className="w-4 h-4" />}
            </button>
            <button onClick={toggleVideo} className={`p-2.5 rounded-xl transition-all ${videoEnabled ? "bg-blue-500/20 text-blue-400 border border-blue-500/30" : isDark ? "bg-gray-800/80 text-gray-400 border border-gray-700" : "bg-white/80 text-gray-600 border border-gray-300"}`}>
              {videoEnabled ? <Video className="w-4 h-4" /> : <VideoOff className="w-4 h-4" />}
            </button>
            <button onClick={toggleScreenShare} className={`p-2.5 rounded-xl transition-all ${screenSharing ? "bg-purple-500/20 text-purple-400 border border-purple-500/30" : isDark ? "bg-gray-800/80 text-gray-400 border border-gray-700" : "bg-white/80 text-gray-600 border border-gray-300"}`}>
              {screenSharing ? <Monitor className="w-4 h-4" /> : <MonitorOff className="w-4 h-4" />}
            </button>
          </div>

          <div className="absolute top-4 right-4 z-10 flex items-center gap-2">
            <div className={`flex items-center gap-1 px-3 py-1.5 rounded-lg text-xs ${isDark ? "bg-gray-800/80 border border-gray-700" : "bg-white/80 border border-gray-300"}`}>
              <Users className="w-3 h-3" />
              <span>{users.length + 1}</span>
            </div>
            {floors.length > 1 && (
              <div className="flex gap-1">
                {floors.map(f => (
                  <button key={f.id} onClick={() => { setCurrentFloorId(f.id); }} className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-colors ${currentFloorId === f.id ? "bg-purple-500/20 text-purple-400 border border-purple-500/30" : isDark ? "bg-gray-800/80 text-gray-400 border border-gray-700" : "bg-white/80 text-gray-600 border border-gray-300"}`}>
                    {f.name}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        <AnimatePresence>
          {sidebarOpen && (
            <motion.aside initial={{ width: 0, opacity: 0 }} animate={{ width: 320, opacity: 1 }} exit={{ width: 0, opacity: 0 }} transition={{ duration: 0.3 }} className={`border-l overflow-y-auto ${isDark ? "bg-gray-900 border-gray-800" : "bg-white border-gray-200"}`}>
              <div className="p-4 space-y-4 w-80">
                <div className={`${isDark ? "bg-gray-800" : "bg-gray-50"} rounded-lg p-4 border ${isDark ? "border-gray-700" : "border-gray-200"}`}>
                  <h3 className="font-semibold mb-2 flex items-center gap-2"><Users className="w-4 h-4" /> Online ({users.length + 1})</h3>
                  <div className="space-y-1 max-h-32 overflow-y-auto">
                    <div className={`flex items-center gap-2 px-2 py-1.5 rounded-lg ${isDark ? "bg-gray-700/50" : "bg-white"}`}>
                      <div className="w-6 h-6 rounded-full bg-purple-500 flex items-center justify-center text-white text-xs font-bold">{myName[0]}</div>
                      <span className="text-sm font-medium">{myName} (you)</span>
                    </div>
                    {users.map(u => (
                      <div key={u.id} className={`flex items-center gap-2 px-2 py-1.5 rounded-lg ${isDark ? "hover:bg-gray-700/50" : "hover:bg-gray-100"}`}>
                        <div className="w-6 h-6 rounded-full bg-cyan-500 flex items-center justify-center text-white text-xs font-bold">{u.name[0]}</div>
                        <span className="text-sm">{u.name}</span>
                        {u.emoji && <span className="ml-auto">{u.emoji}</span>}
                      </div>
                    ))}
                  </div>
                </div>

                <div className={`${isDark ? "bg-gray-800" : "bg-gray-50"} rounded-lg p-4 border ${isDark ? "border-gray-700" : "border-gray-200"}`}>
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-semibold flex items-center gap-2"><Layout className="w-4 h-4" /> Coworking ({coworkingSessions.length})</h3>
                    <button onClick={() => setShowCoworkingModal(true)} className="text-xs px-2 py-1 rounded-lg bg-purple-500/20 text-purple-400 hover:bg-purple-500/30">+ Start</button>
                  </div>
                  <div className="space-y-2 max-h-48 overflow-y-auto">
                    {coworkingSessions.length === 0 && <p className={`text-xs ${isDark ? "text-gray-500" : "text-gray-400"}`}>No active sessions</p>}
                    {coworkingSessions.map(s => {
                      const myId3 = localStorage.getItem("userId");
                      const isHost = s.hostId === myId3;
                      const isJoined = isHost || s.participantIds.includes(myId3 || "");
                      const typeColors: Record<string, string> = { focused: "bg-blue-500/20 text-blue-400", pomodoro: "bg-orange-500/20 text-orange-400", casual: "bg-green-500/20 text-green-400" };
                      const remaining = s.timerState?.remaining || 0;
                      const mins = Math.floor(remaining / 60);
                      const secs = remaining % 60;
                      return (
                        <div key={s.id} className={`rounded-lg p-3 border ${isDark ? "bg-gray-800/50 border-gray-700" : "bg-gray-50 border-gray-200"}`}>
                          <div className="flex items-center justify-between mb-1">
                            <span className="font-medium text-sm">{s.title}</span>
                            <span className={`px-1.5 py-0.5 rounded text-[10px] font-medium ${typeColors[s.type] || typeColors.focused}`}>{s.type}</span>
                          </div>
                          <div className="flex items-center justify-between mb-1">
                            <span className={`text-xs ${isDark ? "text-gray-400" : "text-gray-500"}`}>by {s.host?.name || "Unknown"}</span>
                            <span className={`text-xs font-mono ${remaining < 60 ? "text-red-400" : isDark ? "text-gray-300" : "text-gray-600"}`}>{mins}:{secs.toString().padStart(2, "0")}</span>
                          </div>
                          <div className="flex items-center justify-between">
                            <span className={`text-xs ${isDark ? "text-gray-500" : "text-gray-400"}`}>{s.participantIds.length + 1}/{s.maxParticipants || "∞"} participant(s)</span>
                            <div className="flex gap-1">
                              {isHost ? (<>
                                <button onClick={() => toggleCoworkingTimer(s.id, !s.timerState?.isPaused)} className="px-2 py-0.5 rounded text-[10px] font-medium bg-blue-600/20 text-blue-400 hover:bg-blue-600/30">{s.timerState?.isPaused ? "Resume" : "Pause"}</button>
                                <button onClick={() => endCoworking(s.id)} className="px-2 py-0.5 rounded text-[10px] font-medium bg-red-600/20 text-red-400 hover:bg-red-600/30">End</button>
                              </>) : isJoined ? (
                                <button onClick={() => leaveCoworking(s.id)} className="px-2 py-0.5 rounded text-[10px] font-medium bg-yellow-600/20 text-yellow-400 hover:bg-yellow-600/30">Leave</button>
                              ) : (
                                <button onClick={() => joinCoworking(s.id)} className="px-2 py-0.5 rounded text-[10px] font-medium bg-green-600/20 text-green-400 hover:bg-green-600/30">Join</button>
                              )}
                            </div>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>

                <div className={`${isDark ? "bg-gray-800" : "bg-gray-50"} rounded-lg p-4 border ${isDark ? "border-gray-700" : "border-gray-200"}`}>
                  <h3 className="font-semibold mb-2">Quick Emotes</h3>
                  <div className="flex flex-wrap gap-2">
                    {EMOJIS.map(e => (
                      <button key={e} onClick={() => { if (wsRef.current?.readyState === WebSocket.OPEN) wsRef.current.send(JSON.stringify({ type: "emoji", payload: e })); worldEngine.showLocalEmote(e); }} className="text-xl hover:scale-125 transition-transform">{e}</button>
                    ))}
                  </div>
                </div>

                <div className={`${isDark ? "bg-gray-800" : "bg-gray-50"} rounded-lg p-4 border ${isDark ? "border-gray-700" : "border-gray-200"}`}>
                  <h3 className="font-semibold mb-3 flex items-center gap-2"><Music className="w-4 h-4 text-green-500" /> Spotify</h3>
                  {!spotifyConnected ? (
                    <button onClick={connectSpotify} className="w-full flex items-center justify-center gap-2 bg-[#1DB954] hover:bg-[#1ed760] px-4 py-2 rounded-lg text-sm font-medium text-white"><Music className="w-4 h-4" /> Connect Spotify</button>
                  ) : (
                    <div className="space-y-3">
                      {spotifyTrack ? (
                        <div className={`rounded-lg p-3 ${isDark ? "bg-gray-800" : "bg-gray-100"}`}>
                          <div className="flex items-center gap-3">
                            {spotifyTrack.albumArt && <img src={spotifyTrack.albumArt} alt="" className="w-12 h-12 rounded-lg object-cover" />}
                            <div className="flex-1 min-w-0">
                              <p className="text-sm font-medium truncate">{spotifyTrack.name}</p>
                              <p className={`text-xs truncate ${isDark ? "text-gray-400" : "text-gray-500"}`}>{spotifyTrack.artist}</p>
                            </div>
                          </div>
                          <div className="flex items-center gap-2 mt-2">
                            <button onClick={() => spotifyPlaybackControl("previous")} className={`p-1.5 rounded-lg ${isDark ? "hover:bg-gray-700" : "hover:bg-gray-200"}`}>⏮</button>
                            <button onClick={() => spotifyPlaybackControl(spotifyTrack.isPlaying ? "pause" : "play")} className={`p-1.5 rounded-lg ${isDark ? "hover:bg-gray-700" : "hover:bg-gray-200"}`}>{spotifyTrack.isPlaying ? "⏸" : "▶"}</button>
                            <button onClick={() => spotifyPlaybackControl("next")} className={`p-1.5 rounded-lg ${isDark ? "hover:bg-gray-700" : "hover:bg-gray-200"}`}>⏭</button>
                          </div>
                        </div>
                      ) : <p className={`text-xs ${isDark ? "text-gray-500" : "text-gray-400"}`}>No track playing</p>}
                      <div className="flex items-center justify-between">
                        <span className={`text-xs ${isDark ? "text-gray-400" : "text-gray-500"}`}>Share with others</span>
                        <button onClick={() => toggleSpotifySharing(!spotifySharing)} className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium ${spotifySharing ? "bg-green-600 text-white" : isDark ? "bg-gray-700 text-gray-300" : "bg-gray-200 text-gray-700"}`}><Share2 className="w-3 h-3" />{spotifySharing ? "Sharing" : "Share"}</button>
                      </div>
                    </div>
                  )}
                </div>

                <div className={`${isDark ? "bg-gray-800" : "bg-gray-50"} rounded-lg p-4 border ${isDark ? "border-gray-700" : "border-gray-200"}`}>
                  <h3 className="font-semibold mb-2">Zone Locks</h3>
                  <div className="space-y-2">
                    {Object.entries(zoneLabels).filter(([k]) => !k.startsWith("hallway")).map(([id, name]) => (
                      <button key={id} onClick={() => toggleLockZone(id)} className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm ${lockedZones.has(id) ? "bg-red-600/20 text-red-400" : isDark ? "bg-gray-700/50 text-gray-300" : "bg-gray-100 text-gray-700"}`}>
                        <span>{name}</span>{lockedZones.has(id) ? <Lock className="w-4 h-4" /> : <Unlock className="w-4 h-4" />}
                      </button>
                    ))}
                  </div>
                </div>

                <div className={`text-xs ${isDark ? "text-gray-400" : "text-gray-500"} space-y-1`}>
                  <p><b>WASD</b> — Move | <b>Q</b> — Emote wheel | <b>E</b> — Interact</p>
                  <p><b>1-4</b> — Jump to zone | <b>+/-</b> — Zoom | <b>Ctrl+M</b> — Mini mode</p>
                  <p>Walk near others to auto-connect voice</p>
                </div>
              </div>
            </motion.aside>
          )}
        </AnimatePresence>
      </div>

      <AnimatePresence>
        {showCoworkingModal && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowCoworkingModal(false)}>
            <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0.9, opacity: 0 }} className={`w-full max-w-md rounded-xl border shadow-2xl ${isDark ? "bg-gray-900 border-gray-700" : "bg-white border-gray-300"}`} onClick={e => e.stopPropagation()}>
              <div className="p-6">
                <h2 className="text-xl font-bold mb-4">Start Coworking Session</h2>
                <div className="space-y-4">
                  {coworkingError && (
                    <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm">
                      <AlertCircle className="w-4 h-4 flex-shrink-0" />
                      {coworkingError}
                    </div>
                  )}
                  <div>
                    <label className="text-sm font-medium mb-1 block">Title</label>
                    <input value={coworkingTitle} onChange={e => { setCoworkingTitle(e.target.value); setCoworkingError(null); }} placeholder="e.g. Sprint Planning" maxLength={100} className={`w-full px-3 py-2 rounded-lg border ${isDark ? "bg-gray-800 border-gray-700" : "bg-gray-50 border-gray-300"}`} />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <label className="text-sm font-medium mb-1 block">Type</label>
                      <select value={coworkingType} onChange={e => setCoworkingType(e.target.value)} className={`w-full px-3 py-2 rounded-lg border ${isDark ? "bg-gray-800 border-gray-700" : "bg-gray-50 border-gray-300"}`}>
                        <option value="focused">Focused</option>
                        <option value="pomodoro">Pomodoro</option>
                        <option value="casual">Casual</option>
                      </select>
                    </div>
                    <div>
                      <label className="text-sm font-medium mb-1 block">Duration (min)</label>
                      <input type="number" min={1} max={480} value={coworkingDuration} onChange={e => setCoworkingDuration(Math.max(1, +e.target.value))} className={`w-full px-3 py-2 rounded-lg border ${isDark ? "bg-gray-800 border-gray-700" : "bg-gray-50 border-gray-300"}`} />
                    </div>
                  </div>
                  <div>
                    <label className="text-sm font-medium mb-1 block">Zone</label>
                    <select value={coworkingZone} onChange={e => setCoworkingZone(e.target.value)} className={`w-full px-3 py-2 rounded-lg border ${isDark ? "bg-gray-800 border-gray-700" : "bg-gray-50 border-gray-300"}`}>
                      {Object.entries(zoneLabels).map(([id, name]) => <option key={id} value={id}>{name}</option>)}
                    </select>
                  </div>
                  <div className="flex gap-2 pt-2">
                    <button onClick={() => { setShowCoworkingModal(false); setCoworkingError(null); }} className={`flex-1 px-4 py-2 rounded-lg border ${isDark ? "border-gray-700 hover:bg-gray-800" : "border-gray-300 hover:bg-gray-100"}`}>Cancel</button>
                    <button onClick={startCoworking} disabled={!coworkingTitle} className="flex-1 px-4 py-2 rounded-lg bg-purple-600 hover:bg-purple-500 text-white font-medium disabled:opacity-50">Start Session</button>
                  </div>
                </div>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

      <AnimatePresence>
        {showDeskModal && selectedDesk && (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} exit={{ opacity: 0 }} className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm" onClick={() => setShowDeskModal(false)}>
            <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} exit={{ scale: 0.9, opacity: 0 }} className={`w-full max-w-md rounded-xl border shadow-2xl ${isDark ? "bg-gray-900 border-gray-700" : "bg-white border-gray-300"}`} onClick={e => e.stopPropagation()}>
              <div className="p-6">
                <h2 className="text-xl font-bold mb-4">Desk</h2>
                {selectedDesk.ownerId ? (
                  <p className={`text-sm ${isDark ? "text-gray-400" : "text-gray-600"}`}>This desk belongs to {selectedDesk.owner?.name || "someone"}.</p>
                ) : (
                  <p className={`text-sm ${isDark ? "text-gray-400" : "text-gray-600"}`}>This desk is empty. Click claim to make it yours!</p>
                )}
                <div className="flex gap-2 mt-4">
                  {!selectedDesk.ownerId && <button onClick={() => { /* claim */ setShowDeskModal(false); }} className="flex-1 px-4 py-2 rounded-lg bg-green-600 hover:bg-green-500 text-white font-medium">Claim Desk</button>}
                  <button onClick={() => setShowDeskModal(false)} className={`flex-1 px-4 py-2 rounded-lg border ${isDark ? "border-gray-700 hover:bg-gray-800" : "border-gray-300 hover:bg-gray-100"}`}>Close</button>
                </div>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
