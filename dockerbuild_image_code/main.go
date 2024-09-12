package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"net/url"
	"github.com/gin-gonic/gin"
	"context"
	"log"
	"os"
    secretmanager "cloud.google.com/go/secretmanager/apiv1"
    secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

func accessSecretVersion(secretName string) (string, error) {
    ctx := context.Background()
    client, err := secretmanager.NewClient(ctx)
    if err != nil {
        return "", err
    }

    req := &secretmanagerpb.AccessSecretVersionRequest{
        Name: secretName,
    }

    result, err := client.AccessSecretVersion(ctx, req)
    if err != nil {
        return "", err
    }

    return string(result.Payload.Data), nil
}

const (
	API_KEY            = "projects/alert-flames-286515/secrets/open-weather-api-key/versions/1"
	API_BASE_URL       = "https://api.openweathermap.org/data/2.5/weather?units=imperial&appid="
	OPEN_CAGE_API_KEY  = "projects/alert-flames-286515/secrets/open-cage-api-key/versions/1"
	OPEN_CAGE_BASE_URL = "https://api.opencagedata.com/geocode/v1/json?"
)



func getWeatherInfo(apiKey, openCageAPIKey, location string) (map[string]string, error) {
    geocodeURL := fmt.Sprintf("%sq=%s&key=%s", OPEN_CAGE_BASE_URL, url.QueryEscape(location), openCageAPIKey)
	resp, err := http.Get(geocodeURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var geocodeResult map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&geocodeResult)
	if err != nil {
		return nil, err
	}

	if geocodeResult["results"] == nil || len(geocodeResult["results"].([]interface{})) == 0 {
		return nil, fmt.Errorf("Location not found")
	}

	destination := geocodeResult["results"].([]interface{})[0].(map[string]interface{})
	geometry := destination["geometry"].(map[string]interface{})
	latitude := geometry["lat"].(float64)
	longitude := geometry["lng"].(float64)

	fullAPIURL := fmt.Sprintf("%s%s&lat=%s&lon=%s", API_BASE_URL, apiKey, strconv.FormatFloat(latitude, 'f', -1, 64), strconv.FormatFloat(longitude, 'f', -1, 64))
	resp, err = http.Get(fullAPIURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	weatherInfo := map[string]string{
		"address":     destination["formatted"].(string),
		"coordinates": fmt.Sprintf("(%.4f, %.4f)", latitude, longitude),
		"description": result["weather"].([]interface{})[0].(map[string]interface{})["description"].(string),
		"temperature": fmt.Sprintf("%.1f\u00B0", result["main"].(map[string]interface{})["temp"].(float64)),
	}

	return weatherInfo, nil
}

func main() {
	apiKey, err := accessSecretVersion(API_KEY)
    if err != nil {
        log.Fatalf("Failed to access API_KEY secret: %v", err)
    }

    openCageAPIKey, err := accessSecretVersion(OPEN_CAGE_API_KEY)
    if err != nil {
        log.Fatalf("Failed to access OPEN_CAGE_API_KEY secret: %v", err)
    }
	
	r := gin.Default()

    r.Static("/static", "./static") // Add this line to serve static files

    r.LoadHTMLGlob("templates/*")

    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })

	r.POST("/", func(c *gin.Context) {
		location := c.PostForm("location")
		weatherInfo, err := getWeatherInfo(apiKey, openCageAPIKey, location)
		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{"error": err.Error()})
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{"weather_info": weatherInfo})
	})

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    r.Run(":" + port)
}

