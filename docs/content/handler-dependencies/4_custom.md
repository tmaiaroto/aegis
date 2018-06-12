# Custom

You can inject your own handler dependencies under the `Custom` field. This is great for 3rd party dependencies you
wish to use as well as anything you want to make available to each handler.

These dependencies contain no "configuration" closure (unlike Services) and if your function is simple enough, you
may not even need to pass them through `Custom`. Keep in mind the scoping in Go. If you've defined something in
your package ("main" perhaps), it's available to all of the functions within that package...This often will include
your handlers.

So before you reach to shove everything through to your handlers, think about it a bit. Lambda functions should
be easy to follow and your code should read clean. Remember that `Custom` here is a <span class="nowrap">`map[string]interface{}`</span>
so you will likely encounter the need to check for nil and use type assertions.

Aegis isn't trying to limit you and sometimes you may have no other option, but no one says you have to pass
everything through to each handler in this way.