package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
)

type Line struct {
	UA   string
	User string
}

func main() {
	versions := map[string]string{}
	reqs := map[string]int{}

	dec := json.NewDecoder(os.Stdin)
	for {
		var l Line
		err := dec.Decode(&l)
		if err != nil {
			break
		}
		if l.User != "" && l.UA != "" {
			versions[l.User] = l.UA
			reqs[l.User]++
		}
	}

	var usernames []string
	for user := range versions {
		usernames = append(usernames, user)
	}
	sort.Strings(usernames)

	tw := tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
	for _, user := range usernames {
		fmt.Fprintf(tw, "%s\t%s\t%d\n", user, versions[user], reqs[user])
	}
	tw.Flush()
}
