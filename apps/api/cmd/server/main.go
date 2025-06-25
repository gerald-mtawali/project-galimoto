// our go server for fetching data from openf1 API
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// a response struct to map response from API
type Response struct {
	Data string `json:"data"`
}

// A lap struct to map Laps to
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

func main() {
	const BASE_URL = "https://api.openf1.org/v1"
	response, err := http.Get(fmt.Sprintf("%s/laps?session_key=9158&driver_number=16&lap_number=1", BASE_URL))
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
