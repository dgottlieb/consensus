package server

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var validLeftStr string
var validRightStr string

func templateMap(processes []*Process) map[string]interface{} {
	frequency := -1
	for _, election := range elections {
		if election.Successful {
			frequency = election.RequestFrequency
		}
	}

	mp := map[string]interface{}{
		"P":     processes,
		"E":     elections,
		"left":  validLeftStr,
		"right": validRightStr,
		"F":     frequency,
	}

	return mp
}

func RootHandler(w http.ResponseWriter, r *http.Request, processes []*Process) {
	tmpl, err := template.ParseFiles("templates/root.html")
	if err != nil {
		fmt.Printf("Error parsing template")
	}
	heatMap(processes)
	if err := tmpl.Execute(w, templateMap(processes)); err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}
}

func ElectionHandler(w http.ResponseWriter, r *http.Request, processes []*Process) {
	if err := r.ParseForm(); err != nil {
		http.Error(
			w,
			fmt.Sprintf("Error parsing election form. Err: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	for key := range r.Form { // should only iterate once
		processId, _ := strconv.Atoi(key)
		processes[processId].God <- &Force{Election: &True}
	}

	http.Redirect(w, r, "/", 302)
}

func LagHandler(w http.ResponseWriter, r *http.Request, processes []*Process) {
	if err := r.ParseForm(); err != nil {
		http.Error(
			w,
			fmt.Sprintf("Error parsing election form. Err: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	var processId int
	for key := range r.Form { // should only iterate once
		processId, _ = strconv.Atoi(key)
		for i := range processes[processId].NetworkState.Lag {
			if i == processId {
				continue
			}
			processes[processId].NetworkState.Lag[i] = time.Duration(10+rand.Int31n(100)) * time.Second
			processes[i].NetworkState.Lag[processId] = time.Duration(10+rand.Int31n(100)) * time.Second
		}
	}
	heatMap(processes)
	http.Redirect(w, r, "/", 302)
}

func heatMap(processes []*Process) {
	toCsv(processes)
	cmd := exec.Command("Rscript", "heatmap.R")
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}

func NetworkSplitHandler(writer http.ResponseWriter, request *http.Request, processes []*Process) {
	if err := request.ParseForm(); err != nil {
		http.Error(
			writer,
			fmt.Sprintf("Error parsing split parameters. Err: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	leftStr := request.FormValue("left")
	rightStr := request.FormValue("right")

	leftIdxStrs := strings.Split(leftStr, ",")
	rightIdxStrs := strings.Split(rightStr, ",")

	if len(leftIdxStrs)+len(rightIdxStrs) != len(processes) {
		http.Error(
			writer,
			fmt.Sprintf("Expected each process id to be either on left or right. Left: %v Right %v",
				leftStr, rightStr),
			http.StatusBadRequest,
		)
		return
	}

	leftIdxs := make([]int, len(leftIdxStrs))
	rightIdxs := make([]int, len(rightIdxStrs))

	var err error
	for idx, str := range leftIdxStrs {
		leftIdxs[idx], err = strconv.Atoi(str)
		if err != nil {
			http.Error(
				writer,
				fmt.Sprintf("Bad left input. Received: `%v`", leftStr),
				http.StatusBadRequest,
			)
			return
		}
	}

	for idx, str := range rightIdxStrs {
		rightIdxs[idx], err = strconv.Atoi(str)
		if err != nil {
			http.Error(
				writer,
				fmt.Sprintf("Bad right input. Received: `%v`", rightStr),
				http.StatusBadRequest,
			)
			return
		}
	}

	for _, left := range leftIdxs {
		for _, right := range rightIdxs {
			processes[left].NetworkState.Packetloss[right] = 100
			processes[right].NetworkState.Packetloss[left] = 100
		}
	}

	validLeftStr = leftStr
	validRightStr = rightStr

	http.Redirect(writer, request, "/", 302)
}

func HealNetworkHandler(writer http.ResponseWriter, request *http.Request, processes []*Process) {
	for left := 0; left < len(processes); left++ {
		for right := 0; right < len(processes); right++ {
			processes[left].NetworkState.Lag[right] = 0
			processes[left].NetworkState.Packetloss[right] = 0
			processes[right].NetworkState.Lag[left] = 0
			processes[right].NetworkState.Packetloss[left] = 0
		}
	}

	validLeftStr = ""
	validRightStr = ""

	heatMap(processes)
	http.Redirect(writer, request, "/", 302)
}

func DisplayElectionHistory(writer http.ResponseWriter, request *http.Request) {
	template, err := template.ParseFiles("templates/history.html")
	if err != nil {
		panic(err)
	}

	electionId, err := strconv.Atoi(request.FormValue("id"))
	if err != nil {
		panic(err)
	}

	history := elections[electionId]
	if err := template.Execute(writer, history); err != nil {
		panic(err)
	}
}
