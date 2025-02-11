const socket = new WebSocket("ws://localhost:8080/ws");

socket.onmessage = (event) => {
  const message = event.data;
  console.log(`Received reload message: ${message}`);

  if (message === "css") {
    document.querySelectorAll("link[rel='stylesheet']").forEach((link) => {
      const newLink = link.cloneNode();
      newLink.href = link.href.split("?")[0] + "?v=" + new Date().getTime();
      link.parentNode.replaceChild(newLink, link);
    });
    console.log("CSS hot reloaded!");
  } else if (message === "js") {
    document.querySelectorAll("script").forEach((script) => {
      const newScript = document.createElement("script");
      newScript.src = script.src.split("?")[0] + "?v=" + new Date().getTime();
      newScript.async = script.async;
      script.parentNode.replaceChild(newScript, script);
    });
    console.log("JavaScript hot reloaded!");
  } else if (message === "reload") {
    window.location.reload();
  }
};

socket.onopen = () => console.log("WebSocket connected");
socket.onclose = () => console.log("WebSocket disconnected");
