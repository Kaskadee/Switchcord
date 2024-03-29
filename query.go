package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// RequestURL to game database. We are using IGDB API to search for Switch games, as Nintendo does not have an official API.
const RequestURL = "https://api-v3.igdb.com/games"

// RequestUserKey is the API key used to authenticate against the IGDB API.
const RequestUserKey = "6bbaebf0dad9ba341b35f204904551c7"

// Game represents a game which was returned by the IGDB API.
type Game struct {
	ID        int    `json:"id"`
	Cover     int    `json:"cover"`
	Platforms []int  `json:"platforms"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
}

// SearchGame searches for Nintendo Switch games which match the search term.
// The request is sent to RequestUri.
func SearchGame(searchTerm string) ([]Game, error) {
	// Query game using the IGDB API.
	query := fmt.Sprintf("fields id,name,category,platforms,slug; where platforms = (130) & category = 0; search \"%s\"; limit 10;", searchTerm)
	data, err := queryRequest(query)
	if err != nil {
		return nil, err
	}

	// Deserialize query response to game slice.
	var games []Game
	err = json.Unmarshal(data, &games)
	if err != nil {
		return nil, fmt.Errorf("json error: %w", err)
	}
	return games, nil
}

func queryRequest(query string) ([]byte, error) {
	// Create HTTP client to send POST request to REST API.
	httpClient := &http.Client{
		Timeout: time.Second * 30, // Set timeout to 30 seconds to prevent application hang.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Do not follow redirects.
		},
	}
	// Configure POST request.
	req, _ := http.NewRequest("POST", RequestURL, strings.NewReader(query))
	req.Header.Set("user-agent", "Switchcord/1.0")
	req.Header.Set("accept", "application/json")
	req.Header.Set("user-key", RequestUserKey)
	// Perform POST request.
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	// Check HTTP response code.
	if resp.StatusCode != 200 {
		// Check if response body is not empty.
		if resp.ContentLength > 0 {
			// Try to parse response body to QueryError.
			var queryErr QueryError
			err = parseResponse(resp, &queryErr)
			if err != nil {
				return nil, NewQueryError(resp.StatusCode, resp.Status, nil)
			}
			return nil, queryErr
		}

		return nil, NewQueryError(resp.StatusCode, resp.Status, nil)
	}

	// Read response body.
	body, err := readResponseBody(resp)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func parseResponse(resp *http.Response, obj interface{}) error {
	// Read response body from HTTP response.
	body, err := readResponseBody(resp)
	if err != nil {
		return err
	}

	// Try to deserialize HTTP response to JSON object.
	err = json.Unmarshal(body, &obj)
	if err != nil {
		return fmt.Errorf("json error: %w", err)
	}
	return nil
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	// Try to read the full body from the HTTP response.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("io error: %w", err)
	}
	return body, resp.Body.Close()
}
