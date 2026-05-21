package handlers

import (
	"encoding/json"
	"net/http"
)

type MockSettingsHandler struct {
	activeYear int
}

func NewMockSettingsHandler() *MockSettingsHandler {
	return &MockSettingsHandler{activeYear: 2570}
}

func (h *MockSettingsHandler) GetActiveYear(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]int{"active_year": h.activeYear})
}

func (h *MockSettingsHandler) SetActiveYear(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ActiveYear int `json:"active_year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	h.activeYear = body.ActiveYear
	respond(w, http.StatusOK, map[string]int{"active_year": h.activeYear})
}
