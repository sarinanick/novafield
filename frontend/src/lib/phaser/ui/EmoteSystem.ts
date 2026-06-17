import Phaser from "phaser";

const EMOTES = ["👋", "👍", "❤️", "😂", "🎉", "🤔", "✅", "🔥"];
const EMOTE_DURATION = 3000;

export class EmoteSystem {
  private scene: Phaser.Scene;
  private activeEmote: Phaser.GameObjects.Text | null = null;
  private emoteTimer: number | null = null;
  private quickBar: Phaser.GameObjects.Container | null = null;
  private onSelect?: (emoji: string) => void;

  constructor(scene: Phaser.Scene, onSelect?: (emoji: string) => void) {
    this.scene = scene;
    this.onSelect = onSelect;
  }

  showEmote(x: number, y: number, emoji: string) {
    this.hideEmote();
    this.activeEmote = this.scene.add.text(x, y - 50, emoji, {
      fontSize: "28px",
      backgroundColor: "#000000cc",
      padding: { x: 10, y: 6 },
    }).setOrigin(0.5).setDepth(20);

    this.scene.tweens.add({
      targets: this.activeEmote,
      y: y - 65,
      alpha: { from: 0, to: 1 },
      duration: 300,
      ease: "Back.easeOut",
    });

    this.emoteTimer = window.setTimeout(() => this.hideEmote(), EMOTE_DURATION);
  }

  hideEmote() {
    if (this.activeEmote) {
      this.activeEmote.destroy();
      this.activeEmote = null;
    }
    if (this.emoteTimer !== null) {
      clearTimeout(this.emoteTimer);
      this.emoteTimer = null;
    }
  }

  toggleQuickBar(x: number, y: number) {
    if (this.quickBar) {
      this.quickBar.destroy();
      this.quickBar = null;
      return;
    }

    this.quickBar = this.scene.add.container(x, y - 80);
    this.quickBar.setDepth(25);

    const radius = 60;
    EMOTES.forEach((emoji, i) => {
      const angle = (i / EMOTES.length) * Math.PI * 2 - Math.PI / 2;
      const ex = Math.cos(angle) * radius;
      const ey = Math.sin(angle) * radius;

      const btn = this.scene.add.text(ex, ey, emoji, {
        fontSize: "20px",
        backgroundColor: "#1a1a2ecc",
        padding: { x: 6, y: 4 },
      }).setOrigin(0.5).setInteractive();

      btn.on("pointerdown", () => {
        this.onSelect?.(emoji);
        this.showEmote(x, y, emoji);
        this.quickBar?.destroy();
        this.quickBar = null;
      });

      btn.on("pointerover", () => {
        this.scene.tweens.add({ targets: btn, scale: 1.3, duration: 100 });
      });

      btn.on("pointerout", () => {
        this.scene.tweens.add({ targets: btn, scale: 1, duration: 100 });
      });

      this.quickBar!.add(btn);
    });
  }

  destroy() {
    this.hideEmote();
    if (this.quickBar) {
      this.quickBar.destroy();
      this.quickBar = null;
    }
  }
}
