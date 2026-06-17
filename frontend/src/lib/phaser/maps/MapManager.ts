import Phaser from "phaser";
import { TILE_SIZE, WORLD_WIDTH, WORLD_HEIGHT, ZoneData, DeskData, FloorData, ZONE_COLORS } from "../config";

const DEFAULT_ZONES: ZoneData[] = [
  { id: "work-1", name: "Work Area", type: "work", x: 50, y: 50, w: 400, h: 300, capacity: 12 },
  { id: "meeting-1", name: "Conference Room", type: "meeting", x: 500, y: 50, w: 300, h: 250, capacity: 8 },
  { id: "social-1", name: "Social Lounge", type: "social", x: 50, y: 400, w: 350, h: 250 },
  { id: "lounge-1", name: "Chill Zone", type: "lounge", x: 850, y: 50, w: 280, h: 200 },
  { id: "meeting-2", name: "Virtual Meeting Room", type: "virtual", x: 500, y: 350, w: 300, h: 200, capacity: 20 },
];

const DEFAULT_DESKS: DeskData[] = [
  { id: "desk-1", x: 120, y: 120 },
  { id: "desk-2", x: 220, y: 120 },
  { id: "desk-3", x: 320, y: 120 },
  { id: "desk-4", x: 120, y: 220 },
  { id: "desk-5", x: 220, y: 220 },
  { id: "desk-6", x: 320, y: 220 },
  { id: "desk-7", x: 550, y: 420 },
  { id: "desk-8", x: 650, y: 420 },
];

export class MapManager {
  private scene: Phaser.Scene;
  private floorGraphics: Phaser.GameObjects.Graphics;
  private gridGraphics: Phaser.GameObjects.Graphics;
  private floors: FloorData[] = [];
  private currentFloor: FloorData | null = null;

  constructor(scene: Phaser.Scene) {
    this.scene = scene;
    this.floorGraphics = scene.add.graphics().setDepth(0);
    this.gridGraphics = scene.add.graphics().setDepth(1);

    this.drawBaseMap();
  }

  private drawBaseMap() {
    this.floorGraphics.clear();
    this.gridGraphics.clear();

    this.floorGraphics.fillStyle(0x0a0a0f);
    this.floorGraphics.fillRect(0, 0, WORLD_WIDTH, WORLD_HEIGHT);

    this.floorGraphics.fillStyle(0x111122, 0.3);
    this.floorGraphics.fillRoundedRect(20, 20, WORLD_WIDTH - 40, WORLD_HEIGHT - 40, 16);

    this.gridGraphics.lineStyle(0.5, 0x1a1a3e, 0.15);
    for (let x = 0; x < WORLD_WIDTH; x += TILE_SIZE * 4) {
      this.gridGraphics.lineBetween(x, 0, x, WORLD_HEIGHT);
    }
    for (let y = 0; y < WORLD_HEIGHT; y += TILE_SIZE * 4) {
      this.gridGraphics.lineBetween(0, y, WORLD_WIDTH, y);
    }

    this.floorGraphics.fillStyle(0x16213e, 0.08);
    for (let x = 0; x < WORLD_WIDTH; x += TILE_SIZE) {
      for (let y = 0; y < WORLD_HEIGHT; y += TILE_SIZE) {
        if ((Math.floor(x / TILE_SIZE) + Math.floor(y / TILE_SIZE)) % 2 === 0) {
          this.floorGraphics.fillRect(x, y, TILE_SIZE, TILE_SIZE);
        }
      }
    }

    this.drawDecorations();
  }

  private drawDecorations() {
    const deco = this.scene.add.graphics().setDepth(0.5);

    deco.fillStyle(0x1a1a2e, 0.4);
    deco.fillCircle(200, 600, 60);
    deco.fillCircle(900, 500, 45);
    deco.fillCircle(1100, 150, 35);

    deco.lineStyle(1, 0xa855f7, 0.08);
    deco.strokeCircle(200, 600, 60);
    deco.strokeCircle(900, 500, 45);
    deco.strokeCircle(1100, 150, 35);

    deco.fillStyle(0x06b6d4, 0.04);
    deco.fillRoundedRect(1000, 600, 200, 150, 8);
    deco.lineStyle(1, 0x06b6d4, 0.1);
    deco.strokeRoundedRect(1000, 600, 200, 150, 8);
  }

  getZones(): ZoneData[] {
    return this.currentFloor?.zones ?? DEFAULT_ZONES;
  }

  getDesks(): DeskData[] {
    return this.currentFloor?.desks ?? DEFAULT_DESKS;
  }

  setFloors(floors: FloorData[]) {
    this.floors = floors;
    const defaultFloor = floors.find(f => f.isDefault) ?? floors[0];
    if (defaultFloor) this.setFloor(defaultFloor);
  }

  setFloor(floor: FloorData) {
    this.currentFloor = floor;
    this.drawBaseMap();
  }

  destroy() {
    this.floorGraphics.destroy();
    this.gridGraphics.destroy();
  }
}
