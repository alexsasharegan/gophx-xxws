"use strict";
let ws = new WebSocket(`ws://${location.host}/ws`);
let sensor = Sensor();
let dataNode = getRenderTarget("data-target");
let wsHandler = {
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
ws.addEventListener("message", wsHandler.onMessage);
ws.addEventListener("close", wsHandler.onClose);
ws.addEventListener("error", wsHandler.onError);
function getRenderTarget(id) {
    let node = document.getElementById(id);
    if (!node) {
        throw new Error(`missing HTMLElement #${id}`);
    }
    return node;
}
function Sensor(d) {
    let x = Object.assign({}, d);
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
//# sourceMappingURL=main.js.map