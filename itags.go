package itags

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	authURL           string = "https://hub.docker.com/v2/users/login/"
	repositoryBaseURL string = "https://hub.docker.com/v2/repositories"
	pageSize          int    = 100
)

// Job - tag request to be made
type Job struct {
	Repository string
}

// JobResult - result of Job
type JobResult struct {
	Repository string
	Tags       []Tag
	Error      error
}

// Tag docker repository tag
type Tag struct {
	Name        string    `json:"name"`
	LastUpdated time.Time `json:"last_updated"`
}

// Credentials for docker registry
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// TagSlice tag slice sortable by name
type TagSlice []Tag

func (t Tag) String() string { return t.Name }

// GetTagsResponse - fields of interest from GET tags REST response
type GetTagsResponse struct {
	Count   int    `json:"count"`
	Next    string `json:"next"`
	Results []Tag
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func sleepMillis(millis int) {
	time.Sleep(time.Duration(millis) * time.Millisecond)
}

// GetTags get tag names for a repository
func GetTags(repository string, httpClient *http.Client, jwt string, numWorkers int) []string {
	return GetTagsForRepository(repository, httpClient, jwt, numWorkers)
}

// GetTagsForRepository get tag names for a repository
func GetTagsForRepository(repository string, httpClient *http.Client, jwt string, numWorkers int) []string {
	tags := GetTagDetails([]string{repository}, httpClient, jwt, numWorkers)[repository]
	answer := make([]string, len(tags))
	for i, tag := range tags {
		answer[i] = tag.Name
	}
	return answer
}

// GetTagsForRepositories get tag names for a list of repositories
// Each tag is prefixed by its repository name
func GetTagsForRepositories(repositories []string, httpClient *http.Client, jwt string, numWorkers int) []string {
	tagsByRepo := GetTagDetails(repositories, httpClient, jwt, numWorkers)
	answer := make([]string, 0, 100)
	for repo, tags := range tagsByRepo {
		for _, tag := range tags {
			answer = append(answer, fmt.Sprintf("%s:%s", repo, tag.Name))
		}
	}
	return answer
}

// GetTagDetails get tags for a list of repostiories
func GetTagDetails(repositories []string, httpClient *http.Client, jwt string, numWorkers int) map[string][]Tag {
	if numWorkers < 1 {
		numWorkers = 1
	}
	tags := make(map[string][]Tag)

	// create channels for jobs to perform and their results
	jobs := make(chan Job, len(repositories))
	results := make(chan JobResult, len(repositories))

	// create pool of workers that receive jobs and send results
	for i := 1; i <= numWorkers; i++ {
		go worker(i, httpClient, jwt, jobs, results)
	}

	// create one job for each repository to query
	for _, r := range repositories {
		jobs <- Job{Repository: r}
	}
	close(jobs)

	// collect the results
	for len(tags) < len(repositories) {
		select {
		case r := <-results:
			tags[r.Repository] = append(tags[r.Repository], r.Tags...)
		default:
			sleepMillis(50)
		}
	}
	return tags
}

func worker(id int, httpClient *http.Client, jwt string, jobs <-chan Job, results chan<- JobResult) {
	for job := range jobs {
		results <- queryTagsForRepository(httpClient, jwt, job)
	}
}

func fakeTagBatch(job Job) JobResult {
	sleepMillis(rand.Intn(1000))

	tags := make([]Tag, pageSize)
	for i := 0; i < pageSize; i++ {
		tags[i] = Tag{Name: fmt.Sprintf("tag-%d", i+1)}
	}
	return JobResult{
		Repository: job.Repository,
		Tags:       tags,
	}
}

// Authenticate and get docker token
func Authenticate(httpClient *http.Client, credentials Credentials) (string, error) {

	body, _ := json.Marshal(credentials)

	req, err := http.NewRequest("POST", authURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 400 {
		body, _ = ioutil.ReadAll(res.Body)
		return "", errors.New("login error: " + string(body))
	}
	loginResponse := new(loginResponse)
	err = json.NewDecoder(res.Body).Decode(loginResponse)
	return loginResponse.Token, err
}

func queryTagsForRepository(httpClient *http.Client, jwt string, job Job) JobResult {
	repository := job.Repository

	// handle "official" docker repos
	if !strings.Contains(repository, "/") {
		repository = "library/" + repository
	}
	url := fmt.Sprintf("%s/%s/tags/?page_size=%d",
		repositoryBaseURL, repository, pageSize)

	var tags []Tag
	for url != "" {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return JobResult{Error: err}
		}

		if jwt != "" {
			req.Header.Add("Authorization", fmt.Sprintf("JWT %s", jwt))
		}
		res, err := httpClient.Do(req)
		if err != nil {
			return JobResult{Error: err}
		}
		defer res.Body.Close()

		tagsResponse := new(GetTagsResponse)
		err = json.NewDecoder(res.Body).Decode(tagsResponse)
		if err != nil {
			return JobResult{Error: err}
		}
		if tags == nil {
			tags = make([]Tag, 0, tagsResponse.Count)
		}
		tags = append(tags, tagsResponse.Results...)
		url = tagsResponse.Next

	}
	return JobResult{
		Repository: job.Repository,
		Tags:       tags,
	}
}
