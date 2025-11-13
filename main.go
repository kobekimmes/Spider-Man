package main

import (
	"fmt"
	//"os"
	"time"

	"github.com/kobekimmes/spider-man/crawler"
	//"github.com/kobekimmes/spider-man/db"
)


func main() {

	//db.InitConnection()

	depth := 3
	search := "dfs"

	start := time.Now()
	crawler.Crawl(search, "https://kobekimmes.github.io", depth, false)
	elapsed := time.Since(start)x

	for url, result := range crawler.UrlMap {
		fmt.Printf("%s -> %s\n", url, result.PageTitle)
	}
	fmt.Printf("Results in %v at search depth %d performing %s\n", elapsed, depth, search)

}