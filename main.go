package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type build struct {
	Branch          string   `json:"branch"`
	BuildURL        string   `json:"build_url"`
	Workflows       workflow `json:"workflows"`
	StartTime       string   `json:"start_time"`
	BuildTimeMillis int      `json:"build_time_millis"`
	Status          string   `json:"status"`
	Lifecycle       string   `json:"lifecycle"`
	BuildNum        int      `json:"build_num"`
	UserName        string   `json:"username"`
	RepoName        string   `json:"reponame"`
	CommitterName   string   `json:"committer_name"`
}

type workflow struct {
	JobName string `json:"job_name"`
}

type Items struct {
	Item []Item `json:"items"`
}

type Item struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Arg      string `json:"arg"`
	Icon     icon   `json:"icon"`
}

type icon struct {
	Path string `json:"path"`
}

func main() {
	var token = flag.String("t", "secret", "CircleCI Token")
	var username = flag.String("u", "", "Username")
	var reponame = flag.String("r", "", "Reponame")
	var limit = flag.Int("l", 30, "Limit")
	var filter = flag.String("f", "", "Search Filter")
	var jobURL = flag.String("j", "", "JobURL (to watch)")
	var watchFlag = flag.Bool("w", false, "Watch job")
	var watchTimeout = flag.Duration("wt", 15*time.Minute, "Watch timeout, default 15m")
	flag.Parse()

	var r []build
	switch {
	case *watchFlag:
		b := watch(*watchTimeout, *token, *jobURL)
		fmt.Printf("Job %s #%d [%s]\nStatus: %s - Outcome: %s", b.RepoName, b.BuildNum, b.Workflows.JobName, b.Lifecycle, b.Status)
		return
	case *username != "" && *reponame != "":
		r = search(*token, *username, *reponame, *limit)
	default:
		r = getRecent(*token, *limit)
	}

	items := filterItems(r, *filter)

	j, err := json.Marshal(Items{Item: items})
	if err != nil {
		fmt.Println("cannot marshal alfred response")
		os.Exit(1)
	}
	fmt.Println(string(j))
}

func getRecent(token string, limit int) []build {
	var b []build
	url := fmt.Sprintf("https://circleci.com/api/v1.1/recent-builds?circle-token=%s&shallow=true&limit=%d", token, limit)
	query(url, &b)
	return b
}

func search(token, user, repository string, limit int) []build {
	var b []build
	url := fmt.Sprintf("https://circleci.com/api/v1.1/project/github/%s/%s?circle-token=%s&shallow=true&limit=%d", user, repository, token, limit)
	query(url, &b)
	return b
}

func watch(timeout time.Duration, token, jobURL string) build {
	var b build
	suffix := strings.Replace(jobURL, "https://circleci.com/gh/", "", -1)
	url := fmt.Sprintf("https://circleci.com/api/v1.1/project/github/%s?circle-token=%s&shallow=true", suffix, token)
	timer := time.NewTimer(timeout)
	for {
		select {
		case <-timer.C:
			b.Status = b.Status + "[TIMEOUT]"
			return b
		default:
			query(url, &b)
			if b.Lifecycle == "finished" {
				return b
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func query(url string, b interface{}) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("error while making the call:", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Println("status code was:", res.Status)
		os.Exit(1)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("error while reading the body:", err)
		os.Exit(1)
	}

	err = json.Unmarshal(body, b)
	if err != nil {
		fmt.Println("cannot unmarshal circleci response")
		os.Exit(1)
	}
}

func filterItems(builds []build, filter string) []Item {
	var items []Item
	for _, v := range builds {
		if strings.Contains(v.RepoName+v.Branch+v.Status+v.CommitterName, filter) {
			title := "#" + fmt.Sprint(v.BuildNum) +
				" / " + v.RepoName +
				" / " + v.Branch

			sec := v.BuildTimeMillis / 1000

			user := v.CommitterName
			if user == "" {
				user = v.UserName
			}

			t, _ := time.Parse(time.RFC3339, v.StartTime)
			subtitle := fmt.Sprintf("[%s] U: %s | Start: %v | Elapsed: %d sec", v.Workflows.JobName, user, t.Format("02/01/2006 3:04PM"), sec)

			var color string
			if v.Status == "no_tests" || v.Status == "not_run" || v.Status == "not_running" {
				color = "gray"
			} else if v.Status == "fixed" || v.Status == "success" {
				color = "green"
			} else if v.Status == "queued" || v.Status == "scheduled" {
				color = "purple"
			} else if v.Status == "canceled" || v.Status == "failed" || v.Status == "infrastructure_fail" || v.Status == "timeout" {
				color = "red"
			} else if v.Status == "retried" || v.Status == "running" {
				color = "blue"
			}

			items = append(items, Item{
				Title:    title,
				Subtitle: subtitle,
				Arg:      v.BuildURL,
				Icon:     icon{Path: color + ".png"}})
		}
	}

	return items
}
