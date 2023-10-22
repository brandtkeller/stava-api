package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type authResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
}

type activity struct {
	Id          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Distance    float64 `json:"distance"`
	MovingTime  int     `json:"moving_time"`
	ElapsedTime int     `json:"elapsed_time"`
	Type        string  `json:"type"`
	StartDate   string  `json:"start_date"`
	StartTime   string  `json:"start_time"`
	EndDate     string  `json:"end_date"`
	EndTime     string  `json:"end_time"`
}

type envVars struct {
	StravaClientId     string `mapstructure:"STRAVA_CLIENT_ID"`
	StravaClientSecret string `mapstructure:"STRAVA_CLIENT_SECRET"`
	StravaRefreshToken string `mapstructure:"STRAVA_REFRESH_TOKEN"`
}

func main() {

	// setup logging
	logger := log.Default()

	var config envVars
	// Load Config
	viper.SetConfigName("strava")
	viper.AddConfigPath(".")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal(err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		logger.Fatal(err)
	}

	// Create HTTP Client
	client := http.Client{}

	authUrl := "https://www.strava.com/oauth/token"
	activitesUrl := "https://www.strava.com/api/v3/athlete/activities"

	req, err := http.NewRequest("POST", authUrl, strings.NewReader("client_id=115159&client_secret=6e0451fb8dcfb7b4de3a16f56ffab22eb01df0cf&grant_type=refresh_token&refresh_token=d705e4806714d9a00f4a9a33aaeed4550b9fb252&f=json"))
	if err != nil {
		//Handle Error
		logger.Fatal(err)
	}

	q := req.URL.Query()
	q.Add("client_id", config.StravaClientId)
	q.Add("client_secret", config.StravaClientSecret)
	q.Add("refresh_token", config.StravaRefreshToken)
	q.Add("grant_type", "refresh_token")
	q.Add("f", "json")
	req.URL.RawQuery = q.Encode()

	res, err := client.Do(req)
	if err != nil {
		//Handle Error
		logger.Fatal(err)
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		logger.Fatal(readErr)
	}

	var result authResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		logger.Println("Can not unmarshal JSON")
	}

	logger.Println("Authenticated - Preparing to get activities by page of 200")

	activities := make([]activity, 0)
	page := 1

	for {
		pageActivities := make([]activity, 0)
		req, err = http.NewRequest("GET", activitesUrl, nil)
		if err != nil {
			//Handle Error
			logger.Fatal(err)
		}
		q := req.URL.Query()
		q.Add("per_page", "200")
		q.Add("page", strconv.Itoa(page))
		req.URL.RawQuery = q.Encode()

		req.Header = http.Header{
			"Authorization": []string{"Bearer " + result.AccessToken},
		}

		res, err = client.Do(req)
		if err != nil {
			//Handle Error
			logger.Fatal(err)
		}

		body, readErr = io.ReadAll(res.Body)
		if readErr != nil {
			logger.Fatal(readErr)
		}

		if err := json.Unmarshal(body, &pageActivities); err != nil { // Parse []byte to go struct pointer
			logger.Fatal(err)
		}

		if len(pageActivities) == 200 {
			logger.Printf("Page %d retrieved with %d activities\n", page, len(pageActivities))
			page++
			activities = append(activities, pageActivities...)
		} else {
			logger.Printf("Page %d retrieved with %d activities\n", page, len(pageActivities))
			activities = append(activities, pageActivities...)
			break
		}
	}

	logger.Printf("Number of activities: %d\n", len(activities))

	var deskCount int
	var distance float64

	for _, activity := range activities {
		if strings.ToLower(activity.Name) == "desk treadmill" {
			distance += activity.Distance
			deskCount++
		}
	}

	logger.Printf("Desk Treadmill Activities: %d\n", deskCount)
	logger.Printf("Distance: %f Miles\n", distance*0.000621371)

}