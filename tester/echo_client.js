const WebSocket = require("ws");

const ws = new WebSocket("ws://127.0.0.1:8080/ws");

ws.on("open", () => {
  console.log("Connected to the WebSocket server");
});

ws.on("message", (message) => {
  console.log("Received:", message);
  ws.send(message); // Echo the received message
});

ws.on("error", (error) => {
  console.error("WebSocket error:", error);
});

ws.on("close", () => {
  console.log("Connection closed");
});
