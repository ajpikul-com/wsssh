

Undocumented, this is a set of tools to work with ssh + websockets + more

`wscli` can be used to test either `gorilla/websocket` or `wsssh/wsconn`

Knowing that websockets has 5 channels: `ping`, `pong`, `close`, `text`, `binary`,
`wsconn` wraps it so that the `binary` channel is exposed as a `net.Conn` interface,
so any packages that want a `net.Conn` interface (like `crypto/ssh`) can be routed over websockets.
Finally, we can use `ssh` from starbucks! `gorilla/websocket` naturally handles `ping` `pong` and `close` outside
of the `Read()` loop buffers, but `Text` is routed to a separate buffer than you can `Read()` and `Write()` from without
whatever is attached to your `binary net.Conn()` knowing.