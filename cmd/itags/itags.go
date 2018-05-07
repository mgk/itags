package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mgk/itags"

	docopt "github.com/docopt/docopt-go"
)

// Version Overridden by -ldflags
var Version = "dev"

func main() {
	usage := `itags - list the image tags for one or more docker repositories

  Usage:
    itags [options] REPOSITORY...
    itags -h

  Options:
    -h --help      Show this screen.
    --version      Show version.
    -p --prefix    Prefix each tag with repository name. This only applies when
                   a single repository is specified. When multiple repositories
                   are specified tags are always prefixed with their repository
                   name.
    --workers=<n>  Max # of Number of HTTP requests to run in parallel [default: 20].`

	args, _ := docopt.ParseArgs(usage, nil, Version)

	httpClient := &http.Client{Timeout: 15 * time.Second}
	numWorkers, _ := args.Int("--workers")
	prefix, _ := args.Bool("--prefix")
	repositories := args["REPOSITORY"].([]string)

	var tags []string
	if len(repositories) == 1 {
		tags = itags.GetTags(repositories[0], httpClient, numWorkers)
		if prefix {
			for i, t := range tags {
				tags[i] = repositories[0] + ":" + t
			}
		}
	} else {
		tags = itags.GetTagsForRepositories(repositories, httpClient, numWorkers)
	}
	sort.Strings(tags)
	fmt.Println(strings.Join(tags, "\n"))
}
