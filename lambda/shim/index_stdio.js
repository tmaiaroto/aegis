process.env['PATH'] = process.env['PATH'] + ':' + process.env['LAMBDA_TASK_ROOT'] + ':' + __dirname;
// console.log(process.env['PATH']);

var child_process = require('child_process');
if (go_proc === undefined) {
  // console.log("starting new child process because go_proc was undefined");
  var go_proc = child_process.spawn('./aegis_app', { stdio: ['pipe', 'pipe', process.stderr] });
}

// If the spawned child process has an error, log and return that.
go_proc.on('error', function(err) {
  // process.stderr.write("go_proc errored: "+JSON.stringify(err)+"\n");
  // done(err);

  // Log it to CloudWatch (previously, attempted to return through API Gateway)
  console.error('[aegis] go_proc errored: %s', err);

  // Node.js process exit will create another cold start.
  process.exit(1);
});

// If it exits, call done() with an error and the exit code.
go_proc.on('exit', function(code, signal) {
  // Again, just log it to CloudWatch and exit.
  console.error('[aegis] go_proc exit: code=%s signal=%s', code, signal);
  // console.log('callbacks at exit:', callbacks);
  
  process.exit(1);
});

// Handle data coming back out
var data = null; // <-- initial value, reset to null on each handler()
go_proc.stdout.on('data', function(chunk) {
 
  fails = 0; // reset fails
  if (data === null) {
    data = new Buffer(chunk);
  } else {
    data.write(chunk);
  }
  // check for newline ascii char 10
  // if (data.length && data[data.length-1] == 10) {
  //   var output = JSON.parse(data.toString('UTF-8'));

  //   //console.log("output from go:", output);
  //   //console.log("callbacks:", callbacks);

  //   // Get a reference to the callback and remove it from the parent scope so it's only sent back once.
  //   var c = callbacks[output.headers["request-id"]] || function(err, data) { console.log("callback not found"); console.log("passed data:", data); };
  //   delete callbacks[output.headers["request-id"]];
  //   // Reset data here as well
  //   data = null;

  //   console.log("callback for request id: ", output.headers["request-id"]);
  //   // done(null, output); <-- old
  //   c(null, output);
  // } else {
  //   console.log("no new line");
  //   c(null, {statusCode: 500, body: ""});
  // }

  var output = JSON.parse(data.toString('UTF-8'));

  //console.log("output from go:", output);
  //console.log("callbacks:", callbacks);

  // Get a reference to the callback and remove it from the parent scope so it's only sent back once.
  var c = callbacks[output.headers["request-id"]] || function(err, data) { console.log("[aegis] callback not found"); console.log("[aegis] passed data:", data); };
  delete callbacks[output.headers["request-id"]];
  // Reset data here as well
  data = null;

  // console.log("callback for request id: ", output.headers["request-id"]);
  c(null, output);
});

// Use the new callback instead of done() succeed() etc.
// Keep a map of multiple callbacks in the event of concurrent requests.
var callbacks = {};

exports.handler = function(event, context, cb) {
  // http://docs.aws.amazon.com/lambda/latest/dg/nodejs-prog-model-context.html
  // Not sure I want this...
  context.callbackWaitsForEmptyEventLoop = false;
  
  // add to event the invoke time (oddly not present in context or event)
  // I wish it was from when API Gateway received the request...not sure if there's a way to pass that info.
  var hrTime = process.hrtime();
  event.handlerStartHrTime = hrTime;
  // and normal JavaScript milliseconds
  event.handlerStartTimeMs = new Date().getTime();

  // always output to current context's done
  // done = context.done.bind(context); <-- OLD

  //console.log("event:", event);
  //console.log("context:", context);
  // Use the ID given to us by Lambda/API Gateway.
  callbacks[event.requestContext.requestId] = cb;

  //console.log("event request id:", event.requestContext.requestId);
  // console.log("callbacks before stdin to app:", callbacks);

  // Pipe into the Go app the event data as JSON string
  go_proc.stdin.write(JSON.stringify({
    "event": event,
    "context": context
  })+"\n");

  // Still need to set data in parent scope to null on this new handler.
  // Otherwise we don't know which output goes with which response. When it starts/ends.
  data = null;
  
}