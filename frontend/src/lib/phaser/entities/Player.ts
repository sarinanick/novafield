import Phaser from "phaser";
import { WORLD_WIDTH, WORLD_HEIGHT } from "../config";

const MOVE_SPEED = 160;
const PLAYER_RADIUS = 14;

export class Player {
  container: Phaser.GameObjects.Container;
  nameTag: Phaser.GameObjects.Text;
  statusIcon: Phaser.GameObjects.Text;
  scene: Phaser.Scene;
  x: number;
  y: number;
  private keys!: Record<string, Phaser.Input.Keyboard.Key>;
  private nameText: string;
  private lastSentX = 0;
  private lastSentY = 0;
  private sendTimer = 0;

  constructor(scene: Phaser.Scene, x: number, y: number, name: string, colorIndex: number) {
    this.scene = scene;
    this.x = x;
    this.y = y;
    this.nameText = name;

    const colors = [0xa855f7, 0x06b6d4, 0xec4899, 0x10b981, 0xf59e0b, 0x3b82f6, 0xef4444, 0x8b5cf6];
    const color = colors[colorIndex % colors.length];

    const body = scene.add.circle(0, 0, PLAYER_RADIUS, color, 0.9);
    body.setStrokeStyle(2, 0xffffff, 0.3);

    const eyeL = scene.add.circle(-3, -4, 2, 0xffffff);
    const eyeR = scene.add.circle(3, -4, 2, 0xffffff);
    const pupilL = scene.add.circle(-3, -4, 1, 0x0a0a0f);
    const pupilR = scene.add.circle(3, -4, 1, 0x0a0a0f);

    this.container = scene.add.container(x, y, [body, eyeL, eyeR, pupilL, pupilR]);
    this.container.setDepth(10);
    this.container.setSize(PLAYER_RADIUS * 2, PLAYER_RADIUS * 2);

    this.nameTag = scene.add.text(x, y - 28, name, {
      fontSize: "11px",
      fontFamily: "monospace",
      color: "#ffffff",
      backgroundColor: "#000000aa",
      padding: { x: 6, y: 2 },
      align: "center",
    }).setOrigin(0.5).setDepth(11);

    this.statusIcon = scene.add.text(x + 18, y - 12, "", {
      fontSize: "12px",
    }).setOrigin(0.5).setDepth(11);

    if (scene.input.keyboard) {
      this.keys = {
        up: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.UP),
        down: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.DOWN),
        left: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.LEFT),
        right: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.RIGHT),
        w: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.W),
        s: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.S),
        a: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.A),
        d: scene.input.keyboard.addKey(Phaser.Input.Keyboard.KeyCodes.D),
      };
    }
  }

  update(delta: number): { x: number; y: number; moved: boolean } {
    let vx = 0;
    let vy = 0;

    if (this.keys) {
      if (this.keys.left.isDown || this.keys.a.isDown) vx = -1;
      if (this.keys.right.isDown || this.keys.d.isDown) vx = 1;
      if (this.keys.up.isDown || this.keys.w.isDown) vy = -1;
      if (this.keys.down.isDown || this.keys.s.isDown) vy = 1;
    }

    if (vx !== 0 && vy !== 0) {
      vx *= 0.707;
      vy *= 0.707;
    }

    const speed = MOVE_SPEED * (delta / 1000);
    this.x = Phaser.Math.Clamp(this.x + vx * speed, PLAYER_RADIUS, WORLD_WIDTH - PLAYER_RADIUS);
    this.y = Phaser.Math.Clamp(this.y + vy * speed, PLAYER_RADIUS, WORLD_HEIGHT - PLAYER_RADIUS);

    const bobble = vx !== 0 || vy !== 0 ? Math.sin(Date.now() / 100) * 1.5 : 0;
    this.container.setPosition(this.x, this.y + bobble);
    this.nameTag.setPosition(this.x, this.y - 28);
    this.statusIcon.setPosition(this.x + 18, this.y - 12);

    this.sendTimer += delta;
    const moved = Math.abs(this.x - this.lastSentX) > 1 || Math.abs(this.y - this.lastSentY) > 1;
    if (moved && this.sendTimer > 50) {
      this.lastSentX = this.x;
      this.lastSentY = this.y;
      this.sendTimer = 0;
    }

    return { x: this.x, y: this.y, moved: moved && this.sendTimer === 0 };
  }

  setPosition(x: number, y: number) {
    this.x = x;
    this.y = y;
    this.container.setPosition(x, y);
    this.nameTag.setPosition(x, y - 28);
    this.statusIcon.setPosition(x + 18, y - 12);
  }

  setMicState(on: boolean) {
    this.statusIcon.setText(on ? "🎤" : "");
  }

  destroy() {
    this.container.destroy();
    this.nameTag.destroy();
    this.statusIcon.destroy();
  }
}
