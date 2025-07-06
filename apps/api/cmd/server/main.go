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

func getLapInfo(BaseUrl string) {
	// const BASE_URL = "https://api.openf1.org/v1"
	response, err := http.Get(fmt.Sprintf("%s/laps?session_key=9158&driver_number=16&lap_number=1", BaseUrl))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer response.Body.Close() // always close response body

	if response.StatusCode != http.StatusOK {
		log.Fatalf("API request failed with status code %d", response.StatusCode)
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var responseInfo any

	err_ := json.Unmarshal(responseData, &responseInfo)

	if err_ != nil {
		fmt.Println(err_.Error())
		os.Exit(1)
	}

	var lapData []Lap

	if err := json.Unmarshal(responseData, &lapData); err != nil {
		log.Fatalf("Error parsing json %s", err.Error())
	}

	fmt.Printf("Received %d lap records", len(lapData))

	for i, lap := range lapData {
		fmt.Printf("\n--- Lap %d ---\n", i+1)
		fmt.Printf("Meeting Key: %d\n", lap.MeetingKey)
		fmt.Printf("Session Key: %d\n", lap.SessionKey)
		fmt.Printf("Driver Number: %d\n", lap.DriverNumber)
		fmt.Printf("Lap Number: %d\n", lap.LapNumber)
		fmt.Printf("Date Start: %s\n", lap.DateStart.Format("2006-01-02 15:04:05"))
		fmt.Printf("Is Pit Out Lap: %t\n", lap.IsPitOutLap)

		// Handle nullable fields safely
		if lap.DurationS1 != nil {
			fmt.Printf("Duration Sector 1: %.3f seconds\n", *lap.DurationS1)
		} else {
			fmt.Println("Duration Sector 1: null")
		}

		if lap.DurationS2 != nil {
			fmt.Printf("Duration Sector 2: %.3f seconds\n", *lap.DurationS2)
		} else {
			fmt.Println("Duration Sector 2: null")
		}

		if lap.DurationS3 != nil {
			fmt.Printf("Duration Sector 3: %.3f seconds\n", *lap.DurationS3)
		} else {
			fmt.Println("Duration Sector 3: null")
		}

		if lap.LapDuration != nil {
			fmt.Printf("Lap Duration: %.3f seconds\n", *lap.LapDuration)
		} else {
			fmt.Println("Lap Duration: null")
		}

		if lap.SpeedI1 != nil {
			fmt.Printf("I1 Speed: %d km/h\n", *lap.SpeedI1)
		} else {
			fmt.Println("I1 Speed: null")
		}

		if lap.SpeedI2 != nil {
			fmt.Printf("I2 Speed: %d km/h\n", *lap.SpeedI2)
		} else {
			fmt.Println("I2 Speed: null")
		}

		if lap.StSpeed != nil {
			fmt.Printf("ST Speed: %d km/h\n", *lap.StSpeed)
		} else {
			fmt.Println("ST Speed: null")
		}

		fmt.Printf("Segments Sector 1: %v\n", lap.SegmentsS1)
		fmt.Printf("Segments Sector 2: %v\n", lap.SegmentsS2)
		fmt.Printf("Segments Sector 3: %v\n", lap.SegmentsS3)
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

// Session Handlers
func sessionsHandler(w http.ResponseWriter, r *http.Request) {
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

func handleGetSession(w http.ResponseWriter, r *http.Request, id int) {
	fmt.Println("Handling our /sessions/:id gets")
}

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Fatal("Error loading .env file: ", err)
	}
	// openF1Url := os.Getenv("OPENF1_API_URL")
	// getLapInfo(openF1Url)

	http.HandleFunc("/sessions", sessionsHandler)
	http.HandleFunc("/sessions/", sessionHandler)
	apiUrl := os.Getenv("API_URL")
	apiPort := os.Getenv("API_PORT")
	fmt.Printf("Server is running at %s:%s\n", apiUrl, apiPort)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
