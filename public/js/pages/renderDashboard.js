const containerRow = ({image, name, status}) => `
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

function mapContainerToRow(d) {
    return containerRow({image: d.Image, name: d.Names.join(', '), status: d.State})
}

const imageRow = ({repo, tag, created, imgSize}) => `
    <td>
        ${repo}
    </td>
    <td>
        ${tag}
    </td>
    <td>
        ${created}
    </td>
    <td>
        ${imgSize}
    </td>
`;

function mapImageToRow(img) {
    let splitTag = img.RepoTags.toString().split(":");
    let size = "";
    if (img.Size / 1e9 >= 1) {
        size = (img.Size / 1e9).toFixed(2) + "GB"
    } else {
        size = Math.floor(img.Size / 1e6) + "MB"
    }
    return imageRow({repo: splitTag[0], tag: splitTag[1], created: img.Created, imgSize: size})
}


function render(running) {
    let cont = d3.select("#containers")
        .selectAll(".contStats")
        .data(running.Containers)
        .html((d) => {
            return mapContainerToRow(d);
        });
    cont.enter()
        .append("tr")
        .attr("class", "contStats clickable-row")
        .attr("data-href", (d) => {
            return "/container/" + d.Id
        })
        .html((d) => {
            return mapContainerToRow(d);
        });
    cont.exit().remove();
    let images = d3.select("#images")
        .selectAll(".imgStats")
        .data(running.Images)
        .html((d) => {
            return mapImageToRow(d);
        });
    images.enter()
        .append("tr")
        .attr("class", "imgStats clickable-row")
        .attr("data-href", (d) => {
            return "/container/" + d.Id
        })
        .html((d) => {
            return mapImageToRow(d);
        });
    images.exit().remove()
}

let ws = connect((msg) => {
    console.log(msg);
    console.log(JSON.parse(msg.data));
    let conts = JSON.parse(msg.data);
    render(conts)
});

$(document).ready(() => {
    $('body').on('click', '.clickable-row', function () {
        window.location = $(this).data("href");
    });
});

