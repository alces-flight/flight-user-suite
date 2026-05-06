import { Controller } from "@hotwired/stimulus"

export default class extends Controller {
  connect() {
    const button = this.element
    if (button.dataset.disableWith) {
      button.addEventListener('click', () => {
        // Delay actually disabling the button until the form submission has
        // gone through.
        setTimeout(
          () => { button.disabled = true },
          1
        )

        button.innerText = button.dataset.disableWith
        if (button.dataset.disableWithClass) {
          button.classList.add(button.dataset.disableWithClass)
        }
      })
    }
  }
}
