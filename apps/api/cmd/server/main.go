// our go server for fetching data from openf1 API
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// a response struct to map response from API

type Session struct {
	CircuitKey       int    `json:"circuit_key"`
	CircuitShortName string `json:"circuit_short_name"`
	CountryCode      string `json:"country_code"`
	CountryKey       int    `json:"country_key"`
	CountryName      string `json:"country_name"`
	DateEnd          string `json:"date_end"`
	DateStart        string `json:"date_start"`
	Location         string `json:"location"`
	MeetingKey       int    `json:"meeting_key"`
	SessionKey       int    `json:"session_key"`
	SessionName      string `json:"session_name"`
	SessionType      string `json:"session_type"`
	Year             int    `json:"year"`
}

var (
	sessionsCache  = make(map[int]Session)
	nextSessionKey int
	sessionsMu     sync.Mutex
)

type Lap struct {
	MeetingKey   int       `json:"meeting_key"`
	SessionKey   int       `json:"session_key"`
	DriverNumber int       `json:"driver_number"`
	LapNumber    int       `json:"lap_number"`
	DateStart    time.Time `json:"date_start"`
	DurationS1   *float64  `json:"duration_sector_1"`
	DurationS2   *float64  `json:"duration_sector_2"`
	DurationS3   *float64  `json:"duration_sector_3"`
	SpeedI1      *int      `json:"i1_speed"`
	SpeedI2      *int      `json:"i2_speed"`
	IsPitOutLap  bool      `json:"is_pit_out_lap"`
	LapDuration  *float64  `json:"lap_duration"`
	SegmentsS1   []int     `json:"segments_sector_1"`
	SegmentsS2   []int     `json:"segments_sector_2"`
	SegmentsS3   []int     `json:"segments_sector_3"`
	StSpeed      *int      `json:"st_speed"`
}

// while we're able to use the multiple instances, we'll need a more improved way to perform operator overloading
type PaginationConfig struct {
	Skip          int
	Limit         int
	HasPagination bool
}

type KeyOnlyConfig struct {
	KeysOnly    string
	HasKeysOnly bool
}

// Helper Functions
func ParsePaginationFromRequest(r *http.Request) (PaginationConfig, error) {
	// why is everything only a single letter
	// if there is a skip and limit we return it as a config element
	skipStr := r.URL.Query().Get("skip")
	limitStr := r.URL.Query().Get("limit")

	config := PaginationConfig{}
	if skipStr == "" && limitStr == "" {
		return config, nil
	}
	// now we need to actually parse these values
	config.HasPagination = true // we know that there is pagination in the request URL
	if skipStr != "" {
		// set the value
		skip, err := parseIntParam(skipStr, 0)
		if err != nil || skip < 0 {
			return config, fmt.Errorf("invalid skip parameter")
		}
		config.Skip = skip
	}

	if limitStr != "" {
		// set the value
		// liimit
		limit, err := parseIntParam(limitStr, 100)
		if err != nil || limit < 0 {
			return config, fmt.Errorf("invalid limit parameter")
		}
		config.Limit = limit
	}
	return config, nil
}

func ParseKeysOnly (r *http.Request) (KeyOnlyConfig, error){
	keyStr := r.URL.Query().Get("keysOnly")
	KeyConfig := KeyOnlyConfig{}
	
	if keyStr == "" {
		return KeyConfig, nil 
	}
	KeyConfig.HasKeysOnly = true 
	
	if KeyStr != "" {
		
		
	}
	
}

func fetchSessions(BaseUrl string) ([]Session, error) {
	if BaseUrl == "" {
		fmt.Println("BaseUrl is empty, unable to make request")
		return nil, fmt.Errorf("baseUrl is empty, unable to make request")
	}
	response, urlErr := http.Get(fmt.Sprintf("%s/sessions", BaseUrl))
	if urlErr != nil {
		fmt.Println("Error fetching sessions: ", urlErr)
		return nil, fmt.Errorf("error fetching sessions %d", urlErr)
	}
	// always make sure to close the response body
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		log.Fatalf("API Response Status Code: %d", response.StatusCode)
		return nil, fmt.Errorf("Response Status Code: %d", response.StatusCode)
	}

	responseData, responseErr := io.ReadAll(response.Body)
	if responseErr != nil {
		fmt.Println("Error reading response body: ", responseErr)
		return nil, fmt.Errorf("error reading response body: %s", responseErr)
	}

	var sessions []Session
	err := json.Unmarshal(responseData, &sessions)
	if err != nil {
		fmt.Println("Error unmarshalling response data: ", err)
		return nil, fmt.Errorf("error unmarshalling response data: %s", err)
	}
	return sessions, nil
}

func FormatSessions(sessions []Session) string {
	var formattedSessions string
	for _, session := range sessions {
		formattedSessions += fmt.Sprintf("Session ID: %d, Circuit Name: %s\n", session.SessionKey, session.CircuitShortName)
	}
	return formattedSessions
}

func FindSessionById(sessions []Session, id int) *Session {
	var session *Session
	for _, s := range sessions {
		if s.SessionKey == id {
			session = &s
			return session
		}
	}
	return nil
}

// Session Handlers
func sessionsHandler(w http.ResponseWriter, r *http.Request) {
	// extract optional parameters skip & limit
	// these are optional parameters
	switch r.Method {
	case http.MethodGet:
		handleGetSessions(w, r)
	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

func sessionHandler(w http.ResponseWriter, r *http.Request) {
	// extract the sessionKey from the URL path
	id, err := strconv.Atoi(r.URL.Path[len("/sessions/"):])
	if err != nil {
		http.Error(w, "Invalid session Key Id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		handleGetSession(w, r, id)
	default:
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
	}
}

// business logic of the handler methods
func handleGetSessions(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Handling our /sessions gets")
	// parse the query parameters
	PageConfig, err := ParsePaginationFromRequest(r)
	if err != nil {
		http.Error(w, "invalid query parameters", http.StatusBadRequest)
		return
	}

	if PageConfig.HasPagination {
		handleGetSessionsWithPagination(w, r, PageConfig.Skip, PageConfig.Limit)
		return
	}
	// if the skip and limit are empty strings then we just return without pagination
	handleSessionsNoPagination(w, r)
}

func parseIntParam(param string, defaultValue int) (int, error) {
	// parse query params as ints
	if param == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(param)
}

func handleSessionsNoPagination(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// fetch our session data from our openf1 api
	openF1Url := os.Getenv("OPENF1_API_URL")

	sessions, err := fetchSessions(openF1Url)
	if err != nil {
		http.Error(w, "Error fetching sessions", http.StatusInternalServerError)
		return
	}

	// An optional operation is to cache the session info
	sessionsMu.Lock()
	// defer sessionsMu.Unlock()
	for _, session := range sessions {
		sessionsCache[session.SessionKey] = session
	}
	sessionsMu.Unlock()
	// encode and send response
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		log.Printf("Error encoding sessions: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

func handleGetSessionsWithPagination(w http.ResponseWriter, r *http.Request, skip int, limit int) {
	// this function is responsible for fetching sessions with pagination
	log.Print("Getting sessions with pagination")
	if skip < 0 || limit < 0 {
		http.Error(w, "Invalid pagination parameters", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	openF1Url := os.Getenv("OPENF1_API_URL")
	sessions, err := fetchSessions(openF1Url)
	if err != nil {
		http.Error(w, "Error fetching session data", http.StatusInternalServerError)
		return
	}
	if skip > len(sessions) {
		http.Error(w, "Invalid pagination parameters", http.StatusBadRequest)
		return
	}

	// starting index is skip
	// limit is the number we obtain
	startIndex := skip
	endIndex := min(startIndex+limit, len(sessions))
	sessions = sessions[startIndex:endIndex]
	if err := json.NewEncoder(w).Encode(sessions); err != nil {
		log.Print("Error encoding sessions: ", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func handleGetSession(w http.ResponseWriter, r *http.Request, id int) {
	fmt.Println("Handling our /sessions/:id gets")
	// fetch the sessions
	openF1Url := os.Getenv("OPENF1_API_URL")

	sessions, err := fetchSessions(openF1Url)
	if err != nil {
		http.Error(w, "Error fetching Sessions", http.StatusInternalServerError)
		return
	}
	session := FindSessionById(sessions, id)

	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// encode and send a response
	if err := json.NewEncoder(w).Encode(session); err != nil {
		log.Printf("Error encoding session: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
	}
}

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	http.HandleFunc("/sessions", sessionsHandler)
	http.HandleFunc("/sessions/", sessionHandler)
	apiUrl := os.Getenv("API_URL")
	apiPort := os.Getenv("API_PORT")
	fmt.Printf("Server is running at %s:%s\n", apiUrl, apiPort)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
