# dsys-hw4
Distrubuted Systems Homework #4

```
$ go run . -help
Usage of dsys-hw4:
  -name string
        Name of this instance (default "<no name>")
  -port string
        Port to listen on (default "50050")
```

`name` is used solely for naming log files; logs are written to `<name>.log`. A handy way to view the logs once all processes are running is `tail -fqn 0 *.log`. Or to view the logs retroactively try `sort -k3 *.log`, and better yet, combine them `sort -k3 *.log | tail -n10 && tail -fqn0 *.log` :)

## Run

To start a process, run `go run . -name A -port 5000` - it'll display a message once it's ready to receive input.
Do this a couple more times to start more processes.

Each process represents a peer in a peer-to-peer style network (a complete graph). To connect these peers in a, type the command `connect <port>` into stdin for a source peer, where `<port>` is the port number of the target peer. The connection graph of the source and target will then form a new complete graph (where all peers are connected). In other words, just run `connect` once in the first peer for each other peer, and they should all connect up nicely.

## Taking the lock

... and entering the critical section. There are two commands beside `connect`: `lock` and `unlock`.Unsurprisingly, `lock` will attempt to enter the critical section, coordinating with all other peers to make sure only one is allowed in at a time. Likewise, `unlock` will exit the critical section immediately.

## Why it's correct

We've implemented the Ricart-Agrawala algorithm as presented in the lecture. To make its correctness as obvious as possible, we've implemented it as an [interface in its own file](ricart-agrawala/ricart-agrawala.go) where each line of code almost corresponds one-to-one with the pseudo-code from the slides. Additionally, we're reusing our [Lamport time](ricart-agrawala/lamport.go) implementation (slightly modified) from a previous assignment which we know to be correct because it was accepted.
