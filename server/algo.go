package server

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

var True = true
var False = false

const NumProcesses = 3
const NumMessages = 100

type NetworkState struct {
	Lag time.Duration

	// int between 0 and 100 representing a percentage. 0 drops no
	// packets, 100 drops all packets
	Packetloss int
}

type Message struct {
	Message        string
	From           int
	To             int
	ProcessEpoch   int
	Frequency      int
	FrequencyEpoch int
	NetworkState   NetworkState
}

type Force struct {
	// All nil values represent properties that should be left alone
	Election   *bool
	Lag        *time.Duration
	Packetloss *int // See `NetworkState.Packetloss`
}

type Process struct {
	Id             int
	CurrentEpoch   int
	Frequency      int
	FrequencyEpoch int
	LastVoteEpoch  int
	NextElection   time.Time
	Election       *Election
	Inbox          chan *Message
	Outbox         chan *Message
	God            chan *Force
	Ticker         *time.Ticker
	ElectionForced bool
	NetworkState   *NetworkState
}

type Election struct {
	NewFrequency   int
	FrequencyEpoch int
	NumVotes       int
	NumProcesses   int
}

func NewProcess(id int, mailbox chan *Message) *Process {
	var nextElectionSeed time.Duration
	if id > 0 {
		nextElectionSeed = time.Minute
	}

	return &Process{
		Id:             id,
		CurrentEpoch:   0,
		FrequencyEpoch: 0,
		Frequency:      -1,
		LastVoteEpoch:  0,
		NextElection:   time.Now().Add(nextElectionSeed),
		Inbox:          make(chan *Message, 10),
		Outbox:         mailbox,
		God:            make(chan *Force, 1),
		Ticker:         time.NewTicker(5 * time.Second),
		NetworkState:   &NetworkState{Lag: time.Duration(0), Packetloss: 0},
	}
}

func (process *Process) Spawn() {
	go process.Run()
}

func (process *Process) Run() {
	for {
		wantsNewElection := process.Frequency == -1 ||
			process.ElectionForced == true

		if wantsNewElection &&
			time.Now().After(process.NextElection) {

			process.ElectionForced = false
			// Random wait between attempts for an election.
			secondsToWait := time.Duration(10+rand.Int31n(10)) * time.Second
			process.NextElection = time.Now().Add(secondsToWait)
			process.ElectMe()
		}

		process.Iterate()
	}
}

func (process *Process) Iterate() {
	select {
	case message := <-process.Inbox:
		process.HandleMessage(message)
	case _ = <-process.Ticker.C:
		// Send an update to a random neighbor
		process.SendUpdate(int(rand.Int31n(NumProcesses)))
	case force := <-process.God:
		fmt.Printf("Process #%d Force: %#v\n", process.Id, force)
		if force.Election != nil {
			process.ElectionForced = *force.Election
		}
	}
}

func (process *Process) HandleMessage(message *Message) {
	if process.CurrentEpoch < message.ProcessEpoch {
		// Always update the current epoch
		process.CurrentEpoch = message.ProcessEpoch
	}

	switch message.Message {
	case "heartbeat":
		//fmt.Printf("Received heartbeat. Process: %v Frequency: %v Epoch: %v\n",
		//process.Id, process.Frequency, process.FrequencyEpoch)
		if process.FrequencyEpoch >= message.FrequencyEpoch {
			return
		}

		process.Frequency = message.Frequency
		process.FrequencyEpoch = message.FrequencyEpoch
		fmt.Printf("Updating from heartbeat. Process: %v New Frequency: %v New Epoch: %v\n",
			process.Id, process.Frequency, process.FrequencyEpoch)
	case "elect_me":
		if process.CurrentEpoch > message.ProcessEpoch ||
			process.LastVoteEpoch >= process.CurrentEpoch ||
			process.FrequencyEpoch > message.FrequencyEpoch {
			// Do not vote if: I have a more recent view of time than
			// the message, or I already voted for this epoch, or my
			// last observed frequency change is more recent than the
			// requesting machine.
			fmt.Printf("Not voting. Process: %v Epoch: %v LastVoteEpoch: %v FrequencyEpoch: %v Message: %#v\n",
				process.Id, process.CurrentEpoch, process.LastVoteEpoch, process.FrequencyEpoch, message)
			return
		}

		fmt.Printf("You have my vote. Process %v Frequency: %v Epoch: %v\n",
			process.Id, message.Frequency, message.ProcessEpoch)
		process.LastVoteEpoch = process.CurrentEpoch
		message := &Message{
			Message:      "you_have_my_vote",
			ProcessEpoch: process.CurrentEpoch,
			From:         process.Id,
			To:           message.From,
			NetworkState: *process.NetworkState,
		}
		process.Outbox <- message
	case "you_have_my_vote":
		if message.ProcessEpoch < process.CurrentEpoch {
			return
		}

		fmt.Printf("Received vote. Process: %v Frequency: %v Epoch: %v\n",
			process.Id, process.Election.NewFrequency, process.Election.FrequencyEpoch)
		process.Election.NumVotes++
		if process.Election.NumVotes*2 > process.Election.NumProcesses {
			fmt.Println("New frequency elected.")
			process.Frequency = process.Election.NewFrequency
			process.FrequencyEpoch = process.Election.FrequencyEpoch
			process.PropagateFrequency()
		}
	default:
		fmt.Printf("Process #%d Message received: %v\n", process.Id, message)
	}
}

func (process *Process) SendUpdate(toProcessId int) {
	message := &Message{
		From:           process.Id,
		To:             toProcessId,
		ProcessEpoch:   process.CurrentEpoch,
		Frequency:      process.Frequency,
		FrequencyEpoch: process.FrequencyEpoch,
		Message:        "heartbeat",
		NetworkState:   *process.NetworkState,
	}

	process.Outbox <- message
}

func (process *Process) PropagateFrequency() {
	for peerId := 0; peerId < NumProcesses; peerId++ {
		if process.Id == peerId {
			continue
		}

		process.SendUpdate(peerId)
	}
}

func (process *Process) ElectMe() {
	process.CurrentEpoch++
	process.LastVoteEpoch = process.CurrentEpoch
	process.Election = &Election{
		NewFrequency:   int(rand.Int31n(100)),
		FrequencyEpoch: process.CurrentEpoch,
		NumVotes:       1,
		NumProcesses:   NumProcesses,
	}

	fmt.Printf("Sending elect_me. Process: %v Election: %#v\n",
		process.Id, process.Election)

	for peerId := 0; peerId < NumProcesses; peerId++ {
		if process.Id == peerId {
			continue
		}

		message := &Message{
			Message:        "elect_me",
			ProcessEpoch:   process.CurrentEpoch,
			Frequency:      process.Election.NewFrequency,
			FrequencyEpoch: process.FrequencyEpoch,
			From:           process.Id,
			To:             peerId,
			NetworkState:   *process.NetworkState,
		}

		process.Outbox <- message
	}
}

func Mailbox(processes []*Process, mailbox chan *Message) {
	for messageNum := 0; messageNum < NumMessages; messageNum++ {
		message := <-mailbox
		if rand.Intn(100) < message.NetworkState.Packetloss {
			fmt.Printf("Dropped message: %#v\n", message)
			continue
		}

		if message.NetworkState.Lag > time.Duration(0) {
			go func(lag time.Duration) {
				<-time.After(lag)
				processes[message.To].Inbox <- message
			}(message.NetworkState.Lag)
			continue
		}

		fmt.Printf("Mailbox: %#v\n", message)
		processes[message.To].Inbox <- message
	}
}

func Play() {
	mailbox := make(chan *Message, 10)
	processes := make([]*Process, 0, NumProcesses)
	for processNum := 0; processNum < NumProcesses; processNum++ {
		processes = append(processes, NewProcess(processNum, mailbox))
	}

	for _, process := range processes {
		process.Spawn()
	}

	processes[2].NetworkState.Packetloss = 100

	go func() {
		time.Sleep(5 * time.Second)
		fmt.Println("Forcing an election")
		processes[1].God <- &Force{Election: &True}
	}()

	go Mailbox(processes, mailbox)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		RootHandler(w, r, processes)
	})
	http.HandleFunc("/election", func(w http.ResponseWriter, r *http.Request) {
		ElectionHandler(w, r, processes)
	})
	http.HandleFunc("/lag", LagHandler)
	http.ListenAndServe(":8080", nil)
}
