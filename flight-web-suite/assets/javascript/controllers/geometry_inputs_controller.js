import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  static targets = [ "geometrySelect", "customFields" ]
  static values = { showCustom: { type: Boolean, default: false } }

  initialize() {
    this.updateVisibility = this.updateVisibility.bind(this)
  }

  geometrySelectTargetConnected(elem) {
    elem.addEventListener('change', this.updateVisibility);
  }

  geometrySelectTargetDisconnected(elem) {
    elem.removeEventListener('change', this.updateVisibility);
  }

  updateVisibility() {
    this.showCustomValue = this.geometrySelectTarget.value == "custom";
  }

  showCustomValueChanged(newValue) {
    if (newValue) {
      this.customFieldsTarget.classList.remove('hidden')
    }
    else {
      this.customFieldsTarget.classList.add('hidden')
    }
  }
}
