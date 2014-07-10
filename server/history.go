package server

import "sync"

type ElectionHistory struct {
	Id               int
	ElectorId        int
	RequestFrequency int
	SentMessages     []*Message
	ReceivedMessages []*ReceivedMessage
	Successful       bool
	*sync.Mutex
}

type ReceivedMessage struct {
	message      *Message
	processState *ProcessState
	successful   bool
}

type ProcessState struct {
	Id             int
	CurrentEpoch   int
	FrequencyEpoch int
	LastVoteEpoch  int
}

func (process *Process) ProcessState() *ProcessState {
	return &ProcessState{
		Id:             process.Id,
		CurrentEpoch:   process.CurrentEpoch,
		FrequencyEpoch: process.FrequencyEpoch,
		LastVoteEpoch:  process.LastVoteEpoch,
	}
}

var electionId int = 0
var elections []*ElectionHistory = make([]*ElectionHistory, 0)
var historyMutex *sync.Mutex = new(sync.Mutex)

func NewElection(electorId int, requestFrequency int) *ElectionHistory {
	historyMutex.Lock()
	defer historyMutex.Unlock()

	ret := &ElectionHistory{
		Id:               electionId,
		ElectorId:        electorId,
		RequestFrequency: requestFrequency,
		SentMessages:     make([]*Message, 0),
		ReceivedMessages: make([]*ReceivedMessage, 0),
		Mutex:            new(sync.Mutex),
	}

	electionId += 1
	elections = append(elections, ret)
	return ret
}

func (history *ElectionHistory) Sent(message *Message) {
	history.Lock()
	defer history.Unlock()

	history.SentMessages = append(history.SentMessages, message)
}

func (history *ElectionHistory) ReceivedUnsuccessful(message *Message, process *Process) {
	history.Lock()
	defer history.Unlock()

	history.ReceivedMessages = append(
		history.ReceivedMessages,
		&ReceivedMessage{message, process.ProcessState(), false},
	)
}

func (history *ElectionHistory) ReceivedSuccessful(message *Message, process *Process) {
	history.Lock()
	defer history.Unlock()

	history.ReceivedMessages = append(
		history.ReceivedMessages,
		&ReceivedMessage{message, process.ProcessState(), true},
	)
}
