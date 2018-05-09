package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/mgk/itags"

	docopt "github.com/docopt/docopt-go"
)

// Version - overridden by -ldflags
var Version = "dev"

func main() {
	usage := `itags - list the image tags for one or more docker repositories

Usage:
  itags [options] REPOSITORY...
  itags ---help

Options:
  -h --help                            Show this screen
  --version                            Show version
  -u <username> --username=<username>  Username
  -p <password> --password=<password>  Password
  --prefix                             Prefix each tag with repository name. This only applies when
                                       a single repository is specified. When multiple repositories
                                       are specified tags are always prefixed with their repository
                                       name.
  --workers=<n>                        Max # of Number of HTTP requests to run in parallel [default: 20]`

	args, _ := docopt.ParseArgs(usage, nil, Version)

	httpClient := &http.Client{Timeout: 15 * time.Second}
	numWorkers, _ := args.Int("--workers")
	username, _ := args.String("--username")
	password, _ := args.String("--password")
	prefix, _ := args.Bool("--prefix")
	repositories := args["REPOSITORY"].([]string)

	var jwt = os.Getenv("DOCKER_TOKEN")
	if username != "" {
		creds := itags.Credentials{Username: username, Password: password}
		var err error
		jwt, err = itags.Authenticate(httpClient, creds)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	var tags []string
	if len(repositories) == 1 {
		tags = itags.GetTags(repositories[0], httpClient, jwt, numWorkers)
		if prefix {
			for i, t := range tags {
				tags[i] = repositories[0] + ":" + t
			}
		}
	} else {
		tags = itags.GetTagsForRepositories(repositories, httpClient, jwt, numWorkers)
	}
	sort.Strings(tags)
	fmt.Println(strings.Join(tags, "\n"))
}
