package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

var (
	input  = flag.String("input", "themes.txt", "address file")
	output = flag.String("output", "theme-rank.txt", "rank file")
	token  = flag.String("token", "", "github person token")
)

func main() {
	flag.Parse()
	if *token == "" {
		panic("empty token")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: *token,
	})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	repos, err := getRepositoriesFromFile(*input)
	if err != nil {
		log.Printf("get repos err: %s", err)
		return
	}

	for _, repo := range repos {
		setRepositoryStarCount(client, repo)
	}

	err = saveToFile(*output, repos)
	if err != nil {
		log.Printf("save to file err: %s\n", err)
	}
}

type repository struct {
	address   string
	owner     string
	repo      string
	starCount int
}

func getRepositoriesFromFile(path string) ([]*repository, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return getRepositories(f)
}

func getRepositories(r io.Reader) ([]*repository, error) {
	var (
		repos   = make([]*repository, 0, 100)
		scanner = bufio.NewScanner(r)
	)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "/")
		if len(parts) < 3 {
			log.Printf("invalid line: %s\n", line)
			continue
		}

		repos = append(repos, &repository{
			address: line,
			owner:   parts[1],
			repo:    parts[2],
		})
	}
	return repos, scanner.Err()
}

func setRepositoryStarCount(client *github.Client, repo *repository) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	rsp, _, err := client.Repositories.Get(ctx, repo.owner, repo.repo)
	if err != nil {
		log.Printf("get repo err: %s", err)
		return
	}

	repo.starCount = rsp.GetStargazersCount()
}

func saveToFile(path string, repos []*repository) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer func() {
		err := f.Sync()
		if err != nil {
			log.Printf("sync file err: %s", err)
		}
		f.Close()
	}()

	return save(f, repos)
}

func save(w io.Writer, repos []*repository) error {
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].starCount > repos[j].starCount
	})

	buf := bufio.NewWriter(w)
	defer func() {
		err := buf.Flush()
		if err != nil {
			log.Printf("flush err: %s", err)
		}
	}()

	for _, repo := range repos {
		line := fmt.Sprintf("%4d %s\n", repo.starCount, repo.address)
		_, err := buf.WriteString(line)
		if err != nil {
			return fmt.Errorf("write %s err: %s", line, err)
		}
	}
	return nil
}
