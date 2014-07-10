package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

func RootHandler(w http.ResponseWriter, r *http.Request, processes []*Process) {
	tmpl, err := template.ParseFiles("templates/root.html")
	if err != nil {
		fmt.Printf("Error parsing template")
	}
	if err := tmpl.Execute(w, processes); err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}
}

func ElectionHandler(w http.ResponseWriter, r *http.Request, processes []*Process) {
	processes[1].God <- &Force{Election: &True}
	fmt.Fprintf(w, "Forcing an election")
}

func LagHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := template.ParseFiles("templates/lag.html"); err != nil {
		http.Error(w, "Error parsing lag template", http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Adding lag")
	}
}

func NetworkSplitHandler(writer http.ResponseWriter, request *http.Request, processes []*Process) {
	if err := request.ParseForm(); err != nil {
		http.Error(
			writer,
			fmt.Sprintf("Error parsing split parameters. Err: %v", err),
			http.StatusInternalServerError,
		)
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
}
