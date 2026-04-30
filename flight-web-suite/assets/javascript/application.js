import { Application } from "@hotwired/stimulus"

import "./dropdown";

// Import stimulus controllers. Each entry here needs manually registering
// below.
import NovncController from "./controllers/novnc_controller"

window.Stimulus = Application.start()
Stimulus.register("novnc", NovncController)
