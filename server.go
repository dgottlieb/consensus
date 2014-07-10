package main

import (
	"net/http"
	"html/template"
	"net/http"
	"net/url"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/root.html")
	if err != nil {
		fmt.Printf("Error parsing template")
	}
	tmpl.Execute(w, nil)
}

func electionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Forcing election")
}

func lagHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := template.ParseFiles("templates/lag.html"); err != nil {
		http.Error(w, "Error parsing lag template", http.StatusInternalServerError)
	} else {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing lag", http.StatusInternalServerError)
			return
		}
		lag, err := getFormValues(&r.Form)
		if err != nil {
			http.Error(w, "Unable to get form values", http.StatusInternalServerError)
		}
	//	t.Execute(w, userInput)
		fmt.Fprintf(w, "Adding %s lag", lag)
	}
}

func getFormValues(form *url.Values) (lag string, err error) {
	for key, value := range *form {
		switch key {
		case "lag":
			return value[0], nil
		default:
			return "", fmt.Errorf("Unable to parse form")
		}
	}
	return "", fmt.Errorf("No form values")
}


