function connect (location, onmessage){
    var url = "ws://" + window.location.host + "/" + location;
    var ws = new WebSocket(url);

    ws.onmessage = (msg) => {
       onmessage(msg)
    }
    return ws
}
   
    
   
