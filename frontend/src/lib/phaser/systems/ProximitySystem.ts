import Phaser from "phaser";
import { PROXIMITY_RANGE, RemotePlayerData } from "../config";

export interface ProximityCallbacks {
  onEnterProximity?: (playerId: string) => void;
  onLeaveProximity?: (playerId: string) => void;
}

export class ProximitySystem {
  private scene: Phaser.Scene;
  private nearbyPlayers: Set<string> = new Set();
  private callbacks: ProximityCallbacks;
  private range: number;
  private rangeCircle: Phaser.GameObjects.Arc;
  private rangeGlow: Phaser.GameObjects.Arc;

  constructor(scene: Phaser.Scene, callbacks: ProximityCallbacks, range = PROXIMITY_RANGE) {
    this.scene = scene;
    this.callbacks = callbacks;
    this.range = range;

    this.rangeGlow = scene.add.circle(0, 0, range, 0xa855f7, 0.03);
    this.rangeGlow.setDepth(0);

    this.rangeCircle = scene.add.circle(0, 0, range, 0xa855f7, 0.0);
    this.rangeCircle.setStrokeStyle(1, 0xa855f7, 0.15);
    this.rangeCircle.setDepth(0);
  }

  update(playerX: number, playerY: number, remotePlayers: Map<string, RemotePlayerData>) {
    this.rangeCircle.setPosition(playerX, playerY);
    this.rangeGlow.setPosition(playerX, playerY);

    const currentNearby = new Set<string>();

    remotePlayers.forEach((data, id) => {
      const dx = playerX - data.x;
      const dy = playerY - data.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      if (dist < this.range) {
        currentNearby.add(id);
        if (!this.nearbyPlayers.has(id)) {
          this.callbacks.onEnterProximity?.(id);
        }
      }
    });

    this.nearbyPlayers.forEach(id => {
      if (!currentNearby.has(id)) {
        this.callbacks.onLeaveProximity?.(id);
      }
    });

    this.nearbyPlayers = currentNearby;
  }

  getNearbyPlayerIds(): string[] {
    return Array.from(this.nearbyPlayers);
  }

  destroy() {
    this.rangeCircle.destroy();
    this.rangeGlow.destroy();
  }
}
