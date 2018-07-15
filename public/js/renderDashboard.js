const stat = ({image, name, status}) => `
<td>
    ${image}
</td>
<td>
    ${name}
</td>
<td>
    ${status}
</td>
`;

function mapToStats(d){
    return stat ({image: d.Image,name : d.Names.join(', '), status: d.Status}) 
}

function render(running){
    var cont = d3.select("#images")
        .selectAll(".stats")
        .data(running)
        .html((d) => {return mapToStats(d);})
    cont.enter()
        .append("tr")
        .attr("class", "stats")
        .html((d) => {return mapToStats(d);})
    cont.exit().remove()
}

ws = connect("dashboardWS", (msg) => {
    console.log(msg)
    console.log(JSON.parse(msg.data))
    conts = JSON.parse(msg.data)
    render(conts)})