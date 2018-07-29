"use strict";
let ws = new WebSocket(`ws://${location.host}/ws`);
let sensor = Sensor();
let dataNode = getRenderTarget("data-target");
let handler = {
    onMessage(evt) {
        sensor.update(evt.data);
        dataNode.innerHTML = JSON.stringify(sensor, null, 2);
    },
    onClose(evt) {
        alert("WebSocket connection closed");
        console.warn("ws:close", evt);
    },
    onError(evt) {
        alert("WebSocket Error");
        console.error(evt);
    },
};
ws.addEventListener("message", handler.onMessage);
ws.addEventListener("close", handler.onClose);
ws.addEventListener("error", handler.onError);
function getRenderTarget(id) {
    let node = document.getElementById(id);
    if (!node) {
        throw new Error();
    }
    return node;
}
function Sensor() {
    let x = { x: 0, y: 0, z: 0 };
    return {
        update(y) {
            Object.assign(x, JSON.parse(y));
        },
        toJSON() {
            return x;
        },
        value() {
            return x;
        },
    };
}
//# sourceMappingURL=main.js.map