# wsconn

This readme sucks and I don't care.

The `type Conn struct` provided by gorilla websockets does not fulfil `net.Conn` because 
it doesn't have simple `Read` and `Write` functions.

There are five channels to deal with:

1) Text
2) Binary
3) Ping
4) Pong
5) Close

Our wrapper needs to retain websocket `io.Reader`s over several calls to `WSConn.Read()`, since `WSConn.Read()` may not empty the underlying `io.Reader` the first call and asking for a new `io.Reader` would cause an error.

FIRST TODO:

Setup server and get it running
Set up client to allow user to send whatever payloads.

TODO:

* I hate the name WSTransport- a `Transport` is a thing in `net/http`, an implementation of a `RoundTripper`
* I hate that the struct names Conn as member instead of inherits- but then again I guess I wanted to wrap _everything_
* I hate my implementaiton of `Upgrade` I should have wrapped the `websocket.Upgrader`
* Also hat emy implementation of `Dial, I should have wrapped the whole thing.


