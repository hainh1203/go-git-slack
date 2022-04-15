package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	GoogleSheetApi string
}

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func loadConfig() *Config {
	conf := Config{}
	content, e := ioutil.ReadFile("./config.json")
	handleError(e)
	err := json.Unmarshal(content, &conf)
	handleError(err)
	return &conf
}

type User struct {
	Email string `json:"email"`
}

type Author struct {
	Email string `json:"email"`
}

type Project struct {
	WebURL            string `json:"web_url"`
	PathWithNamespace string `json:"path_with_namespace"`
}

type LastCommit struct {
	Author Author `json:"author"`
}

type ObjectAttributes struct {
	TargetBranch string     `json:"target_branch"`
	SourceBranch string     `json:"source_branch"`
	URL          string     `json:"url"`
	LastCommit   LastCommit `json:"last_commit"`
	Title        string     `json:"title"`
	State        string     `json:"state"`
	Action       string     `json:"action"`
}

type GitlabPayload struct {
	ObjectKind       string           `json:"object_kind"`
	EventType        string           `json:"event_type"`
	User             User             `json:"user"`
	ObjectAttributes ObjectAttributes `json:"object_attributes"`
	Project          Project          `json:"project"`
}

type Content struct {
	Channels map[string]string `json:"channels"`
	Mentions map[string]string `json:"mentions"`
}

type GoogleSheetResponse struct {
	Result  string  `json:"result"`
	Content Content `json:"content"`
}

func getDataGoogleSheet(api string) GoogleSheetResponse {
	resp, err := http.Get(api)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return GoogleSheetResponse{}
	}

	var googleSheetResponse GoogleSheetResponse
	err = json.Unmarshal(body, &googleSheetResponse)
	if err != nil {
		log.Fatalln(err)
		return GoogleSheetResponse{}
	}

	return googleSheetResponse
}

func makeMessage(
	repository string,
	mergedBy string,
	branchBase string,
	branchHead string,
	author string,
	commit string,
	pull string,
	mention string,
) string {
	if mergedBy != author && mention != "" {
		author = "<@" + mention + ">"
	}

	return ":tada::tada: *MERGED* :tada::tada: " +
		"\n • Repository: " + repository +
		"\n • Branch: `" + branchHead + "` into `" + branchBase + "`" +
		"\n • Merged By: " + mergedBy +
		"\n • Author: " + author +
		"\n • Title: " + commit +
		"\n • Pull Request: <" + pull + "|Click here>"
}

func sendMessageToSlack(webhook, message string) {
	u, _ := url.ParseRequestURI(webhook)
	urlStr := fmt.Sprintf("%v", u)

	values := map[string]string{"text": message}
	jsonStr, _ := json.Marshal(values)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonStr))

	resp, err := client.Do(r)
	handleError(err)

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
}

func main() {

	http.HandleFunc("/gitlab", func(w http.ResponseWriter, r *http.Request) {
		config := loadConfig()

		var payload GitlabPayload

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if payload.ObjectKind == "merge_request" && payload.EventType == "merge_request" {
			if payload.ObjectAttributes.State == "merged" && payload.ObjectAttributes.Action == "merge" {
				data := getDataGoogleSheet(config.GoogleSheetApi)

				slackUrl := data.Content.Channels[payload.Project.WebURL]

				if slackUrl != "" {
					message := makeMessage(
						payload.Project.PathWithNamespace,
						payload.User.Email,
						payload.ObjectAttributes.TargetBranch,
						payload.ObjectAttributes.SourceBranch,
						payload.ObjectAttributes.LastCommit.Author.Email,
						payload.ObjectAttributes.Title,
						payload.ObjectAttributes.URL,
						data.Content.Mentions[payload.ObjectAttributes.LastCommit.Author.Email],
					)
					sendMessageToSlack(slackUrl, message)
				}
			}
		}
	})

	port := "8080"

	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) > 0 {
		port = argsWithoutProg[0]
	}

	fmt.Println("http://localhost:" + port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		return
	}
}
