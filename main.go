package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly/v2"
	"github.com/patrickmn/go-cache"
)

type MaxHeap struct {
	arr []time.Time
}

func (h *MaxHeap) Insert(key time.Time) {
	h.arr = append(h.arr, key)
	h.MaxHeapify(len(h.arr) - 1)
}

func (h *MaxHeap) Length() int {
	return len(h.arr)
}

func (h *MaxHeap) MaxHeapify(ind int) {
	for h.arr[parent(ind)].Before(h.arr[ind]) {
		h.arr[parent(ind)], h.arr[ind] = h.arr[ind], h.arr[parent(ind)]
		ind = parent(ind)
	}
}

func (h *MaxHeap) Top() time.Time {
	return h.arr[0]
}

func (h *MaxHeap) Pop() time.Time {
	val := h.Top()
	h.arr[0] = h.arr[len(h.arr)-1]
	h.arr = h.arr[:len(h.arr)-1]
	h.HeapifyDown()
	return val
}

func (h *MaxHeap) HeapifyDown() {
	l := leftChild(0)
	r := rightChild(0)
	comp := 0
	currInd := 0
	for l <= len(h.arr)-1 {
		if l == len(h.arr)-1 {
			comp = l
		} else if h.arr[l].After(h.arr[r]) {
			comp = l
		} else {
			comp = r
		}
		if h.arr[comp].After(h.arr[currInd]) {
			h.arr[comp], h.arr[currInd] = h.arr[currInd], h.arr[comp]
			currInd = comp
			l = leftChild(currInd)
			r = rightChild(currInd)
		} else {
			break
		}
	}
}

func leftChild(ind int) int {
	val := 2*ind + 1
	return val
}

func rightChild(ind int) int {
	val := 2*ind + 2
	return val
}

func parent(ind int) int {
	val := (ind - 1) / 2
	return val
}

type URL struct {
	urls []string
}

type UserRequest struct {
	Url string `json:"Url"`
}

var c *cache.Cache

var paid MaxHeap
var unPaid MaxHeap

var isUnpaidUsing bool

func main() {
	paid := &MaxHeap{}
	unPaid := &MaxHeap{}
	paid.Length()
	unPaid.Length()
	isUnpaidUsing = false
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
		paid.Insert(now)
	} else {
		unPaid.Insert(now)
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
		for (paid.Length() > 0 && paid.Top() != now) || isUnpaidUsing {
			continue
		}
		crawl(url, 2, &u)
		c.Set(url, &u, cache.DefaultExpiration)
		ctx.JSON(200, gin.H{
			"Crawled": u.urls,
		})
		paid.Pop()
	} else {
		for (unPaid.Length() > 0 && unPaid.Top() != now) || isUnpaidUsing || paid.Length() != 0 {
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
