function connect(onmessage) {
    const url = "ws://" + window.location.host + window.location.pathname + "/WS";
    let ws = new WebSocket(url);

    ws.onmessage = (msg) => {
        onmessage(msg)
    };
    return ws
}
    
   
