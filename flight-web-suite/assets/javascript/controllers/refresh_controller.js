import { Controller } from "@hotwired/stimulus"

// Periodically refreshes an image tag.
//
// Connects to data-controller="refresh"
export default class extends Controller {
  static values = {
    interval: { type: Number, default: 60 * 1000 },
    src: String,
  }

  initialize() {
    this.refresh = this.refresh.bind(this)
  }

  connect() {
    this.intervalId = setInterval(this.refresh, this.intervalValue)
  }

  disconnect() {
    clearInterval(this.intervalId)
  }

  srcValueChanged() {
    const initial = this.element.src == null || this.element.src == ""
    if (initial && this.srcValue != null) {
      this.element.src = this.srcValue
    }
    setTimeout(this.refresh, 250)
  }

  refresh() {
    const cacheBuster = Date.now()
    this.element.src = `${this.srcValue}?t=${cacheBuster}`
  }
}
