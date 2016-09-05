//Lets require/import the HTTP module
var http = require('http');
var url = require('url');

//Lets define a port we want to listen to
const PORT=8080;
var map = {a: 0, b: 0, c: 0};
//We need a function which handles requests and send response
function handleRequest(request, response){
  var inUrl = request.url;

  if (inUrl === '/simulateUnevenServers') {
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
  if (inUrl === '/simulateServers') {
    host = request.headers['host'];
    if (host === 'testa:8080'){
      setTimeout(function() {
        map.a += 1;
        console.dir(map);
        response.end('a');
      }, 300);
      return;
    }
    if (host === 'testb:8080') {
      setTimeout(function() {
        map.b += 1;
        console.dir(map);
        response.end('b');
      }, 300);
      return;
    }
    if (host === 'testc:8080') {
      setTimeout(function() {
        map.c += 1;
        console.dir(map);
        response.end('c');
      }, 300);
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
//ab -c 50 -n 1000 http://localhost:8090/simulateUnevenServers
//12.098s total, 577ms mean
//7236K consumed

//most recent edits
//ab -c 50 -n 1000 http://localhost:8090/simulateUnevenServers
//11.596s total, 561ms mean
//7212K consumed

//after removing incorrect optimization (which doesn't affect this test)
//ab -c 50 -n 1000 http://localhost:<port>/simulateUnevenServers
//bestofnlb:
//11.548s total, 560ms mean
//random:
//14.417s total, 683ms mean
//jsq:
//11.415s total, 560ms mean
//vulcand/oxy/roundrobin/rr:
//14.070s total, 670ms mean
//vulcand/oxy/roundrobin/rebalancer:
//14.083s total, 671ms mean

//even server behavior of 300ms
//ab -c 50 -n 1000 http://localhost:<port>/simulateServers
//bestofnlb:
//6.282s total, 312ms mean
//random:
//6.246s total, 310ms mean
//jsq:
//6.132s total, 305ms mean
//vulcand/oxy/roundrobin/rr:
//6.182s total, 307ms mean
//vulcand/oxy/roundrobin/rebalancer:
//6.185s total, 307ms mean
