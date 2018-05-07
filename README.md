# itags
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

Download a [binary](https://github.com/mgk/itags/releases) or:

```bash
go get github.com/mgk/itags/cmd/itags
```

## Usage

```bash
# Get tags for a repo
itags redis

# Get tags for multiple repos
itags hello-world mgkio/figlet hello-world
```

Go pacakge API: see tests for examples.

## Notes

Private repositories are supported, but currently require that you set the
environment variable `DOCKER_TOKEN` to a valid JWT. This is a stop-gap measure.
`itags` will add support for acquiring the JWT via the
`https://hub.docker.com/v2/users/login/` endpoint.

For the ambitious you can get a JWT with `curl` and `jq`:

```bash
export DOCKER_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d '{"username": "MY_USERNAME", "password": "MY_PASSWORD"}' https://hub.docker.com/v2/users/login/ | jq -r .token)
```
replacing `MY_USERNAME` and `MY_PASSWORD` with your docker username and password.

REST purists may correctly point out that `itags` constructs its own URLs and
and is thus not a [HATEOAS](https://en.wikipedia.org/wiki/HATEOAS) client. This
is largely because I wanted to run mutiple HTTP requests in parallel to retrieve
a large number of tags quickly. A HATEOAS mode would be interesting to add: it
would strictly follow the `next` links in the tag listings instead of generating
its own page URLs. It could still run requests in parallel for multiple
repositories, but for a given repository the requests would be sequential.

## Feedback

Comments, corrections, and questions are welcome. Please open [an
issue](https://github.com/mgk/itags/issues) with any feedback.


## References
Many thanks to Jerry Baker for the [docker KB article showing how it's done](https://success.docker.com/article/how-do-i-authenticate-with-the-v2-api).

*Note: the KB article uses a large page size (10,000) but as far as I can tell
the max page size returned by docker is 100.*

### Todos
- prompt for login and generate token
- CI build
- go doc for package
- negative test cases
- support other registries such as gcr.io
- HATEOAS mode
