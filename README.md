# Access Tunnel

Access Tunnel is a server client pair that provides SSH access to behind-firewall devices (the entire fleet) by username.

## Server

You will connect to the server via a certain user @ ssh.
Not sure if the server will do any kind of authentication- the client can to.
The server will figure out which device you want based on username and then proxy your ssh to a local port on that server, where the ssh connection from the client is listening to allow you tunneling to the client.

The server will probably be `groundcontrol.osoximeter.com`
All user names will probably be prefixed by `dev_`

## Client

The client sets up a reverse tunnel with the server. It must supply it's name and some type of credentials. Maybe it can get the chip's serial number. It can probably supply a serial number and maybe a client cert. It also has to find the server.

## Architecture Concept


The final concept would be:

Client connects to server w/ reverse proxy over SSH or SSH over Websockets if it has to
The client authenticates with a signed client cert, probably.
What happens if it loses the cert?
How would it lose the cert? Is the device given a forever cert? How do we associate a certain cert with it's location?
This is probably two factor authentication. The device tries to login. Someone connected to it via bluetooth logs in with their account. The device is given a cert
related to their account, and the device is their responsibility for a particular length of time.

Clients logging with unsigned certs can probabably log in, a certain number of them, and we may be able to manually approve. A device will have a serial number, so we can trace where it's coming from.

So the client connects to a server w/ a reverse proxy over SSH or SSH/Websockets if it has to, and it authenticates as authentic with a client cert that the client acquired.
If it's not authenticated, it can still log in, it just, we're notified. The reverse proxy has to be given a particular port at runtime. We need a database of client IDs, users, and ports. The client would probably connect using the username it wants to be connected to by.

The server then can route new users to the reverse tunnel. As in, open up a client on that port, and connect them. That's probably fine. 

What if the SBC Client starts a bash terminal and and pipes in whatever the server sends down. The server would then provide pipe. That would require opening multiple channels? We'll see.
