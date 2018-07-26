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

function mapToStats(d) {
    return stat({image: d.Image, name: d.Names.join(', '), status: d.Status})
}

function render(running) {
    let cont = d3.select("#images")
        .selectAll(".stats")
        .data(running)
        .html((d) => {
            return mapToStats(d);
        });
    cont.enter()
        .append("tr")
        .attr("class", "stats clickable-row")
        .attr("data-href", (d)=>{
            return "/container/" + d.Id
        })
        .html((d) => {
            return mapToStats(d);
        });
    cont.exit().remove()
}

let ws = connect((msg) => {
    console.log(msg);
    console.log(JSON.parse(msg.data));
    let conts = JSON.parse(msg.data);
    render(conts)
});

$(document).ready(()=>{
    $('body').on('click', '.clickable-row', function() {
        window.location = $(this).data("href");
    });
});

