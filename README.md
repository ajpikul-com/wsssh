# Access Tunnel

Access Tunnel is a server client pair that provides SSH access to behind-firewall devices (the entire fleet) by username.

## Server

You will connect to the server via a certain user @ ssh.
Not sure if the server will do any kind of authentication- the client can.
The server will figure out which device you want based on username and then proxy your ssh to a local port on that server, where the ssh connection from the client is listening to allow you tunneling to the client.

The server will probably be `groundcontrol.osoximeter.com`
`company_device_user` is the recommended format for user logins.

## Client

The client sets up a reverse tunnel with the server. It must supply it's name and some type of credentials. Maybe it can get the chip's serial number. It can probably supply a serial number and maybe a client cert. It also has to find the server.


A particular user on a particular client... will be associated with a particular username/password.


