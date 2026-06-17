import Phaser from "phaser";
import { ZoneData, ZONE_COLORS, TILE_SIZE } from "../config";

export class ZoneSystem {
  private scene: Phaser.Scene;
  private zones: ZoneData[];
  private graphics: Phaser.GameObjects.Graphics;
  private labels: Phaser.GameObjects.Text[] = [];
  private glowGraphics: Phaser.GameObjects.Graphics;
  private currentZone: string | null = null;
  private onZoneChange?: (zoneId: string | null) => void;

  constructor(scene: Phaser.Scene, zones: ZoneData[], onZoneChange?: (zoneId: string | null) => void) {
    this.scene = scene;
    this.zones = zones;
    this.onZoneChange = onZoneChange;

    this.graphics = scene.add.graphics();
    this.graphics.setDepth(2);

    this.glowGraphics = scene.add.graphics();
    this.glowGraphics.setDepth(1);

    this.renderZones();
  }

  private renderZones() {
    this.labels.forEach(l => l.destroy());
    this.labels = [];
    this.graphics.clear();
    this.glowGraphics.clear();

    for (const zone of this.zones) {
      const color = zone.color ?? ZONE_COLORS[zone.type] ?? 0x6b7280;

      this.glowGraphics.fillStyle(color, 0.04);
      this.glowGraphics.fillRoundedRect(zone.x - 4, zone.y - 4, zone.w + 8, zone.h + 8, 8);

      this.graphics.fillStyle(color, 0.08);
      this.graphics.fillRoundedRect(zone.x, zone.y, zone.w, zone.h, 6);

      this.graphics.lineStyle(1.5, color, 0.3);
      this.graphics.strokeRoundedRect(zone.x, zone.y, zone.w, zone.h, 6);

      const gridSize = TILE_SIZE;
      this.graphics.lineStyle(0.5, color, 0.06);
      for (let gx = zone.x; gx < zone.x + zone.w; gx += gridSize) {
        this.graphics.lineBetween(gx, zone.y, gx, zone.y + zone.h);
      }
      for (let gy = zone.y; gy < zone.y + zone.h; gy += gridSize) {
        this.graphics.lineBetween(zone.x, gy, zone.x + zone.w, gy);
      }

      if (zone.locked) {
        this.graphics.lineStyle(2, 0xef4444, 0.5);
        this.graphics.strokeRoundedRect(zone.x, zone.y, zone.w, zone.h, 6);
      }

      const label = this.scene.add.text(
        zone.x + zone.w / 2,
        zone.y + 16,
        `${zone.locked ? "🔒 " : ""}${zone.name}${zone.capacity ? ` (${zone.capacity})` : ""}`,
        {
          fontSize: "11px",
          fontFamily: "monospace",
          color: `#${color.toString(16).padStart(6, "0")}`,
          backgroundColor: "#0a0a0f99",
          padding: { x: 8, y: 3 },
          align: "center",
        }
      ).setOrigin(0.5, 0).setDepth(3);

      this.labels.push(label);
    }
  }

  checkZone(x: number, y: number): string | null {
    for (const zone of this.zones) {
      if (x >= zone.x && x <= zone.x + zone.w && y >= zone.y && y <= zone.y + zone.h) {
        if (this.currentZone !== zone.id) {
          this.currentZone = zone.id;
          this.onZoneChange?.(zone.id);
        }
        return zone.id;
      }
    }
    if (this.currentZone !== null) {
      this.currentZone = null;
      this.onZoneChange?.(null);
    }
    return null;
  }

  getZoneById(id: string): ZoneData | undefined {
    return this.zones.find(z => z.id === id);
  }

  getCurrentZone(): string | null {
    return this.currentZone;
  }

  highlightZone(zoneId: string, active: boolean) {
    const zone = this.zones.find(z => z.id === zoneId);
    if (!zone) return;
    const color = zone.color ?? ZONE_COLORS[zone.type] ?? 0x6b7280;
    if (active) {
      this.glowGraphics.fillStyle(color, 0.12);
      this.glowGraphics.fillRoundedRect(zone.x - 8, zone.y - 8, zone.w + 16, zone.h + 16, 10);
    }
  }

  updateZones(zones: ZoneData[]) {
    this.zones = zones;
    this.renderZones();
  }

  destroy() {
    this.graphics.destroy();
    this.glowGraphics.destroy();
    this.labels.forEach(l => l.destroy());
  }
}
