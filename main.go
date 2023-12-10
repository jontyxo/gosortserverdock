// main.go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestPayload struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponsePayload struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

func processSingle(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	sortedArrays := make([][]int, len(payload.ToSort))

	for i, arr := range payload.ToSort {
		sortedArr := make([]int, len(arr))
		copy(sortedArr, arr)
		sort.Ints(sortedArr)
		sortedArrays[i] = sortedArr
	}

	timeTaken := time.Since(startTime).Nanoseconds()

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func processConcurrent(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	var wg sync.WaitGroup
	wg.Add(len(payload.ToSort))

	var mu sync.Mutex
	sortedArrays := make([][]int, len(payload.ToSort))

	for i, arr := range payload.ToSort {
		go func(index int, array []int) {
			defer wg.Done()
			sortedArr := make([]int, len(array))
			copy(sortedArr, array)
			sort.Ints(sortedArr)

			mu.Lock()
			sortedArrays[index] = sortedArr
			mu.Unlock()
		}(i, arr)
	}

	wg.Wait()
	timeTaken := time.Since(startTime).Nanoseconds()

	response := ResponsePayload{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/process-single", processSingle)
	http.HandleFunc("/process-concurrent", processConcurrent)

	port := 8000
	fmt.Printf("Server listening on port %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
