"use client";

import type { FloorData, RemotePlayerData } from "./phaser/config";
import type { GameCallbacks } from "./phaser/scenes/GameScene";

export class WorldEngine {
  private game: any = null;
  private scene: any = null;
  private Phaser: any = null;
  private mounted = false;

  async mount(containerId: string, callbacks: GameCallbacks) {
    if (typeof window === "undefined") return;

    const Phaser = await import("phaser");
    this.Phaser = Phaser.default;

    const { BootScene } = await import("./phaser/scenes/BootScene");
    const { GameScene } = await import("./phaser/scenes/GameScene");

    const config: Phaser.Types.Core.GameConfig = {
      type: Phaser.default.AUTO,
      parent: containerId,
      width: "100%",
      height: "100%",
      backgroundColor: "#0a0a0f",
      scale: {
        mode: Phaser.default.Scale.RESIZE,
        autoCenter: Phaser.default.Scale.CENTER_BOTH,
      },
      physics: {
        default: "arcade",
        arcade: { gravity: { x: 0, y: 0 }, debug: false },
      },
      scene: [BootScene, GameScene],
      pixelArt: true,
      roundPixels: true,
    };

    this.game = new Phaser.default.Game(config);

    const waitForScene = () => {
      const scene = this.game?.scene.getScene("GameScene");
      if (scene && scene.scene.isActive()) {
        this.scene = scene;
        this.scene.callbacks = callbacks;
        this.mounted = true;
        callbacks.onReady?.();
      } else {
        requestAnimationFrame(waitForScene);
      }
    };
    waitForScene();
  }

  unmount() {
    if (this.game) {
      this.game.destroy(true);
      this.game = null;
      this.scene = null;
      this.mounted = false;
    }
  }

  addRemotePlayer(data: RemotePlayerData) {
    this.scene?.addRemotePlayer(data);
  }

  updateRemotePlayer(data: RemotePlayerData) {
    this.scene?.updateRemotePlayer(data);
  }

  removeRemotePlayer(id: string) {
    this.scene?.removeRemotePlayer(id);
  }

  showRemoteEmote(playerId: string, emoji: string) {
    this.scene?.showRemoteEmote(playerId, emoji);
  }

  setFloors(floors: FloorData[]) {
    this.scene?.setFloors(floors);
  }

  jumpToPosition(x: number, y: number) {
    this.scene?.jumpToPosition(x, y);
  }

  setMiniMode(enabled: boolean) {
    this.scene?.setMiniMode(enabled);
  }

  getPlayerPosition(): { x: number; y: number } {
    return this.scene?.getPlayerPosition() ?? { x: 200, y: 200 };
  }

  getCurrentZone(): string | null {
    return this.scene?.getZoneSystem()?.getCurrentZone() ?? null;
  }

  setMicState(on: boolean) {
    this.scene?.setMicState(on);
  }

  showLocalEmote(emoji: string) {
    const pos = this.getPlayerPosition();
    this.scene?.getEmoteSystem()?.showEmote(pos.x, pos.y, emoji);
  }
}

export const worldEngine = new WorldEngine();
