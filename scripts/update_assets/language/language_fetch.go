package language

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func getToken() string {
	envVariable := "TOKEN"

	value, exists := os.LookupEnv(envVariable)
	if !exists {
		panic(fmt.Sprintf("Environment variable %s is not set", envVariable))
	}

	return value
}

func GithubAuthenticatedGet(url string) []byte {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	token := getToken()

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	return body
}

func CreateRealtimeInformations() []Information {
	// Get all public repositories
	data := GithubAuthenticatedGet("https://api.github.com/users/youdie323323/repos")

	var repos []map[string]any

	err := json.Unmarshal(data, &repos)
	if err != nil {
		panic(err)
	}

	statistics := make(Statistics)

	for _, repo := range repos {
		languageUsagesJson := GithubAuthenticatedGet(repo["languages_url"].(string))
		if bytes.Contains(languageUsagesJson, []byte("Repository access blocked")) {
			continue
		}

		var languageUsages Statistics

		err = json.Unmarshal(languageUsagesJson, &languageUsages)
		if err != nil {
			panic(err)
		}

		// Summate all languages data sizes
		for key, value := range languageUsages {
			if v, ok := statistics[key]; ok {
				statistics[key] = v + value
			} else {
				statistics[key] = value
			}
		}
	}

	return CreateInformationsFromStatistics(statistics)
}
