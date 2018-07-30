let ws = new WebSocket(`ws://${location.host}/ws`);
let sensor = Sensor();
let dataNode = getRenderTarget("data-target");

let wsHandler = {
  onMessage(evt: MessageEvent) {
    sensor.update(evt.data);
    dataNode.innerHTML = JSON.stringify(sensor, null, 2);
  },
  onClose(evt: CloseEvent) {
    alert("WebSocket connection closed");
    console.warn("ws:close", evt);
  },
  onError(evt: Event) {
    alert("WebSocket Error");
    console.error(evt);
  },
};

ws.addEventListener("message", wsHandler.onMessage);
ws.addEventListener("close", wsHandler.onClose);
ws.addEventListener("error", wsHandler.onError);

function getRenderTarget(id: string) {
  let node = document.getElementById(id);
  if (!node) {
    throw new Error(`missing HTMLElement #${id}`);
  }

  return node;
}

interface SensorData {
  [x: string]: any;
}

interface SensorMut {
  update(json: string): void;
  copy(): SensorMut;
  toJSON(): SensorData;
}

function Sensor(d?: SensorData): SensorMut {
  let x: SensorData = { ...d };

  return {
    update(y) {
      Object.assign(x, JSON.parse(y));
    },
    toJSON() {
      return x;
    },
    copy() {
      return Sensor(x);
    },
  };
}
