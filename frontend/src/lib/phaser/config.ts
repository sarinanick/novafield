import Phaser from "phaser";
import { BootScene } from "./scenes/BootScene";
import { GameScene } from "./scenes/GameScene";

export const TILE_SIZE = 32;
export const WORLD_WIDTH = 2400;
export const WORLD_HEIGHT = 1800;
export const PROXIMITY_RANGE = 150;

export const ZONE_COLORS: Record<string, number> = {
  work: 0x3b82f6,
  social: 0xf59e0b,
  meeting: 0x10b981,
  lounge: 0xa855f7,
  hallway: 0x6b7280,
  virtual: 0x06b6d4,
};

export interface ZoneData {
  id: string;
  name: string;
  type: string;
  x: number;
  y: number;
  w: number;
  h: number;
  color?: number;
  capacity?: number;
  locked?: boolean;
}

export interface DeskData {
  id: string;
  x: number;
  y: number;
  ownerId?: string;
  ownerName?: string;
  locked?: boolean;
  color?: string;
  objects?: Array<{ emoji: string; x: number; y: number; scale: number; rotation: number }>;
}

export interface FloorData {
  id: string;
  name: string;
  level: number;
  zones: ZoneData[];
  desks: DeskData[];
  width: number;
  height: number;
  isDefault?: boolean;
}

export interface RemotePlayerData {
  id: string;
  name: string;
  x: number;
  y: number;
  floor?: string;
  status?: string;
  micOn?: boolean;
  videoOn?: boolean;
}

export function createGameConfig(parent: string): Phaser.Types.Core.GameConfig {
  return {
    type: Phaser.AUTO,
    parent,
    width: "100%",
    height: "100%",
    backgroundColor: "#0a0a0f",
    scale: {
      mode: Phaser.Scale.RESIZE,
      autoCenter: Phaser.Scale.CENTER_BOTH,
    },
    physics: {
      default: "arcade",
      arcade: {
        gravity: { x: 0, y: 0 },
        debug: false,
      },
    },
    scene: [BootScene, GameScene],
    pixelArt: true,
    roundPixels: true,
  };
}
