package scraper

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	html2md "github.com/JohannesKaufmann/html-to-markdown"
)

func CLI(args []string) int {
	var app appEnv
	err := app.fromArgs(args)
	if err != nil {
		return 2
	}

	if err = app.run(); err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		return 1
	}
	return 0
}

type appEnv struct {
	hc      http.Client
	pubName string
	// substack API cookie
	cookie string
	// html(default) or md
	outputType string
	// folder to write output to. defaults to current directory
	destFolder string
	// fetch posts since this date (YYYY-MM-DD)
	since time.Time
}

func (app *appEnv) fromArgs(args []string) error {
	// Shallow copy of default client
	app.hc = *http.DefaultClient
	fl := flag.NewFlagSet("substackscraper", flag.ContinueOnError)

	fl.StringVar(&app.pubName, "pub", "", "Name of the Substack publication to scrape (required)")
	fl.StringVar(&app.cookie, "cookie", "", "Substack API cookie (remove the `substack.sid` prefix)")
	fl.StringVar(&app.outputType, "output", "html", "Output format: html(default) or md")
	fl.StringVar(&app.destFolder, "dest", ".", "Destination folder to write output to. Defaults to current directory")
	var sinceStr string
	fl.StringVar(&sinceStr, "since", "1970-01-01", "Fetch posts since this date (YYYY-MM-DD). Defaults to 1970-01-01")
	if err := fl.Parse(args); err != nil {
		return err
	}

	if app.pubName == "" {
		fmt.Fprintln(os.Stderr, "missing required flag: -pub")
		fl.Usage()
		return flag.ErrHelp
	}

	if app.outputType != "html" && app.outputType != "md" {
		fmt.Fprintf(os.Stderr, "invalid output type: %s", app.outputType)
		fl.Usage()
		return flag.ErrHelp
	}

	since, err := time.Parse("2006-01-02", sinceStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing since date: %v\n", err)
		fl.Usage()
		return flag.ErrHelp
	}
	app.since = since

	return nil
}

func (app *appEnv) run() error {
	converter := html2md.NewConverter("", true, nil)
	archive, err := app.fetchArchive()
	if err != nil {
		return err
	}
	fmt.Printf("fetching %d posts\n", len(archive))

	for _, p := range archive {
		slug := p.Slug
		post, err := app.fetchPost(slug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error fetching post %s: %v\n", slug, err)
			continue
		}

		if app.outputType == "md" {
			content, err := converter.ConvertString(post.Body)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error converting post %s: %v\n", slug, err)
				continue
			}
			post.Body = content
		}

		err = app.writePost(post)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing post %s: %v\n", slug, err)
			continue
		}

		// wait on rate limiter
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (app *appEnv) fetchJSON(url string, data interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cookie", "substack.sid="+app.cookie)
	resp, err := app.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(data)
}

// fetch all archived posts after app.since
func (app *appEnv) fetchArchive() (archiveApiResponse, error) {
	offset := 0
	var results archiveApiResponse
	for {
		var ar archiveApiResponse
		err := app.fetchJSON(buildUrl(app.pubName, fmt.Sprintf("archive?offset=%d&limit=%d", offset, 50)), &ar)
		if err != nil {
			return nil, err
		}
		for _, a := range ar {
			if a.PostDate.After(app.since) {
				results = append(results, a)
			}
		}
		if len(ar) < 50 || ar[49].PostDate.Before(app.since) {
			// we're done, there's nothing left
			break
		}
		offset += 50

		// wait on rate limiter
		time.Sleep(1 * time.Second)
	}
	return results, nil
}

func (app *appEnv) fetchPost(slug string) (postApiResponse, error) {
	var pr postApiResponse
	err := app.fetchJSON(buildUrl(app.pubName, fmt.Sprintf("posts/%s", slug)), &pr)
	if err != nil {
		return postApiResponse{}, err
	}
	return pr, nil
}

func (app *appEnv) writePost(post postApiResponse) error {
	f, err := os.Create(fmt.Sprintf("%s/%s.%s", app.destFolder, post.Slug, app.outputType))

	if err != nil {
		return err
	}
	defer f.Close()

	if app.outputType == "md" {
		_, err = f.WriteString(fmt.Sprintf(
			`---
title: %s
date: %s
alias: []
tags: [%s]
---

# %s

> %s

%s`, post.Title, post.PostDate.Format("2006-01-02"), post.SectionSlug, post.Title, post.Subtitle, post.Body))
	} else {
		_, err = f.WriteString(fmt.Sprintf("<h1>%s</h1><h2>%s</h2>%s", post.Title, post.Subtitle, post.Body))
	}
	return err
}
