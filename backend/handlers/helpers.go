package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ErrorBody struct {
	Error string `json:"error"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling JSON: %v", err))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	if code != http.StatusNoContent && code != http.StatusNotModified {
		num, err := w.Write(data)
		if err != nil {
			log.Printf("Error writing body: %v. Wrote %d bytes out of %d", err, num, len(data))
			return
		}
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	log.Println(msg)

	params := ErrorBody{
		Error: msg,
	}

	respondWithJSON(w, code, params)
}
