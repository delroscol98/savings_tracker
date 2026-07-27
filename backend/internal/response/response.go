package response

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type (
	FieldErrors map[string][]string
	ErrorBody   struct {
		Error string `json:"error"`
	}
	ValidationErrorBody struct {
		Error  string      `json:"error"`
		Fields FieldErrors `json:"fields"`
	}
)

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error marshalling JSON: %v", err))
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

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	log.Println(msg)

	params := ErrorBody{
		Error: msg,
	}

	RespondWithJSON(w, code, params)
}

func RespondWithValidationError(w http.ResponseWriter, code int, params ValidationErrorBody) {
	log.Println(params.Error)

	RespondWithJSON(w, code, params)
}
