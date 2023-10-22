package main

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly/v2"
	"github.com/patrickmn/go-cache"
)

type IntHeap []time.Time

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i].Before(h[j]) }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x any) {
	*h = append(*h, x.(time.Time))
}

func (h *IntHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func (h *IntHeap) Top() any {
	old := *h
	return old[old.Len()-1]
}

type URL struct {
	urls []string
}

type UserRequest struct {
	Url string `json:"Url"`
}

var c *cache.Cache

var paid IntHeap
var unPaid IntHeap

var isUnpaidUsing bool

func main() {
	paid := &IntHeap{}
	unPaid := &IntHeap{}
	isUnpaidUsing = false
	heap.Init(paid)
	heap.Init(unPaid)

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
	status := ctx.Query("paid")
	var isPaid bool
	if status == "true" {
		isPaid = true
	} else {
		isPaid = false
	}
	now := time.Now()
	if isPaid {
		heap.Push(&paid, now)
	} else {
		heap.Push(&unPaid, now)
	}
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
	if isPaid {
		for (paid.Len() > 0 && paid.Top() != now) || isUnpaidUsing {
			continue
		}
		crawl(url, 2, &u)
		c.Set(url, &u, cache.DefaultExpiration)
		ctx.JSON(200, gin.H{
			"Crawled": u.urls,
		})
		paid.Pop()
	} else {
		for (unPaid.Len() > 0 && unPaid.Top() != now) || isUnpaidUsing || paid.Len() != 0 {
			continue
		}
		isUnpaidUsing = true
		crawl(url, 2, &u)
		c.Set(url, &u, cache.DefaultExpiration)
		ctx.JSON(200, gin.H{
			"Crawled": u.urls,
		})
		unPaid.Pop()
		isUnpaidUsing = false
	}
}

func crawl(url string, depth int, u *URL) {
	if depth <= 0 {
		return
	}
	flag := true
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
		if flag {
			fmt.Println("Trying to visit url: ", url, " again one last time.")
			flag = false
			c.Visit(url)
		} else {
			fmt.Println("Maybe this url is permanently down. URL: ", url)
		}
	})
	c.Visit(url)
}
