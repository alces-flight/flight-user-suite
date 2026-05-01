import { Controller } from "@hotwired/stimulus"
import { enter, leave } from "el-transition"


// Connects to data-controller="menu-toggle"
export default class extends Controller {
    static targets = ["openIcon", "closeIcon", "menu", "trigger"];
    static values = { isOpen: { type: Boolean, default: false } }
    static classes = ["active"];

    isOpenValueChanged(isOpen) {
        if (isOpen) {
            if (this.hasOpenIconTarget && this.hasCloseIconTarget) {
                leave(this.openIconTarget).then(() => { enter(this.closeIconTarget); });
            }
            if (this.hasOpenIconTarget && this.hasActiveClass) {
                this.openIconTarget.classList.add(this.activeClass);
            }
            if (this.hasMenuTarget) {
                enter(this.menuTarget);
            }
        } else {
            if (this.hasOpenIconTarget && this.hasCloseIconTarget) {
                leave(this.closeIconTarget).then(() => { enter(this.openIconTarget); });
            }
            if (this.hasOpenIconTarget && this.hasActiveClass) {
                this.openIconTarget.classList.remove(this.activeClass);
            }
            if (this.hasMenuTarget) {
                leave(this.menuTarget);
            }
        }
    }

    toggle(evt) {
        this.isOpenValue = !this.isOpenValue;
        if (!this.hasTriggerTarget) {
            evt.stopPropagation();
        }
    }

    close(evt) {
        if (!this.hasTriggerTarget || !this.triggerTarget.contains(evt.target)) {
            // Ignore if this click came from us or our descendants (it would
            // have been handled in #toggle)
            this.isOpenValue = false;
        }
    }
}
