import Phaser from "phaser";
import { TILE_SIZE } from "../config";

export class BootScene extends Phaser.Scene {
  constructor() {
    super({ key: "BootScene" });
  }

  preload() {
    this.generateTileTextures();
    this.generateAvatarTextures();
    this.generateObjectTextures();
  }

  create() {
    this.scene.start("GameScene");
  }

  private generateTileTextures() {
    const g = this.add.graphics();

    g.fillStyle(0x1a1a2e);
    g.fillRect(0, 0, TILE_SIZE, TILE_SIZE);
    g.lineStyle(1, 0x16213e, 0.3);
    g.strokeRect(0, 0, TILE_SIZE, TILE_SIZE);
    g.generateTexture("tile-floor", TILE_SIZE, TILE_SIZE);
    g.clear();

    g.fillStyle(0x0f3460);
    g.fillRect(0, 0, TILE_SIZE, TILE_SIZE);
    g.lineStyle(1, 0x533483, 0.5);
    g.strokeRect(0, 0, TILE_SIZE, TILE_SIZE);
    g.generateTexture("tile-wall", TILE_SIZE, TILE_SIZE);
    g.clear();

    g.fillStyle(0x1a1a2e);
    g.fillRect(0, 0, TILE_SIZE, TILE_SIZE);
    g.lineStyle(1, 0xe94560, 0.15);
    g.strokeRect(0, 0, TILE_SIZE, TILE_SIZE);
    g.generateTexture("tile-grid", TILE_SIZE, TILE_SIZE);
    g.clear();

    g.destroy();
  }

  private generateAvatarTextures() {
    const size = 32;
    const g = this.add.graphics();

    const colors = [0xa855f7, 0x06b6d4, 0xec4899, 0x10b981, 0xf59e0b, 0x3b82f6, 0xef4444, 0x8b5cf6];
    colors.forEach((color, i) => {
      g.clear();
      g.fillStyle(color, 0.2);
      g.fillCircle(size / 2, size / 2, size / 2);
      g.fillStyle(color, 0.6);
      g.fillCircle(size / 2, size / 2, size / 3);
      g.fillStyle(color);
      g.fillCircle(size / 2, size / 2 - 2, 5);
      g.fillStyle(0xffffff);
      g.fillCircle(size / 2, size / 2 - 2, 2);
      g.generateTexture(`avatar-${i}`, size, size);
    });

    g.clear();
    g.fillStyle(0xffffff, 0.15);
    g.fillCircle(size / 2, size / 2, size / 2);
    g.lineStyle(2, 0xffffff, 0.4);
    g.strokeCircle(size / 2, size / 2, size / 2 - 2);
    g.generateTexture("avatar-me", size, size);

    g.destroy();
  }

  private generateObjectTextures() {
    const g = this.add.graphics();

    const deskW = 48;
    const deskH = 32;
    g.fillStyle(0x2d1b69);
    g.fillRoundedRect(0, 0, deskW, deskH, 4);
    g.lineStyle(1, 0xa855f7, 0.3);
    g.strokeRoundedRect(0, 0, deskW, deskH, 4);
    g.generateTexture("desk", deskW, deskH);
    g.clear();

    g.fillStyle(0x10b981);
    g.fillRoundedRect(0, 0, deskW, deskH, 4);
    g.lineStyle(1, 0x34d399, 0.5);
    g.strokeRoundedRect(0, 0, deskW, deskH, 4);
    g.generateTexture("desk-claimed", deskW, deskH);
    g.clear();

    g.fillStyle(0x7c3aed, 0.15);
    g.fillCircle(12, 12, 12);
    g.fillStyle(0x7c3aed);
    g.fillTriangle(9, 7, 9, 17, 17, 12);
    g.generateTexture("icon-play", 24, 24);
    g.clear();

    g.fillStyle(0xef4444, 0.15);
    g.fillCircle(12, 12, 12);
    g.lineStyle(2, 0xef4444);
    g.strokeRect(8, 8, 8, 8);
    g.generateTexture("icon-lock", 24, 24);
    g.clear();

    g.destroy();
  }
}
