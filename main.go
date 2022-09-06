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
	"strings"
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
	Email    string `json:"email"`
	Username string `json:"username"`
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

type Channel struct {
	SlackUrl string `json:"slack_url"`
	Actions  string `json:"actions"`
}

type Content struct {
	Channels map[string]Channel `json:"channels"`
	Mentions map[string]string  `json:"mentions"`
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

func makeMergedMessage(
	repository string,
	mergedBy string,
	branchBase string,
	branchHead string,
	author string,
	commit string,
	pull string,
) string {
	return ":tada::tada: *MERGED* :tada::tada: " +
		"\n • Repository: " + repository +
		"\n • Branch: `" + branchHead + "` into `" + branchBase + "`" +
		"\n • Merged By: " + mergedBy +
		"\n • Author: " + author +
		"\n • Title: " + commit +
		"\n • Pull Request: <" + pull + "|Click here>"
}

func makeOpenedMessage(
	repository string,
	branchBase string,
	branchHead string,
	author string,
	commit string,
	pull string,
) string {
	return ":alphabet-yellow-p::alphabet-yellow-l::alphabet-yellow-e::alphabet-yellow-a::alphabet-yellow-s::alphabet-yellow-e::alphabet-white-r::alphabet-white-e::alphabet-white-v::alphabet-white-i::alphabet-white-e::alphabet-white-w: " +
		"\n • Repository: " + repository +
		"\n • Branch: `" + branchHead + "` into `" + branchBase + "`" +
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

func getMentionByEmail(mentions map[string]string, email string) string {
	var mention = mentions[email]

	if mention != "" {
		return "<@" + mention + ">"
	}

	return email
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
			data := getDataGoogleSheet(config.GoogleSheetApi)

			channel := data.Content.Channels[payload.Project.WebURL]

			if channel.SlackUrl != "" {
				if payload.ObjectAttributes.State == "merged" &&
					payload.ObjectAttributes.Action == "merge" &&
					strings.Contains(channel.Actions, payload.ObjectAttributes.Action) {
					sendMessageToSlack(channel.SlackUrl, makeMergedMessage(
						payload.Project.PathWithNamespace,
						payload.User.Username,
						payload.ObjectAttributes.TargetBranch,
						payload.ObjectAttributes.SourceBranch,
						getMentionByEmail(data.Content.Mentions, payload.ObjectAttributes.LastCommit.Author.Email),
						payload.ObjectAttributes.Title,
						payload.ObjectAttributes.URL,
					))
				}

				if payload.ObjectAttributes.State == "opened" &&
					payload.ObjectAttributes.Action == "open" &&
					strings.Contains(channel.Actions, payload.ObjectAttributes.Action) {
					sendMessageToSlack(channel.SlackUrl, makeOpenedMessage(
						payload.Project.PathWithNamespace,
						payload.ObjectAttributes.TargetBranch,
						payload.ObjectAttributes.SourceBranch,
						getMentionByEmail(data.Content.Mentions, payload.ObjectAttributes.LastCommit.Author.Email),
						payload.ObjectAttributes.Title,
						payload.ObjectAttributes.URL,
					))
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
