import { Application } from "@hotwired/stimulus"

import "./dropdown";

// Import stimulus controllers. Each entry here needs manually registering
// below.
import DisableWithController from "./controllers/disable_with_controller.js"
import NovncController from "./controllers/novnc_controller"
import MenuToggleController from "./controllers/menu_toggle_controller.js"
import RefreshController from "./controllers/refresh_controller.js"

window.Stimulus = Application.start()
Stimulus.register("disable-with", DisableWithController)
Stimulus.register("novnc", NovncController)
Stimulus.register("menu-toggle", MenuToggleController)
Stimulus.register("refresh", RefreshController)
