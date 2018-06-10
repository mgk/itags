// itags is a command line utility that lists the tags for
// dockey hub repositories.
package main

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/mgk/itags"

	docopt "github.com/docopt/docopt-go"
)

// Version - overridden by -ldflags for releases
var Version = "dev"

func main() {
	usage := `itags - list the image tags for one or more docker repositories

To list tags for private repostories either specify a username and password -or-
a docker JWT. See https://github.com/mgk/itags#notes-on-private-repos for how to
get a JWT.

Usage:
  itags [options] REPOSITORY...
  itags ---help

Options:
  -h --help                            Show this screen
  --version                            Show version
  -u <username> --username=<username>  Username
  -p <password> --password=<password>  Password
  --jwt <jwt>                          JWT to use, ignored it username and password specified
  --sort-by (name | lastUpdated)       Sort results by repo and name or lastUpdate [default: name]
  --show-last-updated                  Include lastUpdated in output
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
	prefix = prefix || len(repositories) > 1
	showLastUpdated, _ := args.Bool("--show-last-updated")

	sortBy, _ := args.String("--sort-by")
	if sortBy != "name" && sortBy != "lastUpdated" {
		fmt.Println("--sort-by must be 'name' or 'lastUpdated'")
		os.Exit(1)
	}

	jwt, _ := args.String("--jwt")
	if username != "" {
		creds := itags.Credentials{Username: username, Password: password}
		var err error
		jwt, err = itags.Authenticate(httpClient, creds)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	tags := itags.GetTagsForRepositories(repositories, httpClient, jwt, numWorkers)
	if sortBy == "name" {
		sort.Sort(itags.ByRepoAndName(tags))
	} else {
		sort.Sort(itags.ByLastUpdated(tags))
	}
	var nameWidth int
	if showLastUpdated {
		for _, tag := range tags {
			nameWidth = max(nameWidth, len(tag.Repo)+len(tag.Name)+1)
		}
	}
	for _, tag := range tags {
		printTag(tag, prefix, nameWidth, showLastUpdated)
	}
}

func printTag(tag itags.Tag, repoPrefix bool, namewWidth int, showLastUpdated bool) {
	var name string
	if repoPrefix {
		name = tag.Repo + ":" + tag.Name
	} else {
		name = tag.Name
	}
	if showLastUpdated {
		nameFormat := fmt.Sprintf("%%%ds", namewWidth)
		fmt.Printf(nameFormat+" %s\n", tag.Name, tag.LastUpdated)
	} else {
		fmt.Println(name)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}
