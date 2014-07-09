An implementation and simulation of the consensus algorithm described at http://antirez.com/news/71

Definitions
===========
A `Process` is an in memory structure representing a process participating in the consensus protocol that cares about the output value of the algorithm. Despite being in memory, processes are not guaranteed communication with other processes.

Each process has an `Inbox` which is a channel of messages that represent an ordered list of complete messages received over the "network".

Each process also has an `Outbox` for sending messages to other processes.

The `Mailbox` (aka the "network") is a centralized execution context that deals with continued reading of the `Outbox` of each process and forwarding the message to the appropriate recipients `Inboxes`. It's expected that most `Forces` will be take effect in this layer.

`God` is a piece that can issue `Forces`. The general idea is that `God` will play with the "network" causing lag, packetloss and partitions.

`Frequency` is the piece of data in that each process in the simulation is trying to agree upon via the consensus algorithm.

All language refering to `Votes` and `Epoch` are specific to the algorithm.
