## Developer Tip

use `make local` to point everything to local versions
use `make official` before push

## wsconn

Knowing that websockets has 5 channels: `ping`, `pong`, `close`, `text`, `binary`,
`wsconn` wraps it so that the `binary` channel is exposed as a `net.Conn` interface,
so any packages that want a `net.Conn` interface (like `crypto/ssh`) can be routed over websockets.
Finally, we can use `ssh` from starbucks! `gorilla/websocket` naturally handles `ping` `pong` and `close` outside
of the `Read()` loop buffers, but `Text` is routed to a separate buffer than you can `Read()` and `Write()` from without
whatever is attached to your `binary net.Conn()` knowing.


### testclient and server

`testclient` and `server` are just tests

these are essentiall proofs for wsconn. Write's don't verify they went through. They just assume that no one takes only part of their write. Reads on the other hand will check to see if there is more to read until there is none.

Client and Server are basically the same here except client fully expects to have control of flow and will have a bunch of confused functions if server closes on it.

## wscli

`wscli` can be used to test either `gorilla/websocket` or `wsssh/wsconn`. **NOTE** THIS IS AN AWESOME PROJECT BUT IT IS FAR FROM DONE

## old

`old_ssh_tunnel` is graveyard reference code


