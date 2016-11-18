
const MAX_FAILS = 4;

var child_process = require('child_process'),
	go_proc = null,
	done = console.log.bind(console),
	fails = 0
	http = require('http');

(function new_go_proc() {

	// pipe stdin/out, blind passthrough stderr
	go_proc = child_process.spawn('./aegis_app');

	go_proc.on('error', function(err) {
		process.stderr.write("go_proc errored: "+JSON.stringify(err)+"\n");
		if (++fails > MAX_FAILS) {
			process.exit(1); // force container restart after too many fails
		}
		new_go_proc();
		done(err);
	});

	go_proc.on('exit', function(code) {
		process.stderr.write("go_proc exited prematurely with code: "+code+"\n");
		if (++fails > MAX_FAILS) {
			process.exit(1); // force container restart after too many fails
		}
		new_go_proc();
		done(new Error("Exited with code "+code));
	});
})();

exports.handler = function(event, context) {

	// always output to current context's done
	done = context.done.bind(context);

	// go_proc.stdin.write(JSON.stringify({
	// 	"event": event,
	// 	"context": context
	// })+"\n");

	var port = 9500;
	if(event.stageVariables !== undefined && event.stageVariables.lambdaPort !== undefined && event.stageVariables.lambdaPort !== null) {
		port = event.stageVariables.lambdaPort;
	}

	// http passthrough
	var options = {
	  hostname: 'localhost',
	  port: port,
	  path: event.path || '/',
	  method: event.httpMethod || 'GET',
	  headers: event.headers || {
	    'Content-Type': 'application/json',
	    'Content-Length': Buffer.byteLength(event.body)
	  }
	};

	// console.log("EVENT:");
	// console.log(event);

	// console.log("OPTIONS:");
	// console.log(options);

	callback = function(response) {
	  var resBody = '';

	  //another chunk of data has been recieved, so append it to `str`
	  response.on('data', function (chunk) {
	    resBody += chunk;
	  });

	  response.on('error', function(e) {
		//console.log(`problem with request: ${e.message}`);
		done(e);
	  });

	  //the whole response has been recieved, so we just print it out here
	  response.on('end', function () {
	  	console.log("HTTP REQUEST END");
	  	console.log(resBody)
	  	//var j = JSON.parse(resBody);
	  	//var s = JSON.stringify(j);
	    done(null, {
        	"statusCode": "200",
        	"headers": {
        		"Content-Type": "application/json"
        	},
        	//"body": JSON.parse(resBody)
        	"body": JSON.stringify(resBody)
        });
	  });
	}
	http.request(options, callback).end();


}