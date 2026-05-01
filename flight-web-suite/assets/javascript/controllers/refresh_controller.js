import { Controller } from "@hotwired/stimulus"

// Periodically refreshes an image tag.
//
// Connects to data-controller="refresh"
export default class extends Controller {
  static values = { interval: { type: Number, default: 60 * 1000 } }

  connect() {
    this.initialSrc = this.element.src
    this.intervalId = setInterval(this.refresh.bind(this), this.intervalValue)
  }

  disconnect() {
    clearInterval(this.intervalId)
  }

  refresh() {
    const cacheBuster = new Date().getTime()
    this.element.src = `${this.initialSrc}?t=${cacheBuster}`
  }
}
