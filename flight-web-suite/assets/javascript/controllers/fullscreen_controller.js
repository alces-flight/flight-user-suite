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
  }

  toggle() {
    this.isFullscreenValue = !this.isFullscreenValue
  }

  updateFromDocument() {
    this.isFullscreenValue = this.fullscreenableTarget === document.fullscreenElement
  }
}
