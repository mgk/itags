# itags
[![Build Status](https://img.shields.io/travis/mgk/itags.svg)](https://travis-ci.org/mgk/itags)
[![Go Report Card](https://goreportcard.com/badge/github.com/mgk/itags)](https://goreportcard.com/report/github.com/mgk/itags)
[![Coverage Status](https://coveralls.io/repos/github/mgk/itags/badge.svg)](https://coveralls.io/github/mgk/itags)
[![GoDoc](https://godoc.org/github.com/mgk/itags/cmd/itags?status.svg)](https://godoc.org/github.com/mgk/itags/cmd/itags)
![Flux Cap](https://img.shields.io/badge/flux%20capacitor-1.21%20GW-orange.svg)

Quickly get tags for docker repositories using the docker hub REST API.

It is surprisingly difficult to list the tags for a given docker image
repository. There is no built in `docker` command for this nor does it appear to
be part of the docker SDKs. There is a docker REST API to get tags. It requires
issuing multiple HTTP requests, one per page of 100 tags, and some JSON parsing.
This is quite doable with `jq` and some scripting, but it is hardly convenient
and it can be slow to get the tags for multiple repositories. That's why I wrote
`itags`.

`itags` is a command line utility and go package for listing the tags for one or
more docker images. itags is concurrent, running up to 20 HTTP requests in
parallel. This means that:

```bash
itags alpine busybox debian docker golang \
      nginx openjdk postgres python redis ubuntu
```

takes ~1.5 seconds to retrieve 3358 tags.

## Install

Download a [released binary](https://github.com/mgk/itags/releases).

Or to use the latest development version:

```bash
go get github.com/mgk/itags/cmd/itags
```

## Usage

```bash
# Get tags for a repo
itags redis

# Get tags for multiple repos
itags hello-world mgkio/figlet hello-world

# Get tags for my-private-repo
itags -u my-username -p secret my-username/my-private-repo

# Get tags for private repo using DOCKER_TOKEN
export DOCKER_TOKEN=my-token
itags my-username/my-private-repo
```

Go pacakge API: see tests for examples.

## Notes on private repos
If username and password are supplied they are used to get a docker
[JWT](https://jwt.io/) which is sent with each request to the docker registry.
Otherwise if the environment variable `DOCKER_TOKEN` is set it is used instead.

You can get a JWT with `curl` and [jq](https://stedolan.github.io/jq/):

```bash
export DOCKER_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "MY_USERNAME", "password": "MY_PASSWORD"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```
replacing `MY_USERNAME` and `MY_PASSWORD` with your docker username and password.

## Feedback

Comments, corrections, and questions are welcome. Please open [an
issue](https://github.com/mgk/itags/issues) with any feedback.


## References
Many thanks to Jerry Baker for the [docker KB article showing how it's done](https://success.docker.com/article/how-do-i-authenticate-with-the-v2-api).

*Note: the KB article uses a large page size (10,000) but as far as I can tell
the max page size returned by docker is 100.*
