function connect(location, onmessage) {
    const url = "ws://" + window.location.host + window.location.pathname + "/" + location;
    let ws = new WebSocket(url);

    ws.onmessage = (msg) => {
        onmessage(msg)
    };
    return ws
}
   
    
   
