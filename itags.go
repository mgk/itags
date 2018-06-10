package itags

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	authURL           string = "https://hub.docker.com/v2/users/login/"
	repositoryBaseURL string = "https://hub.docker.com/v2/repositories"
	pageSize          int    = 100
)

// A Tag represents a docker repository tag
type Tag struct {
	Repo        string
	Name        string    `json:"name"`
	LastUpdated time.Time `json:"last_updated"`
}

// ByRepoAndName - sort tags by repo and name
type ByRepoAndName []Tag

func (tags ByRepoAndName) Len() int      { return len(tags) }
func (tags ByRepoAndName) Swap(i, j int) { tags[i], tags[j] = tags[j], tags[i] }
func (tags ByRepoAndName) Less(i, j int) bool {
	if tags[i].Repo == tags[j].Repo {
		return tags[i].Name < tags[j].Name
	}
	return tags[i].Repo < tags[j].Repo
}

// ByLastUpdated - sort tags by last updated
type ByLastUpdated []Tag

func (tags ByLastUpdated) Len() int           { return len(tags) }
func (tags ByLastUpdated) Swap(i, j int)      { tags[i], tags[j] = tags[j], tags[i] }
func (tags ByLastUpdated) Less(i, j int) bool { return tags[i].LastUpdated.Before(tags[j].LastUpdated) }

// Credentials is the username and password to
// login to a docker registry
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type jobRequest struct {
	Repository string
}

type jobResult struct {
	Repository string
	Tags       []Tag
	Error      error
}

type getTagsResponse struct {
	Count   int    `json:"count"`
	Next    string `json:"next"`
	Results []Tag
}

// GetTags gets tag names for a list of repositories
// Returns tag names with each tag prefixed by its repository name
func GetTags(repository string, httpClient *http.Client, jwt string, numWorkers int) []Tag {
	return GetTagsForRepository(repository, httpClient, jwt, numWorkers)
}

// GetTagsForRepository gets tag names for a repositorie
// Returns tag names found in the repository
func GetTagsForRepository(repository string, httpClient *http.Client, jwt string, numWorkers int) []Tag {
	return GetTagDetails([]string{repository}, httpClient, jwt, numWorkers)[repository]
}

// GetTagsForRepositories gets tag names for a list of repositories
// Returns tag names with each tag prefixed by its repository name
func GetTagsForRepositories(repositories []string, httpClient *http.Client, jwt string, numWorkers int) []Tag {
	tagsByRepo := GetTagDetails(repositories, httpClient, jwt, numWorkers)
	answer := make([]Tag, 0, 100)
	for _, tags := range tagsByRepo {
		for _, tag := range tags {
			answer = append(answer, tag)
		}
	}
	return answer
}

// GetTagDetails returns tags for a list of repostiories
func GetTagDetails(repositories []string, httpClient *http.Client, jwt string, numWorkers int) map[string][]Tag {
	if numWorkers < 1 {
		numWorkers = 1
	}
	tags := make(map[string][]Tag)

	// create channels for jobs to perform and their results
	jobs := make(chan jobRequest, len(repositories))
	results := make(chan jobResult, len(repositories))

	// create pool of workers that receive jobs and send results
	for i := 1; i <= numWorkers; i++ {
		go worker(i, httpClient, jwt, jobs, results)
	}

	// create one job request for each repository to query
	for _, r := range repositories {
		jobs <- jobRequest{Repository: r}
	}
	close(jobs)

	// collect the results
	for len(tags) < len(repositories) {
		select {
		case r := <-results:
			for i := range r.Tags {
				r.Tags[i].Repo = r.Repository
			}
			tags[r.Repository] = append(tags[r.Repository], r.Tags...)
		default:
			sleepMillis(50)
		}
	}
	return tags
}

// Authenticate returns a docker token
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

func worker(id int, httpClient *http.Client, jwt string,
	jobs <-chan jobRequest, results chan<- jobResult) {

	for job := range jobs {
		results <- queryTagsForRepository(httpClient, jwt, job)
	}
}

func queryTagsForRepository(httpClient *http.Client, jwt string, job jobRequest) jobResult {
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
			return jobResult{Error: err}
		}

		if jwt != "" {
			req.Header.Add("Authorization", fmt.Sprintf("JWT %s", jwt))
		}
		res, err := httpClient.Do(req)
		if err != nil {
			return jobResult{Error: err}
		}
		defer res.Body.Close()

		tagsResponse := new(getTagsResponse)
		err = json.NewDecoder(res.Body).Decode(tagsResponse)
		if err != nil {
			return jobResult{Error: err}
		}
		if tags == nil {
			tags = make([]Tag, 0, tagsResponse.Count)
		}
		tags = append(tags, tagsResponse.Results...)
		url = tagsResponse.Next

	}
	return jobResult{
		Repository: job.Repository,
		Tags:       tags,
	}
}

func sleepMillis(millis int) {
	time.Sleep(time.Duration(millis) * time.Millisecond)
}
