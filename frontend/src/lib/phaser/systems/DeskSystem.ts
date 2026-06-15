import Phaser from "phaser";
import { DeskData } from "../config";

export class DeskSystem {
  private scene: Phaser.Scene;
  private desks: DeskData[];
  private sprites: Map<string, Phaser.GameObjects.Container> = new Map();
  private labels: Map<string, Phaser.GameObjects.Text> = new Map();
  private hoveredDesk: string | null = null;
  private tooltip: Phaser.GameObjects.Container | null = null;
  private onDeskClick?: (deskId: string) => void;

  constructor(scene: Phaser.Scene, desks: DeskData[], onDeskClick?: (deskId: string) => void) {
    this.scene = scene;
    this.desks = desks;
    this.onDeskClick = onDeskClick;
    this.renderDesks();
  }

  private renderDesks() {
    this.sprites.forEach(s => s.destroy());
    this.labels.forEach(l => l.destroy());
    this.sprites.clear();
    this.labels.clear();

    for (const desk of this.desks) {
      const container = this.scene.add.container(desk.x, desk.y);
      container.setDepth(4);
      container.setSize(48, 32);

      const bg = this.scene.add.rectangle(0, 0, 48, 32, desk.ownerId ? 0x10b981 : 0x2d1b69, desk.ownerId ? 0.3 : 0.5);
      bg.setStrokeStyle(1, desk.ownerId ? 0x34d399 : 0xa855f7, 0.4);

      const monitor = this.scene.add.rectangle(0, -4, 20, 14, 0x1a1a2e, 0.8);
      monitor.setStrokeStyle(1, desk.ownerId ? 0x34d399 : 0x7c3aed, 0.5);

      const stand = this.scene.add.rectangle(0, 6, 6, 6, 0x374151, 0.6);

      container.add([bg, monitor, stand]);

      if (desk.ownerName) {
        const label = this.scene.add.text(desk.x, desk.y + 22, desk.ownerName, {
          fontSize: "9px",
          fontFamily: "monospace",
          color: "#a0a0a0",
          backgroundColor: "#0a0a0f99",
          padding: { x: 4, y: 1 },
        }).setOrigin(0.5).setDepth(5);
        this.labels.set(desk.id, label);
      }

      if (desk.objects) {
        for (const obj of desk.objects) {
          const emoji = this.scene.add.text(obj.x, obj.y, obj.emoji, {
            fontSize: `${Math.round(16 * obj.scale)}px`,
          }).setOrigin(0.5).setAngle(obj.rotation).setDepth(5);
          container.add(emoji);
        }
      }

      container.setInteractive(
        new Phaser.Geom.Rectangle(-24, -16, 48, 32),
        Phaser.Geom.Rectangle.Contains
      );

      container.on("pointerover", () => {
        this.hoveredDesk = desk.id;
        bg.setFillStyle(desk.ownerId ? 0x10b981 : 0x2d1b69, 0.6);
        this.showTooltip(desk);
      });

      container.on("pointerout", () => {
        this.hoveredDesk = null;
        bg.setFillStyle(desk.ownerId ? 0x10b981 : 0x2d1b69, desk.ownerId ? 0.3 : 0.5);
        this.hideTooltip();
      });

      container.on("pointerdown", () => {
        this.onDeskClick?.(desk.id);
      });

      this.sprites.set(desk.id, container);
    }
  }

  private showTooltip(desk: DeskData) {
    this.hideTooltip();
    const text = desk.ownerId ? `${desk.ownerName}'s desk` : "Empty desk — Click to claim";
    this.tooltip = this.scene.add.container(desk.x, desk.y - 30);
    this.tooltip.setDepth(20);

    const bg = this.scene.add.rectangle(0, 0, text.length * 6 + 16, 20, 0x000000, 0.85);
    bg.setStrokeStyle(1, 0xa855f7, 0.3);
    const label = this.scene.add.text(0, 0, text, {
      fontSize: "10px",
      fontFamily: "monospace",
      color: "#e0e0e0",
    }).setOrigin(0.5);
    this.tooltip.add([bg, label]);
  }

  private hideTooltip() {
    if (this.tooltip) {
      this.tooltip.destroy();
      this.tooltip = null;
    }
  }

  getHoveredDesk(): string | null {
    return this.hoveredDesk;
  }

  updateDesks(desks: DeskData[]) {
    this.desks = desks;
    this.renderDesks();
  }

  destroy() {
    this.sprites.forEach(s => s.destroy());
    this.labels.forEach(l => l.destroy());
    this.hideTooltip();
  }
}
