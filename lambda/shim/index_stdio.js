
const MAX_FAILS = 4;

process.env['PATH'] = process.env['PATH'] + ':' + process.env['LAMBDA_TASK_ROOT'] + ':' + __dirname;
// console.log(process.env['PATH']);

var child_process = require('child_process'),
  go_proc = null,
  done = console.log.bind(console),
  fails = 0;

// The following (old) approach has a bug where an internal server error is returned due to multiple requests
// which trigger multiple Lambdas trying to use the same pipe.
// go_proc stdin write error:
// {
//   "code": "ECONNRESET",
//   "errno": "ECONNRESET",
//   "syscall": "read"
// }
// {
//   "errorMessage": "read ECONNRESET",
//   "errorType": "Error",
//   "stackTrace": [
//       "exports._errnoException (util.js:870:11)",
//       "Pipe.onread (net.js:544:26)"
//   ]
// }

// https://github.com/mhart/epipebomb
// function epipeBomb(stream, callback) {
//   if (stream == null) stream = process.stdout
//   if (callback == null) callback = process.exit

//   function epipeFilter(err) {
//     if (err.code === 'EPIPE') return callback()

//     // If there's more than one error handler (ie, us),
//     // then the error won't be bubbled up anyway
//     if (stream.listeners('error').length <= 1) {
//       stream.removeAllListeners()     // Pretend we were never here
//       stream.emit('error', err)       // Then emit as if we were never here
//       stream.on('error', epipeFilter) // Then reattach, ready for the next error!
//     }
//   }

//   stream.on('error', epipeFilter)
// }
// epipeBomb();
// ^ didn't seem to help.

// TODO: Try to avoid this? I can't imagine copying, not moving, this app over each time is great for performance.
// To avoid permission issues...
// Really hate to copy this file each time (and mv didn't work).
// var fs = require('fs');
// if (!fs.existsSync('/tmp/aegis_app')) {
//   child_process.execSync('cp aegis_app /tmp/aegis_app && chmod +x /tmp/aegis_app');
// }

// Debug
// var fs = require('fs');
// items = fs.readdirSync('./');
// for (var i=0; i<items.length; i++) {
//     console.log("");
//     console.log("");
//     console.log("FILE: " + items[i]);
//     var stats = fs.statSync(items[i]);
//     console.log("------------------------");
//     console.log(stats);
//     console.log();
 
//     if (stats.isFile()) {
//         console.log('    file');
//     }
//     if (stats.isDirectory()) {
//         console.log('    directory');
//     }
 
//     console.log('    size: ' + stats["size"]);
//     console.log('    mode: ' + stats["mode"]);
// }

// This creates problems.
// The problem is that when the Lambda is invoked quick enough, the pipe could be closed already but a new 
// go_proc.stdin.write() is called from the hanlder. This results in the ECONNRESET or EPIPE error.
// Basically, there's nothing to write to. This then returns an error.
// So why not just spawn a new process on each Lambda invocation? I know the "re-use" part of it
// becomes an performance concern, but the Go process spawns fast and we don't really want
// The go process to be re-used anyway. We want to think about them as unique processes. Stateless.

// (function new_go_proc() {
//   // pipe stdin/out, blind passthru stderr
//   go_proc = child_process.spawn('/tmp/aegis_app', { stdio: ['pipe', 'pipe', process.stderr] });

//   //child_process.execSync('chmod +x aegis_app'); // can't do this, operation not permitted
//   // go_proc = child_process.spawn('./aegis_app', { stdio: ['pipe', 'pipe', process.stderr] }); // this used to work, why not now? ¯\_(ツ)_/¯

//   go_proc.on('error', function(err) {
//     process.stderr.write("go_proc errored: "+JSON.stringify(err)+"\n");
//     if (++fails > MAX_FAILS) {
//       process.exit(1); // force container restart after too many fails
//     }
//     new_go_proc();
//     done(err);
//   });

//   go_proc.on('exit', function(code) {
//     process.stderr.write("go_proc exited prematurely with code: "+code+"\n");
//     if (++fails > MAX_FAILS) {
//       process.exit(1); // force container restart after too many fails
//     }
//     new_go_proc();
//     done(new Error("Exited with code "+code));
//   });

//   go_proc.stdin.on('error', function(err) {
//     process.stderr.write("go_proc stdin write error: "+JSON.stringify(err)+"\n");
//     if (++fails > MAX_FAILS) {
//       process.exit(1); // force container restart after too many fails
//     }
//     new_go_proc();
//     done(err);
//   });

//   var data = null;
//   go_proc.stdout.on('data', function(chunk) {
//     fails = 0; // reset fails
//     if (data === null) {
//       data = new Buffer(chunk);
//     } else {
//       data.write(chunk);
//     }
//     // check for newline ascii char 10
//     if (data.length && data[data.length-1] == 10) {
//       var output = JSON.parse(data.toString('UTF-8'));
//       data = null;
//       done(null, output);
//     };
//   });
// })();

// New, simplified shim.
// Still looks to re-use an existing child process. Catches it not running on child process exit event or Lambda cold start.
// var processRunning = false;
// var go_proc;

exports.handler = function(event, context) {
  // Again, errors with the sockets being closed.
  // ie. "Error: This socket has been ended by the other party" at Socket.writeAfterFIN [as write] (net.js:268:12)
  // if (!processRunning && !go_proc) {
  //   console.info("Spawning child Go process.");
  //   go_proc = child_process.spawn('/tmp/aegis_app', { stdio: ['pipe', 'pipe', process.stderr] });
  //   processRunning = true;
  // } else {
  //   console.info("Re-using existing child Go process.");
  // }

  // Just spawn a new child process to handle each invocation. Hopefully the Go app doesn't take a while to start.
  // It's going to greatly depend on the app, but I don't see value in a huge Go app as a "microservice" or "cloud function."
  // So I'm going to assume 99.9% of all Go apps running on AWS Lambda are small in nature and run fast.
  //
  // TODO: Look into the Lambda container re-use and re-using child processes...It "should" work...But in reality there
  // are a lot of situations where things get closed and then are trying to be written to again which results in annoying errors.
  var go_proc = child_process.spawn('./aegis_app', { stdio: ['pipe', 'pipe', process.stderr] });

  // add to event the invoke time (oddly not present in context or event)
  // I wish it was from when API Gateway received the request...not sure if there's a way to pass that info.
  var hrTime = process.hrtime();
  event.handlerStartHrTime = hrTime;
  // and normal JavaScript milliseconds
  event.handlerStartTimeMs = new Date().getTime();

  // always output to current context's done
  done = context.done.bind(context);

  go_proc.stdin.write(JSON.stringify({
    "event": event,
    "context": context
  })+"\n");

  // Handle data sent back out by the spawned child process.
  var data = null;
  go_proc.stdout.on('data', function(chunk) {
    fails = 0; // reset fails
    if (data === null) {
      data = new Buffer(chunk);
    } else {
      data.write(chunk);
    }
    // check for newline ascii char 10
    if (data.length && data[data.length-1] == 10) {
      var output = JSON.parse(data.toString('UTF-8'));
      data = null;
      done(null, output);
    };
  });

  // If the spawned child process has an error, log and return that.
  go_proc.on('error', function(err) {
    process.stderr.write("go_proc errored: "+JSON.stringify(err)+"\n");
    done(err);
  });

  // If it exits, mark the processRunning as false so it can restart on next Lambda invocation.
  // Also call done() with an error and the exit code.
  go_proc.on('exit', function(code) {
    process.stderr.write("go_proc exited prematurely with code: "+code+"\n");
    processRunning = false;
    if (code !== 0) {
      done(new Error("Exited with code "+code));
    } else {
      // If it exited with a 0 status code then technically nothing was wrong.
      // Is it expected? That's another question, but it's not "wrong." Return an empty response.
      // Application logging should help.
      done(null, {})
    }
  });
}