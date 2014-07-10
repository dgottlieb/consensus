package server

import (
	"fmt"
	"html/template"
	"net/http"
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
	r.ParseForm()
	fmt.Println(r.Form)
	for key, _ := range r.Form {
		fmt.Println(key)
	}
	fmt.Println(r.FormValue("0"))
	fmt.Fprintf(w, "Forcing an election")
}

func LagHandler(w http.ResponseWriter, r *http.Request, processes []*Process) {
	if _, err := template.ParseFiles("templates/lag.html"); err != nil {
		http.Error(w, "Error parsing lag template", http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Adding lag")
	}
}
