import Phaser from "phaser";
import { RemotePlayerData } from "../config";

const PLAYER_RADIUS = 14;

export class RemotePlayer {
  container: Phaser.GameObjects.Container;
  nameTag: Phaser.GameObjects.Text;
  statusIcon: Phaser.GameObjects.Text;
  emoteBubble: Phaser.GameObjects.Text | null = null;
  scene: Phaser.Scene;
  id: string;
  private targetX: number;
  private targetY: number;
  private currentX: number;
  private currentY: number;
  private emoteTimer: number | null = null;

  constructor(scene: Phaser.Scene, data: RemotePlayerData) {
    this.scene = scene;
    this.id = data.id;
    this.targetX = data.x;
    this.targetY = data.y;
    this.currentX = data.x;
    this.currentY = data.y;

    const colors = [0xa855f7, 0x06b6d4, 0xec4899, 0x10b981, 0xf59e0b, 0x3b82f6, 0xef4444, 0x8b5cf6];
    const hash = data.name.split("").reduce((a, c) => a + c.charCodeAt(0), 0);
    const color = colors[hash % colors.length];

    const body = scene.add.circle(0, 0, PLAYER_RADIUS, color, 0.7);
    body.setStrokeStyle(1.5, 0xffffff, 0.2);

    const eyeL = scene.add.circle(-3, -4, 2, 0xffffff, 0.8);
    const eyeR = scene.add.circle(3, -4, 2, 0xffffff, 0.8);

    this.container = scene.add.container(data.x, data.y, [body, eyeL, eyeR]);
    this.container.setDepth(9);
    this.container.setSize(PLAYER_RADIUS * 2, PLAYER_RADIUS * 2);

    this.nameTag = scene.add.text(data.x, data.y - 28, data.name, {
      fontSize: "10px",
      fontFamily: "monospace",
      color: "#cccccc",
      backgroundColor: "#000000aa",
      padding: { x: 5, y: 2 },
      align: "center",
    }).setOrigin(0.5).setDepth(10);

    this.statusIcon = scene.add.text(data.x + 18, data.y - 12, "", {
      fontSize: "11px",
    }).setOrigin(0.5).setDepth(10);

    this.updateFromData(data);
  }

  updateFromData(data: RemotePlayerData) {
    this.targetX = data.x;
    this.targetY = data.y;
    if (data.micOn) this.statusIcon.setText("🎤");
    else if (data.videoOn) this.statusIcon.setText("📹");
    else this.statusIcon.setText("");
  }

  update(_delta: number) {
    const lerpSpeed = 0.12;
    this.currentX += (this.targetX - this.currentX) * lerpSpeed;
    this.currentY += (this.targetY - this.currentY) * lerpSpeed;

    this.container.setPosition(this.currentX, this.currentY);
    this.nameTag.setPosition(this.currentX, this.currentY - 28);
    this.statusIcon.setPosition(this.currentX + 18, this.currentY - 12);

    if (this.emoteBubble) {
      this.emoteBubble.setPosition(this.currentX, this.currentY - 48);
    }
  }

  showEmote(emoji: string) {
    if (this.emoteBubble) this.emoteBubble.destroy();
    this.emoteBubble = this.scene.add.text(this.currentX, this.currentY - 48, emoji, {
      fontSize: "24px",
      backgroundColor: "#000000cc",
      padding: { x: 8, y: 4 },
    }).setOrigin(0.5).setDepth(15);

    if (this.emoteTimer !== null) clearTimeout(this.emoteTimer);
    this.emoteTimer = window.setTimeout(() => {
      if (this.emoteBubble) { this.emoteBubble.destroy(); this.emoteBubble = null; }
      this.emoteTimer = null;
    }, 3000);
  }

  destroy() {
    this.container.destroy();
    this.nameTag.destroy();
    this.statusIcon.destroy();
    if (this.emoteBubble) this.emoteBubble.destroy();
  }
}
