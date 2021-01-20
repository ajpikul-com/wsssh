all: export GOPRIVATE=github.com/ayjayt/AccessTunnel
all:
	$(MAKE) -C sshoverws/ || ( echo '**ERROR'; false )
	$(MAKE) -C server/ || ( error '**ERROR'; false )
	$(MAKE) -C client/ || ( error '**ERROR'; false )

run:
	tmux respawn-window -t :server -k
	sleep .2
	tmux respawn-window -t :client -k

env:
	tmux new-window -n client "$(CURDIR)/client/client; sleep infinity"
	tmux new-window -n server "$(CURDIR)/server/server; sleep infinity"
