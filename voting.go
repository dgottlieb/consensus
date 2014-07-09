package main

import (
	"time"
	"fmt"
	"math/rand"
	"runtime"
)

const NumProcesses = 3
const NumMessages = 100

type Message struct {
	Message string
	From int
	To int
	ProcessEpoch int
	Frequency int
	FrequencyEpoch int
}

type Force struct {
}

type Process struct {
	Id int
	CurrentEpoch int
	Frequency int
	FrequencyEpoch int
	LastVoteEpoch int
	NextElection time.Time
	Election *Election
	Inbox chan *Message
	Outbox chan *Message
	God chan *Force
	Ticker *time.Ticker
}

type Election struct {
	NewFrequency int
	FrequencyEpoch int
	NumVotes int
	NumProcesses int
}

func NewProcess(id int, mailbox chan *Message) *Process {
	return &Process{
		Id: id,
		CurrentEpoch: 0,
		FrequencyEpoch: 0,
		Frequency: -1,
		LastVoteEpoch: 0,
		Inbox: make(chan *Message, 10),
		Outbox: mailbox,
		Ticker: time.NewTicker(5 * time.Second),
	}
}

func (process *Process) Spawn() {
	go process.Run()
}

func (process *Process) Run() {
	for {
		if process.Frequency == -1 &&
			time.Now().After(process.NextElection) {
			secondsToWait := time.Duration(30 + rand.Int31n(30)) * time.Second
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
		fmt.Printf("Process #%d Force: %d\n", process.Id, force)
	}
}

func (process *Process) HandleMessage(message *Message) {
	if process.CurrentEpoch < message.ProcessEpoch {
		// Always update the current epoch
		process.CurrentEpoch = message.ProcessEpoch
	}

	switch message.Message {
	case "heartbeat":
		fmt.Printf("Received heartbeat. Process: %v Frequency: %v Epoch: %v\n",
			process.Id, process.Frequency, process.FrequencyEpoch)
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
			Message: "you_have_my_vote",
			ProcessEpoch: process.CurrentEpoch,
			From: process.Id,
			To: message.From,
		}
		process.Outbox <- message
	case "you_have_my_vote":
		if message.ProcessEpoch < process.CurrentEpoch {
			return
		}

		fmt.Printf("Received vote. Process: %v Frequency: %v Epoch: %v\n",
			process.Id, process.Election.NewFrequency, process.Election.FrequencyEpoch)
		process.Election.NumVotes++
		if process.Election.NumVotes * 2 > process.Election.NumProcesses {
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
		From: process.Id,
		To: toProcessId,
		ProcessEpoch: process.CurrentEpoch,
		Frequency: process.Frequency,
		FrequencyEpoch: process.FrequencyEpoch,
		Message: "heartbeat",
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
		NewFrequency: int(rand.Int31n(100)),
		FrequencyEpoch: process.CurrentEpoch,
		NumVotes: 1,
		NumProcesses: NumProcesses,
	}

	fmt.Printf("Sending elect_me. Process: %v Election: %#v\n",
		process.Id, process.Election)

	for peerId := 0; peerId < NumProcesses; peerId++ {
		if process.Id == peerId {
			continue
		}

		message := &Message{
			Message: "elect_me",
			ProcessEpoch: process.CurrentEpoch,
			Frequency: process.Election.NewFrequency,
			FrequencyEpoch: process.FrequencyEpoch,
			From: process.Id,
			To: peerId,
		}

		process.Outbox <- message
	}
}

func Mailbox(processes []*Process, mailbox chan *Message) {
	for messageNum := 0; messageNum < NumMessages; messageNum++ {
		message := <-mailbox
		fmt.Printf("Mailbox: %v\n", message)
		processes[message.To].Inbox <- message
	}
}

func main() {
	runtime.GOMAXPROCS(4)
	rand.Seed(time.Now().Unix())
	mailbox := make(chan *Message, 10)
	processes := make([]*Process, 0, NumProcesses)
	for processNum := 0; processNum < NumProcesses; processNum++ {
		processes = append(processes, NewProcess(processNum, mailbox))
	}

	for _, process := range processes {
		process.Spawn()
	}

	Mailbox(processes, mailbox)
}
