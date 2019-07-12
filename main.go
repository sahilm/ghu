package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	const ghEnvVar = "GITHUB_API_TOKEN"
	const patternEnvVar = "REPO_PATTERN"

	token, ok := os.LookupEnv(ghEnvVar)
	if !ok {
		log.Fatal(fmt.Sprintf("%s must be provided", ghEnvVar))
	}
	pattern, ok := os.LookupEnv(patternEnvVar)
	if !ok {
		log.Fatal(fmt.Sprintf("%s must be provided", patternEnvVar))
	}
	patternRegexp := regexp.MustCompile(pattern)

	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)))
	var reposToUnwatch []*github.Repository
	opt := github.ListOptions{PerPage: 100}
	for {
		repos, resp, err := client.Activity.ListWatched(ctx, "", &opt)
		if err != nil {
			log.Fatal(err)
		}
		for _, repo := range repos {
			if ok := patternRegexp.MatchString(repo.GetFullName()); ok {
				reposToUnwatch = append(reposToUnwatch, repo)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	for _, repo := range reposToUnwatch {
		fmt.Println(repo.GetFullName())
	}

	fmt.Print("Y/N?: ")
	reader := bufio.NewReader(os.Stdin)
	for {
		b, _, _ := reader.ReadLine()
		resp := string(b)

		if strings.EqualFold(resp, "Y") {
			fmt.Println("unwatching...")
			for _, repo := range reposToUnwatch {
				ownerAndRepo := strings.Split(repo.GetFullName(), "/")
				_, err := client.Activity.DeleteRepositorySubscription(ctx, ownerAndRepo[0], ownerAndRepo[1])
				if err != nil {
					log.Fatal(err)
				}
			}
			os.Exit(0)
		}
		if strings.EqualFold(resp, "N") {
			os.Exit(0)
		}
	}

}
