package itags

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
)

var numWorkers = 20

func mockHTTP(fixture string) (*recorder.Recorder, *http.Client) {
	tape, err := recorder.New("fixtures/" + fixture)
	if err != nil {
		log.Fatalln(err)
	}
	httpClient := &http.Client{Timeout: 5 * time.Second, Transport: tape}
	if err != nil {
		log.Fatalln(err)
	}
	return tape, httpClient
}

var requestAndBodyMatcher = func(r *http.Request, i cassette.Request) bool {
	var b bytes.Buffer
	if _, err := b.ReadFrom(r.Body); err != nil {
		return false
	}
	r.Body = ioutil.NopCloser(&b)
	return cassette.DefaultMatcher(r, i) && (b.String() == "" || b.String() == i.Body)
}

func assertEqual(t *testing.T, message string, expected, actual interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("%s: expected: %v, actual: %v", message, expected, actual)
	}
}

func TestGetTagsSmallOfficialRepository(t *testing.T) {
	repo := "hello-world"
	tape, http := mockHTTP("hello-world")
	defer tape.Stop()
	tags := GetTags(repo, http, "", numWorkers)
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	sort.Strings(names)

	assertEqual(t, "hello world",
		[]string{
			"latest",
			"linux",
			"nanoserver",
			"nanoserver-1709",
			"nanoserver-sac2016",
			"nanoserver1709",
		}, names)
}

func TestGetTagsSmallOneWorkerIsMinimum(t *testing.T) {
	repo := "hello-world"
	tape, http := mockHTTP("hello-world")
	defer tape.Stop()
	tags := GetTags(repo, http, "", 0)
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	sort.Strings(names)

	assertEqual(t, "hello world",
		[]string{
			"latest",
			"linux",
			"nanoserver",
			"nanoserver-1709",
			"nanoserver-sac2016",
			"nanoserver1709",
		}, names)
}

func TestGetTagsSmallUnofficialRepository(t *testing.T) {
	tape, http := mockHTTP("figlet")
	defer tape.Stop()

	tags := GetTags("mgkio/figlet", http, "", numWorkers)
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	sort.Strings(names)

	assertEqual(t, "figlet", []string{"1", "latest"}, names)
}

// GetTags is an alias for GetTagsForRepository
func TestGetTagsForRepository(t *testing.T) {
	tape, http := mockHTTP("hello-world")
	defer tape.Stop()

	tags := GetTagsForRepository("hello-world", http, "", numWorkers)
	sort.Sort(ByRepoAndName(tags))
	names := make([]string, len(tags))
	for i, tag := range tags {
		assertEqual(t, "tag repo", "hello-world", tag.Repo)
		names[i] = tag.Name
	}
	assertEqual(t, "hello world",
		[]string{
			"latest",
			"linux",
			"nanoserver",
			"nanoserver-1709",
			"nanoserver-sac2016",
			"nanoserver1709",
		}, names)
}

func TestGetTagsMultiplePages(t *testing.T) {
	tape, http := mockHTTP("redis")
	defer tape.Stop()

	tags := GetTags("redis", http, "", numWorkers)
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	sort.Strings(names)

	assertEqual(t, "redis", []string{
		"2", "2-32bit", "2.6", "2.6-32bit", "2.6.17", "2.6.17-32bit", "2.8",
		"2.8-32bit", "2.8.10", "2.8.11", "2.8.12", "2.8.13", "2.8.14",
		"2.8.15", "2.8.16", "2.8.17", "2.8.18", "2.8.19", "2.8.20", "2.8.21",
		"2.8.21-32bit", "2.8.22", "2.8.22-32bit", "2.8.23", "2.8.23-32bit", "2.8.6",
		"2.8.7", "2.8.8", "2.8.9", "3", "3-32bit", "3-alpine", "3-nanoserver",
		"3-windowsservercore", "3.0", "3.0-32bit", "3.0-alpine", "3.0-nanoserver",
		"3.0-windowsservercore", "3.0.0", "3.0.1", "3.0.2", "3.0.2-32bit", "3.0.3",
		"3.0.3-32bit", "3.0.4", "3.0.4-32bit", "3.0.5", "3.0.5-32bit", "3.0.504-nanoserver",
		"3.0.504-windowsservercore", "3.0.6", "3.0.6-32bit", "3.0.6-alpine", "3.0.7",
		"3.0.7-32bit", "3.0.7-alpine", "3.0.7-nanoserver", "3.0.7-windowsservercore", "3.2",
		"3.2-32bit", "3.2-alpine", "3.2-nanoserver", "3.2-windowsservercore", "3.2.0",
		"3.2.0-32bit", "3.2.0-alpine", "3.2.1", "3.2.1-32bit", "3.2.1-alpine", "3.2.10",
		"3.2.10-32bit", "3.2.10-alpine", "3.2.100-nanoserver", "3.2.100-windowsservercore",
		"3.2.11", "3.2.11-32bit", "3.2.11-alpine", "3.2.2", "3.2.2-32bit", "3.2.2-alpine",
		"3.2.3", "3.2.3-32bit", "3.2.3-alpine", "3.2.4", "3.2.4-32bit", "3.2.4-alpine",
		"3.2.5", "3.2.5-32bit", "3.2.5-alpine", "3.2.6", "3.2.6-32bit", "3.2.6-alpine",
		"3.2.6-nanoserver", "3.2.6-windowsservercore", "3.2.7", "3.2.7-32bit", "3.2.7-alpine",
		"3.2.8", "3.2.8-32bit", "3.2.8-alpine", "3.2.9", "3.2.9-32bit", "3.2.9-alpine", "32bit",
		"4", "4-32bit", "4-alpine", "4.0", "4.0-32bit", "4.0-alpine", "4.0.0", "4.0.0-32bit",
		"4.0.0-alpine", "4.0.1", "4.0.1-32bit", "4.0.1-alpine", "4.0.2", "4.0.2-32bit",
		"4.0.2-alpine", "4.0.4", "4.0.4-32bit", "4.0.4-alpine", "4.0.5", "4.0.5-32bit",
		"4.0.5-alpine", "4.0.6", "4.0.6-32bit", "4.0.6-alpine", "4.0.7", "4.0.7-32bit",
		"4.0.7-alpine", "4.0.8", "4.0.8-32bit", "4.0.8-alpine", "4.0.9", "4.0.9-32bit",
		"4.0.9-alpine", "alpine", "latest", "nanoserver", "windowsservercore",
	}, names)
}
func TestGetTagsForRepositories(t *testing.T) {
	tape, http := mockHTTP("hello-and-figlet")
	defer tape.Stop()

	tags := GetTagsForRepositories([]string{"hello-world", "mgkio/figlet"}, http, "", numWorkers)
	sort.Sort(ByRepoAndName(tags))
	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Repo + ":" + tag.Name
	}
	assertEqual(t, "hello world and figlet", []string{
		"hello-world:latest",
		"hello-world:linux",
		"hello-world:nanoserver",
		"hello-world:nanoserver-1709",
		"hello-world:nanoserver-sac2016",
		"hello-world:nanoserver1709",
		"mgkio/figlet:1",
		"mgkio/figlet:latest",
	}, names)
}

func TestGetDetailsLargeSingleRepo(t *testing.T) {
	tape, http := mockHTTP("large-single")
	defer tape.Stop()

	tags := GetTagDetails([]string{"ubuntu"}, http, "", numWorkers)
	assertEqual(t, "tags", 1, len(tags))
	assertEqual(t, "ubuntu repo", 257, len(tags["ubuntu"]))
	for _, tag := range tags["ubuntu"] {
		assertEqual(t, "tag repo", "ubuntu", tag.Repo)
	}
}
func TestGetDetailsLargeMultipleRepos(t *testing.T) {
	tape, http := mockHTTP("large")
	defer tape.Stop()

	counts := map[string]int{
		"alpine":   11,
		"busybox":  95,
		"debian":   177,
		"docker":   667,
		"golang":   561,
		"nginx":    153,
		"openjdk":  622,
		"postgres": 178,
		"python":   496,
		"redis":    142,
		"ubuntu":   257,
	}
	repos := make([]string, 0, len(counts))
	for repo := range counts {
		repos = append(repos, repo)
	}
	tags := GetTagDetails(repos, http, "", numWorkers)
	for repo, count := range counts {
		assertEqual(t, repo, count, len(tags[repo]))
	}
}

func TestAuthenticateGoodCreds(t *testing.T) {
	tape, httpClient := mockHTTP("good-creds")
	defer tape.Stop()
	tape.SetMatcher(requestAndBodyMatcher)

	badCreds := Credentials{Username: "user", Password: "secret"}
	token, err := Authenticate(httpClient, badCreds)
	assertEqual(t, "error", nil, err)
	assertEqual(t, "token incorrect", "ey...blah.blah...Sq", token)
}

func TestAuthenticateBadCreds(t *testing.T) {
	tape, http := mockHTTP("bad-creds")
	defer tape.Stop()
	tape.SetMatcher(requestAndBodyMatcher)

	badCreds := Credentials{Username: "user", Password: "nope-not-it"}
	token, err := Authenticate(http, badCreds)
	assertEqual(t, "token should be blank", "", token)
	if strings.Index(err.Error(), "login error") == -1 {
		t.Errorf("expected login error, got %#v", err)
	}
}
