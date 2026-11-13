const backend_addr = "http://drawing-board.duckdns.org:1216/";

const canvas = document.getElementById("drawing-screen");
const color_picker = document.getElementById("color-picker");
const size_picker = document.getElementById("size-picker");
const ctx = canvas.getContext("2d");




function getMousePosScaled(evt) {
  const rect = canvas.getBoundingClientRect();
  const scaleX = canvas.width / rect.width;
  const scaleY = canvas.height / rect.height;
  return {
    x: (evt.clientX - rect.left) * scaleX,
    y: (evt.clientY - rect.top) * scaleY
  };
}

function addPointToLine( points, current_point, radius, color ){

    if( points == null ){

        return [current_point];

    }else{

        ctx.beginPath();
        ctx.moveTo(points[points.length-1].x, points[points.length-1].y);
        ctx.lineTo(current_point.x, current_point.y);
        ctx.lineWidth = radius;
        ctx.strokeStyle = color;
        ctx.stroke();

        points.push({
            x: current_point.x,
            y: current_point.y,
            width: radius,
            color: color
        });
        return points;

    }

}

function drawLine( line ){
    for( let i=1;  i<line.length;  i++ ){
        ctx.beginPath();
        ctx.moveTo(line[i-1].x, line[i-1].y);
        ctx.lineTo(line[i].x, line[i].y);
        ctx.lineWidth = line[i].width;
        ctx.strokeStyle = line[i].color;
        ctx.stroke();
    }
}

// pray for no race condition
var lines = [];
var next_to_recv = 0;

var points = null;
var radius = 5;
var color = "#f00";
var is_drawing = false;
var number_of_points=0;

document.getElementById('drawing-screen').addEventListener('pointerdown', () => {

    points = null;
    is_drawing = true;
});
document.getElementById('drawing-screen').addEventListener('pointerup', () => {
    is_drawing = false;
    lines.push(points);
    console.log(number_of_points)
});
document.getElementById('drawing-screen').addEventListener('pointermove', (evt) => {
    if( ! is_drawing ) return;
    let cp = getMousePosScaled(evt);
    points = addPointToLine(points, cp, radius, color);
});

color_picker.addEventListener('input', (e) => {
    color = e.target.value;
});

size_picker.addEventListener('input', (e) => {
    radius = Number(e.target.value);
});


setInterval(async()=>{
    let lines_to_send = lines;
    if(points!=null && points.length > 1){
        lines_to_send.push(points)
        points = [points[points.length-1]]
    }
    lines = [];

    if( lines_to_send.length > 0){
        let resp = await fetch(backend_addr+"add",{
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({lines: lines_to_send.map((line)=>{
                return {points: line};
            })})
        });

        if (!resp.ok) {
        throw new Error(`HTTP error! Status: ${resp.status}`);
        }
    }

    let resp = await fetch(backend_addr+"get",{
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({first: next_to_recv})
    });

    if (!resp.ok) {
      throw new Error(`HTTP error! Status: ${resp.status}`);
    }

    const result = await resp.json();
    next_to_recv = result.last;
    result.lines.forEach(line => {
        try{
            drawLine(line.points);
            number_of_points += line.points.length;
        }catch{
            console.log(`error draving the lines: ${line}`);
        }
    });

}, 1000);
