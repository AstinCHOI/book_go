package main

import (
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/html"
	"runtime"
	"sync"
)

type FollowingResult struct {
	url string
	name string
}

type StarsResult struct {
	repo string
}

type FetchedUrl struct {
	m map[string]error
	sync.Mutex
}

type FetchedRepo struct {
	m map[string]struct{}
	sync.Mutex
}

type Crawler interface {
	Crawl()
}

type Pipeline struct {
	request chan Crawler
	done 	chan struct{}
	wg 		*sync.WaitGroup
}

type GitHubFollowing struct {
	fetchedUrl 	*FetchedUrl
	p 			*Pipeline
	stars 		*GitHubStars
	result 		chan FollowingResult
	url 		string
}

type GitHubStars struct {
	fetchedUrl 	*FetchedUrl
	fechedRepo  *FechedRepo
	p 			*Pipeline
	result 		chan StarsResult
	url 		string
}

func (g *GitHubFollowing) Request(url string) {
	g.p.request <- &GitHubFollowing {
		fetchedUrl: g.fetchedUrl,
		p:			g.p,
		result:		g.result,
		stars:		g.stars,
		url:		url,
	}
}

func (g *GitHubStars) Request(url string) {
	g.p.request <- &GitHubStars {
		fetchedUrl:	 g.fetchedUrl,
		fetchedRepo: g.fetchedRepo,
		p:			 g.p,
		result:		 g.result,
		url:		 url,
	}
}

func (g *GitHubFollowing) Parse(doc *html.Node) <-chan string {
	name := make(chan string)
	go func() {
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "img" {
				for _, a := rnage n.Attr {
					if a.Key == "class" && a.Val == "avatar left" {
						for _, a := range n.Attr {
							if a.Key == "alt" {
								name <- a.Val
								break
							}
						}
					}

					if a.Key == "class" && a.Val == "gravatar" {
						user := n.Parent.Attr[0].Val

						g.Request("https://github.com" + user + "/following")

						g.stars.Request("https://github.com/stars" + user)
						break
					}
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)
	}()

	return name
}