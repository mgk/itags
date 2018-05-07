package itags

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	RepositoryBaseURL string = "https://hub.docker.com/v2/repositories"
	PageSize          int    = 100
)

// Job - tag request to be made
type Job struct {
	Repository string
	Page       int
}

// JobResult - result of Job
type JobResult struct {
	Repository string
	Page       int
	Count      int
	Tags       []Tag
	Error      error
}

// Tag docker repository tag
type Tag struct {
	Name        string    `json:"name"`
	LastUpdated time.Time `json:"last_updated"`
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

// GetTags get tag names for a repostiory
func GetTags(repository string, httpClient *http.Client, numWorkers int) []string {
	return GetTagsForRepository(repository, httpClient, numWorkers)
}

// GetTagsForRepository get tag names for a repository
func GetTagsForRepository(repository string, httpClient *http.Client, numWorkers int) []string {
	tags := GetTagDetails([]string{repository}, httpClient, numWorkers)[repository]
	// log.Printf("tags --> %v\n", tags)
	answer := make([]string, len(tags))
	for i, tag := range tags {
		answer[i] = tag.Name
	}
	return answer
}

// GetTagsForRepositories get tag names for a list of repositories
// Each tag is prefixed by its repository name
func GetTagsForRepositories(repositories []string, httpClient *http.Client, numWorkers int) []string {
	tagsByRepo := GetTagDetails(repositories, httpClient, numWorkers)
	answer := make([]string, 0, 100)
	for repo, tags := range tagsByRepo {
		for _, tag := range tags {
			answer = append(answer, fmt.Sprintf("%s:%s", repo, tag.Name))
		}
	}
	return answer
}

// GetTagDetails get tags for a list of repostiories
func GetTagDetails(repositories []string, httpClient *http.Client, numWorkers int) map[string][]Tag {
	if numWorkers < 1 {
		numWorkers = 1
	}
	tags := make(map[string][]Tag)

	jobs := make(chan Job, len(repositories))
	firstResults := make(chan JobResult, len(repositories))
	for i := 1; i <= numWorkers; i++ {
		go worker(httpClient, jobs, firstResults)
	}

	for _, r := range repositories {
		jobs <- Job{Repository: r, Page: 1}
	}
	close(jobs)

	batches := make([]Job, 0, len(repositories))
	for range repositories {
		r := <-firstResults
		if r.Error != nil {
			fmt.Printf("%v", r.Error)
		}
		tags[r.Repository] = append(tags[r.Repository], r.Tags...)
		pages := (r.Count - 1) / PageSize
		for i := 0; i < pages; i++ {
			batches = append(batches, Job{Repository: r.Repository, Page: i + 2})
		}
	}

	results := make(chan JobResult, min(len(batches), 1000))
	jobs = make(chan Job, min(len(batches), 1000))
	for i := 1; i <= numWorkers; i++ {
		go worker(httpClient, jobs, results)
	}

	for _, b := range batches {
		select {
		case jobs <- b:
		default:
			sleepMillis(20)
		}
	}
	close(jobs)

	for range batches {
		r := <-results
		tags[r.Repository] = append(tags[r.Repository], r.Tags...)
	}
	return tags
}

func worker(httpClient *http.Client, jobs <-chan Job, results chan<- JobResult) {
	for job := range jobs {
		// results <- fakeTagBatch(job)
		results <- tagBatch(httpClient, job)
	}
}

func fakeTagBatch(job Job) JobResult {
	sleepMillis(rand.Intn(1000))

	tags := make([]Tag, PageSize)
	for i := 0; i < PageSize; i++ {
		tags[i] = Tag{Name: fmt.Sprintf("tag-%d", i+1)}
	}
	var count int
	if job.Page == 1 {
		count = rand.Intn(130)
	}
	return JobResult{
		Repository: job.Repository,
		Page:       job.Page,
		Count:      count,
		Tags:       tags,
	}
}

var jwt = os.Getenv("DOCKER_TOKEN")

func tagBatch(httpClient *http.Client, job Job) JobResult {
	repository := job.Repository
	if !strings.Contains(repository, "/") {
		repository = "library/" + repository
	}
	url := fmt.Sprintf("%s/%s/tags/?page=%d&page_size=%d",
		RepositoryBaseURL, repository, job.Page, PageSize)

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
	return JobResult{
		Repository: job.Repository,
		Page:       job.Page,
		Count:      tagsResponse.Count,
		Tags:       tagsResponse.Results,
	}
}