import { Controller } from "@hotwired/stimulus"
import RFB from '@novnc/novnc'

function noop() {}

export default class extends Controller {
  static targets = [ "status", "canvas", "previewImage", "reconnectBtn", "disconnectBtn", "copyFallback" ]
  static values = {
    host: String,
    port: Number,
    path: String,
    password: String,
    viewOnly: { type: Boolean, default: false },
    // One of `connecting`, `connected`, `disconnected`.
    status: { type: String, default: "connecting" },
  }

  connect() {
    this.url = this.buildUrl()
    console.log("Initialising NoVNC connection", this.url)
    this.createNoVncConnection()
  }

  buildUrl() {
    let url;
    if (window.location.protocol === "https:") {
      url = 'wss';
    } else {
      url = 'ws';
    }
    url += '://' + window.location.hostname
    url += ':' + window.location.port;
    url += '/' + this.pathValue;
    url += '?host=' + this.hostValue
    url += '&port=' + this.portValue
    return url
  }

  // Callbacks for RFB.
  onRfbStatusChange(newValue, oldValue) {
    console.log('RFB status change:', oldValue, "->", newValue)
    if (newValue != null && newValue.type === "connect") {
      this.rfb.focus();
      this.statusValue = "connected"
    }
  }

  onRfbDisconnect(e) {
    this.statusValue = "disconnected"
  }

  onRfbPasswordRequired() {
    console.log("Providing password")
    // The password is provided when constructing the RFB instance, but if its
    // required again FSR (re-connection perhaps), we do so here.
    this.rfb.sendCredentials({password: this.passwordValue})
  }

  async onRfbClipboard(ev) {
    if (ev && ev.detail && ev.detail.text) {
      try {
        await navigator.clipboard.write([
          new ClipboardItem({
            'text/plain': new Blob([ev.detail.text], { type: 'text/plain' }),
          })
        ]);
      } catch (err) {
        console.log('NoVNC clipboard error:', err)
        try {
          // Some browsers might not support the ClipboardItem API or have it
          // disabled.  This is a fallback to support those browsers.
          //
          // The textarea must be visible in order to copy from it
          this.copyFallbackTarget.current.value = ev.detail.text;
          this.copyFallbackTarget.current.style.display = 'block';
          this.copyFallbackTarget.current.select();

          if (document.execCommand('copy')) {
            console.log("Copy succeeded!");
          } else {
            console.log("Copy failed!");
          }
        } catch(fallbackErr) {
          console.log('err:', fallbackErr);  // eslint-disable-line no-console
        }
        this.copyFallbackTarget.current.style.display = 'none';
        this.copyFallbackTarget.current.value = '';
      }
    }
  }

  statusValueChanged(neww, oldd) {
    if (this.statusValue === "connected") {
      setTimeout(() => {
        this.statusTarget.innerText = this.statusText()
        this.reconnectBtnTarget.classList.add("hidden")
        this.disconnectBtnTarget.classList.remove("hidden")
        this.previewImageTarget.classList.add("hidden")
        this.canvasTarget.classList.remove("hidden")
      }, 750)
    } else {
      this.reconnectBtnTarget.classList.remove("hidden")
      this.disconnectBtnTarget.classList.add("hidden")
      this.previewImageTarget.classList.remove("hidden")
      this.canvasTarget.classList.add("hidden")
      this.statusTarget.innerText = this.statusText()
    }
  }

  statusText() {
    switch (this.statusValue) {
      case "connecting":
        return "Connecting..."
      case "connected":
        return "Connected"
      case "disconnected":
        return "Disconnected"
    }
  }

  // Call this to force a reconnection.
  reconnect() {
    this.statusValue = "connecting"
    this.createNoVncConnection();
  }

  // Call this to force a disconnection.
  disconnect() {
    this.rfb.disconnect()
  }

  createNoVncConnection() {
    console.log('creating RFB connection');
    if (this.rfb != null && this.statusValue !== "disconnected") {
      this.rfb.disconnect();
      this.rfb = null;
    }
    this.statusValue = "connecting"
    this.rfb = createConnection({
      url: this.url,
      domEl: this.canvasTarget,
      onClipboard: this.onRfbClipboard.bind(this),
      onDisconnect: this.onRfbDisconnect.bind(this),
      onConnect: this.onRfbStatusChange.bind(this),
      onPasswordPrompt: this.onRfbPasswordRequired.bind(this),
      password: this.passwordValue,
      viewOnly: this.viewOnlyValue
    });
  }
}

function createConnection({
  url,
  domEl,
  onClipboard=noop,
  onDisconnect=noop,
  onConnect=noop,
  onPasswordPrompt=noop,
  password,
  viewOnly,
}) {
  let rfb = null;
  try {
    console.log('RFB connecting to', url);
    rfb = new RFB(
      domEl,
      url,
      password && {credentials: {password}}
    );
    // Don't set a background.
    rfb.background = ""
    rfb.addEventListener('connect', onConnect);
    rfb.addEventListener('disconnect', onDisconnect);
    rfb.addEventListener('credentialsrequired', onPasswordPrompt);
    rfb.addEventListener('clipboard', onClipboard);
    rfb.scaleViewport = true;
    rfb.resizeSession = false;
    rfb.viewOnly = viewOnly;
  } catch (err) {
    console.error(`Unable to create RFB client: ${err}`)
    return onDisconnect({detail: {clean: false}})
  }

  return rfb
}
