package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestGetRepositories(t *testing.T) {
	s := `abc/def/123
ffff
456/789/ghi`
	r := strings.NewReader(s)
	repos, err := getRepositories(r)
	if err != nil {
		t.Fatalf("%s\n", err)
	}
	for _, repo := range repos {
		fmt.Printf("%+v\n", repo)
	}
}

func TestFmt(t *testing.T) {
	fmt.Printf("%4d", 1234)
}
