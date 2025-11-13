package crawler


import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"

	"github.com/kobekimmes/spider-man/models"
	"github.com/kobekimmes/spider-man/q"
)

type Mapping map[string]*models.PageResult


var m sync.Mutex
var UrlMap = Mapping{}

func Crawl(search string, start string, depth int, debug bool) {

	switch search {
	case "dfs":
		CrawlDepthFirst(start, depth, debug)
	case "bfs":
		CrawlBreadthFirst(start, depth, debug)
	case "bfs concurrent":
		CrawlConcurrentBFS(start, depth, debug)
	default:
		fmt.Print("None selected...\n")
	}
}


func CrawlDepthFirst(pageUrl string, depth int, debug bool) {
	if depth <= 0 {
		if debug { fmt.Print("DEPTH REACHED\n") }
		return
	}

	if debug { fmt.Printf("CRAWLING: '%s'\n", pageUrl) }

	if _, ok := UrlMap[pageUrl]; ok {
		if debug { fmt.Printf("DUPLICATE: %s already visited\n", pageUrl) }
		return
	}

	pageResult, err := FetchPage(pageUrl)
	if err != nil {
		if debug { fmt.Printf("ERROR: %v\n", err) }
		return
	}

	UrlMap[pageResult.PageUrl] = pageResult

	for _, url := range pageResult.FoundUrls {
		CrawlDepthFirst(url, depth - 1, debug)
	}
}


func CrawlBreadthFirst(pageUrl string, depth int, debug bool) {

	type job struct {
		pageUrl string
		currentDepth int
	}
	
	queue := q.Q[job]{}
	queue.Enqueue(
		job {
			pageUrl: pageUrl,
			currentDepth: 0,
		},
	)

	for queue.Size() > 0 {

		currentJob, _ := queue.Dequeue()

		if debug { fmt.Printf("CRAWLING: '%s'\n", currentJob.pageUrl) }
	
		if currentJob.currentDepth >= depth {
			if debug { fmt.Print("DEPTH REACHED\n") }
			continue
		}

		if _, ok := UrlMap[currentJob.pageUrl]; ok {
			if debug { fmt.Printf("DUPLICATE: '%s' already visited\n", pageUrl) }
			continue
		}

		pageResult, err := FetchPage(currentJob.pageUrl)
		if err != nil {
			if debug { fmt.Printf("ERROR: %v\n", err) }
			continue
		}

		UrlMap[pageResult.PageUrl] = pageResult

		for _, foundUrl := range pageResult.FoundUrls {
			queue.Enqueue(job {
				pageUrl: foundUrl,
				currentDepth: currentJob.currentDepth + 1,
			})
			if debug { fmt.Printf("ENQUEUING: '%s'\n", foundUrl) }
		}
	}
}


func CrawlConcurrentBFS(pageUrl string, depth int, debug bool) {
	type job struct {
		pageUrl      string
		currentDepth int
	}

	queue := make(chan job, 1000)
	active := make(chan int, 1)
	done := make(chan struct{})

	queue <- job {
		pageUrl: pageUrl, 
		currentDepth: 0,
	}
	workers := 10


	counter := func() {
		count := 1
		for delta := range active {
			count += delta
			if count == 0 {
				close(queue)
				close(done)
				return
			}
		}
	}
	go counter()

	worker := func() {
		for currentJob := range queue {

			if debug { fmt.Printf("CRAWLING: '%s'\n", currentJob.pageUrl) }

			if currentJob.currentDepth >= depth {
				active <- -1
				if debug { fmt.Print("DEPTH REACHED\n") }
				continue
			}

			m.Lock()
			if _, ok := UrlMap[currentJob.pageUrl]; ok {
				m.Unlock()
				active <- -1
				if debug { fmt.Printf("DUPLICATE: '%s' already visited\n", currentJob.pageUrl) }
				continue
			}
			m.Unlock()

			pageResult, err := FetchPage(currentJob.pageUrl)
			if err != nil {
				active <- -1
				if debug { fmt.Printf("ERROR: %v\n", err) }
				continue
			}

			m.Lock()
			UrlMap[pageResult.PageUrl] = pageResult
			m.Unlock()

			for _, foundUrl := range pageResult.FoundUrls {
				select {
				case queue <- job {
					pageUrl:      foundUrl,
					currentDepth: currentJob.currentDepth + 1,
				}:
					active <- 1
					if debug { fmt.Printf("ENQUEUING: '%s'\n", foundUrl) }
				default:
					if debug { fmt.Printf("QUEUE FULL: dropping '%s'\n", foundUrl) }
				}			
			}

			active <- -1
		}
	}

	for range workers {
		go worker()
	}

	<-done
	close(active)
}



func FindUrls(pageBody string) []string {
	re := regexp.MustCompile(`https?://[a-zA-Z0-9./?=_-]+`)
	matches := re.FindAllString(pageBody, -1)
	return matches
}


func FetchPage(pageURL string) (*models.PageResult, error) {
	resp, err := http.Get(pageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	buffer := make([]byte, 10) 
	_, err = io.ReadFull(resp.Body, buffer)
	if err != nil && err != io.EOF {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	pageBody := string(buffer)
	foundUrls := FindUrls(pageBody)
	re := regexp.MustCompile(`(?i)<title>(.*?)</title>`)
	matches := re.FindSubmatch(buffer)

	pageTitle := "Unknown Page Title"
	if len(matches) > 1 {
		pageTitle = string(matches[1])
	}

	return &models.PageResult{
		PageUrl:   pageURL,
		PageTitle: pageTitle,
		PageBody:  pageBody,
		FoundUrls: foundUrls,
	}, nil
}