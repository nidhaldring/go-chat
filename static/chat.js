// @ts-check
const chatId = window.location.href.split("/").pop();
const inputText = /** @type  {HTMLInputElement} */ (
  document.getElementById("text")
);
const chatterName = /** @type  {HTMLHeadingElement} */ (
  document.getElementById("name")
);
const eventTypes = {
  init: "init",
  message: "message",
  status: "status",
};

/**
 * @param text {string}
 */
const appendTextToChat = (text) => {
  const chatArea = /** @type  {HTMLElement} */ (
    document.getElementById("chat")
  );
  chatArea.innerHTML = chatArea.innerHTML + "\n" + text;
};

const clearChatInput = () => {
  inputText.value = "";
};

/**
 * @param msg {string} will be a comma separated list
 */
const updateStatus = (msg) => {
  const status = /** @type  {HTMLHeadingElement} */ (
    document.getElementById("status")
  );
  status.innerHTML = msg
    .split(",")
    .map((m) => m[0].toUpperCase() + m.slice(1))
    .join("<br/>");
};

const socket = new WebSocket(`ws://localhost:5000/chat/connect/${chatId}`);
socket.addEventListener("message", (event) => {
  const { msg, sender, type } = JSON.parse(event.data);

  switch (type) {
    case eventTypes.init:
      chatterName.innerHTML = `Hi <strong>${sender}</strong>, type something!`;
      break;
    case eventTypes.message:
      appendTextToChat(`${sender}: ${msg}`);
      break;
    case eventTypes.status:
      updateStatus(msg);
      break;
  }
});

const btn = /** @type {HTMLButtonElement} */ (
  document.getElementById("sendBtn")
);
btn.onclick = () => {
  socket.send(inputText.value);
  appendTextToChat(`me: ${inputText.value}`);
  clearChatInput();
};
