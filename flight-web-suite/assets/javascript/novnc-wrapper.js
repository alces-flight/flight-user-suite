import RFB from '@novnc/novnc'

function noop() {}

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
    console.log('connecting to', url);
    rfb = new RFB(
      domEl,
      url,
      password && {credentials: {password}}
    );
    rfb.addEventListener('connect', onConnect);
    rfb.addEventListener('disconnect', onDisconnect);
    rfb.addEventListener('credentialsrequired', onPasswordPrompt);
    rfb.addEventListener('clipboard', onClipboard);
    // TODO: Add this back. Requires set height and width for domEl (not just 100%).
    // rfb.scaleViewport = true;
    rfb.resizeSession = false;
    rfb.viewOnly = viewOnly;
  } catch (err) {
    console.error(`Unable to create RFB client: ${err}`)
    return onDisconnect({detail: {clean: false}})
  }

  return rfb
}

class NoVncContainer {
  constructor(noVNCCanvas, url, password, copyFallback=noop, viewOnly=false, onDisconnected=noop, onBeforeConnect=noop) {
    this.url = url
    this.password = password
    this.viewOnly = viewOnly
    this.onDisconnected = onDisconnected
    this.onBeforeConnect = onBeforeConnect

    this.noVNCCanvas = noVNCCanvas
    this.copyFallback = copyFallback

    this.status = "connecting"
  }

  onStatusChange = () => {
    this.rfb.focus();
    this.status = "connected"
    this.onBeforeConnect();
  };

  onClipboard = async (ev) => {
    if (ev && ev.detail && ev.detail.text) {
      try {
        await navigator.clipboard.write([
          // eslint-disable-next-line no-undef
          new ClipboardItem({
            'text/plain': new Blob([ev.detail.text], { type: 'text/plain' }),
          })
        ]);
      } catch (err) {
        console.log('err:', err);  // eslint-disable-line no-console
        try {
          // Firefox does not support the ClipboardItem API at the time or writing.
          // It should *hopefully* support it in the future, so it is still attempted.
          // Until then, a fallback on a text area is used.
          //
          // The textarea must be visible in order to copy from it
          this.copyFallback.current.value = ev.detail.text;
          this.copyFallback.current.style.display = 'block';
          this.copyFallback.current.select();

          if (document.execCommand('copy')) {
            console.log("Copy succeeded!");
          } else {
            console.log("Copy failed!");
          }
        } catch(fallback_err) {
          console.log('err:', fallback_err);  // eslint-disable-line no-console
        }
        this.copyFallback.current.style.display = 'none';
        this.copyFallback.current.value = '';
      }
    }
  }

  onDisconnect = (e) => {
    this.onDisconnected(
      !e.detail.clean || this.status !== 'connected'
    )
    this.status = 'disconnected'
  }

  onUserDisconnect = () => this.rfb.disconnect();

  onPasswordRequired = () => {
    console.log("providing password")
    // XXX Something has gone all kinds of wrong here.  We should have
    // configured with the password in the first instance.
    this.rfb.sendCredentials({password: this.password})
  }

  onReconnect() {
    this.createConnection();
  }

  createConnection() {
    console.log('creating RFB connection');
    if (this.rfb != null && this.status !== 'disconnected') {
      this.rfb.disconnect();
      this.rfb = null;
    }
    this.rfb = createConnection({
      url: this.url,
      domEl: this.noVNCCanvas,
      onClipboard: this.onClipboard,
      onDisconnect: this.onDisconnect,
      onConnect: this.onStatusChange,
      onPasswordPrompt: this.onPasswordRequired,
      password: this.password,
      viewOnly: this.viewOnly
    });
  }

  resize() {
    if (this.rfb != null && typeof this.rfb._windowResize === 'function') {
      this.rfb._windowResize();
    }
  }

  setClipboardText(text) {
    if (this.rfb != null) {
      this.rfb.clipboardPasteFrom(text);
    }
  }
}

function readDataVariable(el, name, defaultValue) {
  return el.dataset[`novnc${name}`] || defaultValue
}

function initNoVnc() {
  const domEl = document.getElementById("novnc-wrapper")
  if (domEl == null) {
    return
  }
  console.log("initialising NoVNC")

  // Read parameters specified in the URL query string
  // By default, use the host and port of server that served this file
  const host = readDataVariable(domEl, 'Host', window.location.hostname);
  const port = readDataVariable(domEl, 'Port', window.location.port);
  const password = readDataVariable(domEl, 'Password');
  const path = readDataVariable(domEl, 'Path', 'websockify');

  // Build the websocket URL used to connect
  let url;
  if (window.location.protocol === "https:") {
    url = 'wss';
  } else {
    url = 'ws';
  }
  url += '://' + host;
  if(port) {
    url += ':' + port;
  }
  url += '/' + path;

  const vnc = new NoVncContainer(
    document.getElementById("novnc-canvas"),
    url,
    password,
    document.getElementById("novnc-canvas"),
  )
  vnc.createConnection()
}

document.addEventListener("DOMContentLoaded", initNoVnc)
