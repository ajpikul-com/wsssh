## Developer Tip


* ssh on the binary connection doesn't seem "closeable"- it goes after underlying connection
so it continues to sit on read, which is good, because something needs to sit on read as that's
basically the main loop
* but it doesn't close actually close anything because close is blocked!
* also ping takes two or three pings to realize that connection is closed

* use `make local` to point everything to local versions
* use `make official` before push

## wsconn

Knowing that websockets has 5 channels: `ping`, `pong`, `close`, `text`, `binary`,
`wsconn` wraps it so that the `binary` channel is exposed as a `net.Conn` interface,
so any packages that want a `net.Conn` interface (like `crypto/ssh`) can be routed over websockets.
Finally, we can use `ssh` from starbucks! `gorilla/websocket` naturally handles `ping` `pong` and `close` outside
of the `Read()` loop buffers, but `Text` is routed to a separate buffer than you can `Read()` and `Write()` from without
whatever is attached to your `binary net.Conn()` knowing.


### testclient and server

All of this is basically old, although they do a little more than sysboss/sysclient do. Those are a better start though.

`*-ssh` contians interesting info.

So does `old_ssh_tunnel`

## wscli

`wscli` can be used to test either `gorilla/websocket` or `wsssh/wsconn`. **NOTE** THIS IS AN AWESOME PROJECT BUT IT IS FAR FROM DONE

## old

`old_ssh_tunnel` is graveyard reference code


