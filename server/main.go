package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type point struct {
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
	Width int     `json:"width"`
	Color string  `json:"color"`
}

type line struct {
	Points []point `json:"points"`
}

var maxNumberOfPoints = 100000

var (
	numberOfPoints = 0
	lineOffset     = 0
	lines          = make([]line, 0)
	linesMutex     sync.RWMutex
)

func addPoints(w http.ResponseWriter, r *http.Request) {
	type reqbody struct {
		Lines []line `json:"lines"`
	}

	var req reqbody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println("add json decoding error:", err)
		http.Error(w, "wrong format", http.StatusNotAcceptable)
		return
	}

	linesMutex.Lock()

	for _, line := range req.Lines {
		numberOfPoints += len(line.Points)
	}

	if numberOfPoints > maxNumberOfPoints {
		amountToDelete := len(lines) / 10

		for i := 0; i < amountToDelete; i++ {
			numberOfPoints -= len(lines[i].Points)
		}
		lineOffset += amountToDelete
		lines = lines[amountToDelete:]
	}

	lines = append(lines, req.Lines...)
	linesMutex.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	type reqbody struct {
		First int `json:"first"`
	}
	type respbody struct {
		Last  int    `json:"last"`
		Lines []line `json:"lines"`
	}

	var req reqbody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println("get json decoding error:", err)
		http.Error(w, "wrong format", http.StatusNotAcceptable)
		return
	}

	linesMutex.RLock()
	defer linesMutex.RUnlock()

	req.First -= lineOffset
	if req.First < 0 {
		req.First = 0
	}

	var resp respbody
	resp.Last = len(lines)

	if req.First < len(lines) {
		resp.Lines = append([]line(nil), lines[req.First:]...)
	} else {
		resp.Lines = []line{}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Println("get json encoding error:", err)
		http.Error(w, "json encoding", http.StatusInternalServerError)
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Dynamically allow whatever origin is making the request (localhost, IP address, etc.)
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle the browser's preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass the request to your actual handler
		next(w, r)
	}
}

func main() {
	http.HandleFunc("/add", enableCORS(addPoints))
	http.HandleFunc("/get", enableCORS(getPoints))

	fmt.Println("Listening on 0.0.0.0:8001")
	err := http.ListenAndServe("0.0.0.0:8001", nil)
	if err != nil {
		fmt.Println("Error:", err.Error())
	}
}
