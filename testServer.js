//Lets require/import the HTTP module
var http = require('http');
var url = require('url');

//Lets define a port we want to listen to
const PORT=8080;
var map = {a: 0, b: 0, c: 0};
//We need a function which handles requests and send response
function handleRequest(request, response){
  var inUrl = request.url;

  if (inUrl === '/simulateServers') {
    host = request.headers['host'];
    if (host === 'testa:8080'){
      setTimeout(function() {
        map.a += 1;
        console.dir(map);
        response.end('a');
      }, 1000);
      return;
    }
    if (host === 'testb:8080') {
      setTimeout(function() {
        map.b += 1;
        console.dir(map);
        response.end('b');
      }, 666);
      return;
    }
    if (host === 'testc:8080') {
      setTimeout(function() {
        map.c += 1;
        console.dir(map);
        response.end('c');
      }, 334);
      return;
    }
  }
  response.end('huh');
}
//Create a server
var server = http.createServer(handleRequest);

//Lets start our server
server.listen(PORT, function(){
    //Callback triggered when server is successfully listening. Hurray!
    console.log("Server listening on: http://localhost:%s", PORT);
});

//baseline:
//ab -c 50 -n 1000 http://localhost:8090/simulateServers
//12.098s total, 577ms mean
//7236K consumed

//most recent edits
//ab -c 50 -n 1000 http://localhost:8090/simulateServers
//11.596s total, 561ms mean
//7212K consumed

//after removing incorrect optimization (which doesn't affect this test)
//ab -c 50 -n 1000 http://localhost:8090/simulateServers
//11.548s total, 560ms mean
//7176K consumed
