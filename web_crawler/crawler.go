package main

import (
    "fmt"
    "golang.org/x/net/html"
    "log"
    "net/http"
    "runtime"
    "sync"
)

// to check duplication
var fetched = struct {
	m map[string]error
	sync.Mutex
}{m: make(map[string]error)}


func fetch(url string) (*html.Node, error) {
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return doc, nil
}

func parseFollowing(doc *html.Node) []string {
	var urls = make([]string, 0)

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "avatar left" {
					for _, a := range n.Attr {
						if a.Key == "alt" {
							fmt.Println(a.Val) // User name
							break
						}
					}
				}

				if a.Key == "class" && a.Val == "gravatar" {
					user := n.Parent.Attr[0].Val // Firat attr: href

					urls = append(urls, "https://github.com"+user+"/following")
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c) // recursive call to travel all siblings and childs
		}
	}
	f(doc)

	return urls
}

func crawl(url string) {
	fetched.Lock()
	if _, ok := fetched.m[url]; ok {
		fetched.Unlock()
		return
	}
	fetched.Unlock()

	doc, err := fetch(url)

	fetched.Lock()
	fetched.m[url] = err 
	fetched.Unlock()

	urls := parseFollowing(doc) // Print user info and Get Following urls

	done := make(chan bool)
	for _, u := range urls { // create go rutine as num(urls)
		go func(url string) {
			crawl(url)
			done <- true
		}(u)
	}

	for i := 0; i < len(urls); i++ {
		<-done // wait for all go rutines
	}
}

func main() {
	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	crawl("https://github.com/astinchoi/following")
}