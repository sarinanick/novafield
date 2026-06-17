import Phaser from "phaser";
import { WORLD_WIDTH, WORLD_HEIGHT, FloorData, RemotePlayerData } from "../config";
import { Player } from "../entities/Player";
import { RemotePlayer } from "../entities/RemotePlayer";
import { CameraManager } from "../systems/CameraManager";
import { ZoneSystem } from "../systems/ZoneSystem";
import { DeskSystem } from "../systems/DeskSystem";
import { ProximitySystem } from "../systems/ProximitySystem";
import { MapManager } from "../maps/MapManager";
import { EmoteSystem } from "../ui/EmoteSystem";

export interface GameCallbacks {
  onMove?: (x: number, y: number) => void;
  onZoneChange?: (zoneId: string | null) => void;
  onDeskClick?: (deskId: string) => void;
  onDeskHover?: (deskId: string | null) => void;
  onEmote?: (emoji: string) => void;
  onProximityEnter?: (playerId: string) => void;
  onProximityLeave?: (playerId: string) => void;
  onReady?: () => void;
}

export class GameScene extends Phaser.Scene {
  private player!: Player;
  private remotePlayers: Map<string, RemotePlayer> = new Map();
  private cameraManager!: CameraManager;
  private zoneSystem!: ZoneSystem;
  private deskSystem!: DeskSystem;
  private proximitySystem!: ProximitySystem;
  private mapManager!: MapManager;
  private emoteSystem!: EmoteSystem;
  callbacks: GameCallbacks = {};
  private miniMode = false;
  private miniContainer: Phaser.GameObjects.Container | null = null;

  constructor() {
    super({ key: "GameScene" });
  }

  init(data: { callbacks?: GameCallbacks }) {
    this.callbacks = data.callbacks ?? {};
  }

  create() {
    this.mapManager = new MapManager(this);
    this.zoneSystem = new ZoneSystem(this, this.mapManager.getZones(), (zoneId) => {
      this.callbacks.onZoneChange?.(zoneId);
    });

    this.deskSystem = new DeskSystem(this, this.mapManager.getDesks(), (deskId) => {
      this.callbacks.onDeskClick?.(deskId);
    });

    this.player = new Player(this, 200, 200, "You", 0);

    this.cameraManager = new CameraManager(this);
    this.cameraManager.recenter(200, 200);

    this.proximitySystem = new ProximitySystem(this, {
      onEnterProximity: (id) => this.callbacks.onProximityEnter?.(id),
      onLeaveProximity: (id) => this.callbacks.onProximityLeave?.(id),
    });

    this.emoteSystem = new EmoteSystem(this, (emoji) => {
      this.callbacks.onEmote?.(emoji);
    });

    this.setupKeyboard();
    this.callbacks.onReady?.();
  }

  private setupKeyboard() {
    if (!this.input.keyboard) return;

    this.input.keyboard.on("keydown-E", () => {
      const deskId = this.deskSystem.getHoveredDesk();
      if (deskId) {
        this.callbacks.onDeskClick?.(deskId);
      }
    });

    this.input.keyboard.on("keydown-Q", () => {
      this.emoteSystem.toggleQuickBar(this.player.x, this.player.y);
    });

    this.input.keyboard.on("keydown-ONE", () => this.jumpToZone(0));
    this.input.keyboard.on("keydown-TWO", () => this.jumpToZone(1));
    this.input.keyboard.on("keydown-THREE", () => this.jumpToZone(2));
    this.input.keyboard.on("keydown-FOUR", () => this.jumpToZone(3));

    this.input.keyboard.on("keydown-PLUS", () => this.cameraManager.zoomIn());
    this.input.keyboard.on("keydown-MINUS", () => this.cameraManager.zoomOut());
  }

  private jumpToZone(index: number) {
    const zones = this.mapManager.getZones();
    if (zones[index]) {
      const z = zones[index];
      this.player.setPosition(z.x + z.w / 2, z.y + z.h / 2);
      this.cameraManager.recenter(this.player.x, this.player.y);
      this.callbacks.onMove?.(this.player.x, this.player.y);
    }
  }

  update(_time: number, delta: number) {
    const result = this.player.update(delta);

    if (result.moved) {
      this.callbacks.onMove?.(result.x, result.y);
    }

    this.remotePlayers.forEach(rp => rp.update(delta));

    this.cameraManager.follow(this.player.x, this.player.y);
    this.cameraManager.update();

    this.zoneSystem.checkZone(this.player.x, this.player.y);

    const remoteData = new Map<string, RemotePlayerData>();
    this.remotePlayers.forEach((rp, id) => {
      remoteData.set(id, { id, name: "", x: rp.container.x, y: rp.container.y });
    });
    this.proximitySystem.update(this.player.x, this.player.y, remoteData);
  }

  addRemotePlayer(data: RemotePlayerData) {
    if (this.remotePlayers.has(data.id)) return;
    const rp = new RemotePlayer(this, data);
    this.remotePlayers.set(data.id, rp);
  }

  updateRemotePlayer(data: RemotePlayerData) {
    const rp = this.remotePlayers.get(data.id);
    if (rp) {
      rp.updateFromData(data);
    }
  }

  removeRemotePlayer(id: string) {
    const rp = this.remotePlayers.get(id);
    if (rp) {
      rp.destroy();
      this.remotePlayers.delete(id);
    }
  }

  showRemoteEmote(playerId: string, emoji: string) {
    const rp = this.remotePlayers.get(playerId);
    if (rp) rp.showEmote(emoji);
  }

  setFloors(floors: FloorData[]) {
    this.mapManager.setFloors(floors);
    this.zoneSystem.updateZones(this.mapManager.getZones());
    this.deskSystem.updateDesks(this.mapManager.getDesks());
  }

  jumpToPosition(x: number, y: number) {
    this.player.setPosition(x, y);
    this.cameraManager.recenter(x, y);
    this.callbacks.onMove?.(x, y);
  }

  setMiniMode(enabled: boolean) {
    this.miniMode = enabled;
    if (enabled) {
      this.cameraManager.zoomOut();
      this.cameraManager.zoomOut();
      this.cameraManager.zoomOut();
    } else {
      this.cameraManager.recenter(this.player.x, this.player.y);
    }
  }

  getPlayerPosition(): { x: number; y: number } {
    return { x: this.player.x, y: this.player.y };
  }

  getZoneSystem(): ZoneSystem {
    return this.zoneSystem;
  }

  getEmoteSystem(): EmoteSystem {
    return this.emoteSystem;
  }

  setMicState(on: boolean) {
    this.player.setMicState(on);
  }
}
