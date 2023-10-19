package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly/v2"
	"github.com/patrickmn/go-cache"
)

type URL struct {
	urls []string
}

type UserRequest struct {
	Url string `json:"Url"`
}

var c *cache.Cache

func main() {
	c = cache.New(60*time.Minute, 60*time.Minute)
	r := gin.Default()
	r.GET("/crawl", getCrawlResult)
	r.Run()
}

func getCrawlResult(ctx *gin.Context) {
	var req UserRequest
	if err := ctx.BindJSON(&req); err != nil {
		return
	}
	url := req.Url
	obj, found := c.Get(url)
	if found {
		urll := obj.(*URL)
		ctx.JSON(200, gin.H{
			"Crawled": urll.urls,
		})
		c.Set(url, obj, cache.DefaultExpiration)
		return
	}
	u := URL{}
	crawl(url, 2, &u)
	c.Set(url, &u, cache.DefaultExpiration)
	ctx.JSON(200, gin.H{
		"Crawled": u.urls,
	})
}

func crawl(url string, depth int, u *URL) {
	if depth <= 0 {
		return
	}
	c := colly.NewCollector()
	c.CheckHead = true
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		newUrl := e.Attr("href")
		nextUrl := e.Request.AbsoluteURL(newUrl)
		crawl(nextUrl, depth-1, u)
	})
	c.OnResponse(func(r *colly.Response) {
		u.urls = append(u.urls, url)
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Error in visiting site ", url, " with error: ", err)
	})
	c.Visit(url)
}
