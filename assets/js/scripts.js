
$( document ).ready(function() {
    
    $.getJSON( "/getstate", function( data ) {
      
      console.log(data)

        for(var k in data){

            if (data[k].ProcessRunning){

                $("#list ul").append('<li><div class="name">' + data[k].Name +'</div> <div class="control running">Running <button type="button" onclick="stop(\'' + data[k].Backend + '\', \'' + data[k].Backend + '\')">Stop</button></div></li>');    
            }else{

                $("#list ul").append('<li><div class="name">' + data[k].Name +'</div> <div class="control stopped">Stopped <button type="button" onclick="start(\'' + data[k].Backend + '\', \'' + data[k].Backend + '\')">Start</button></div></li>');
            }
        }
    });
});


function start(backend, instance) {
    xmlhttp = new XMLHttpRequest();
    xmlhttp.open("GET", "start?process="+backend+"&instance=" + instance, true);
    xmlhttp.send();
    location.reload();
}
function stop(backend, instance) {
    xmlhttp = new XMLHttpRequest();
    xmlhttp.open("GET", "stop?process="+backend+"&instance=" + instance, true);
    xmlhttp.send();
    location.reload();
}
