import Phaser from "phaser";
import { WORLD_WIDTH, WORLD_HEIGHT } from "../config";

const MIN_ZOOM = 0.5;
const MAX_ZOOM = 2;
const ZOOM_SPEED = 0.1;
const LERP_SPEED = 0.08;

export class CameraManager {
  private camera: Phaser.Cameras.Scene2D.Camera;
  private targetX = 0;
  private targetY = 0;
  private targetZoom = 1;
  private isDragging = false;
  private dragStartX = 0;
  private dragStartY = 0;
  private cameraStartX = 0;
  private cameraStartY = 0;
  private following = true;

  constructor(scene: Phaser.Scene) {
    this.camera = scene.cameras.main;
    this.camera.setBounds(0, 0, WORLD_WIDTH, WORLD_HEIGHT);
    this.camera.setZoom(1.5);

    scene.input.on("wheel", (_pointer: any, _go: any, _dx: number, dy: number) => {
      this.targetZoom = Phaser.Math.Clamp(
        this.targetZoom - Math.sign(dy) * ZOOM_SPEED,
        MIN_ZOOM,
        MAX_ZOOM
      );
    });

    scene.input.on("pointerdown", (pointer: Phaser.Input.Pointer) => {
      if (pointer.rightButtonDown() || pointer.middleButtonDown()) {
        this.isDragging = true;
        this.following = false;
        this.dragStartX = pointer.x;
        this.dragStartY = pointer.y;
        this.cameraStartX = this.camera.scrollX;
        this.cameraStartY = this.camera.scrollY;
      }
    });

    scene.input.on("pointermove", (pointer: Phaser.Input.Pointer) => {
      if (this.isDragging) {
        const dx = (this.dragStartX - pointer.x) / this.camera.zoom;
        const dy = (this.dragStartY - pointer.y) / this.camera.zoom;
        this.camera.scrollX = this.cameraStartX + dx;
        this.camera.scrollY = this.cameraStartY + dy;
      }
    });

    scene.input.on("pointerup", () => {
      this.isDragging = false;
    });
  }

  follow(x: number, y: number) {
    if (!this.following) return;
    this.targetX = x;
    this.targetY = y;
  }

  update() {
    if (this.following) {
      this.camera.scrollX += (this.targetX - this.camera.width / 2 / this.camera.zoom - this.camera.scrollX) * LERP_SPEED;
      this.camera.scrollY += (this.targetY - this.camera.height / 2 / this.camera.zoom - this.camera.scrollY) * LERP_SPEED;
    }
    this.camera.zoom += (this.targetZoom - this.camera.zoom) * 0.1;
  }

  recenter(x: number, y: number) {
    this.following = true;
    this.targetX = x;
    this.targetY = y;
    this.camera.centerOn(x, y);
  }

  zoomIn() {
    this.targetZoom = Phaser.Math.Clamp(this.targetZoom + ZOOM_SPEED, MIN_ZOOM, MAX_ZOOM);
  }

  zoomOut() {
    this.targetZoom = Phaser.Math.Clamp(this.targetZoom - ZOOM_SPEED, MIN_ZOOM, MAX_ZOOM);
  }

  getBounds() {
    return {
      x: this.camera.scrollX,
      y: this.camera.scrollY,
      width: this.camera.width / this.camera.zoom,
      height: this.camera.height / this.camera.zoom,
    };
  }
}
