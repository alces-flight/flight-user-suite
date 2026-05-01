import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = ["fullscreenable"]
  static values = { isFullscreen: { type: Boolean, default: false } }

  initialize() {
    this.updateFromDocument = this.updateFromDocument.bind(this)
  }

  connect() {
    document.addEventListener('fullscreenchange', this.updateFromDocument)
  }

  disconnect() {
    document.removeEventListener('fullscreenchange', this.updateFromDocument)
  }

  isFullscreenValueChanged(isFullscreen) {
    if (isFullscreen) {
      this.fullscreenableTarget.requestFullscreen().catch(
        () => {
          // Fullscreening failed FSR
          this.isFullscreenValue = false
        }
      )
    }
    else if (this.fullscreenableTarget === document.fullscreenElement) {
      // Not sure how this would be triggered given the fullscreen VNC canvas
      // covers the button, but for completeness:
      document.exitFullscreen()
    }
  }

  toggle() {
    this.isFullscreenValue = !this.isFullscreenValue
  }

  updateFromDocument() {
    this.isFullscreenValue = this.fullscreenableTarget === document.fullscreenElement
  }
}
