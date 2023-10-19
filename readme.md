#go-web-crawler

Hello there guys, This is a project for making a web crawler in go with some additional functionality like caching and other things.

Current Things built: 
    1) Crawler: Done
    2) Caching: Done
    3) Retrying: Not Done
    4) Priority Customers: Not Done
    5) Concurrency: Not Done
    6) Admin Access : Not Done

To use this:
Run : "go get" to fetch all the dependencies
Run : "go run main.go" to run the main file

This will create a localhost server which is listening on port 8080.
The only endpoint it serves currently is /crawl which is a get request.

This /crawl expects a url as a basepoint. This is a sample body for your http request to this server.
{
    "Url":"<Your URL>"
}