# Scraper for Greyhound Bet
This scraper is to scrap the website - [Greyhound Bet].
We have 2 scraping engines for now. 
  - Web driver [Chromedp]-based scraping engine
  - API based scraping engine

### Tech

Built in [Go]


### Installation

This scraper requires [Go] to run.  Please follow the [Go Installation guide].

For [Chromedp] based scrapping engine :
```sh
$ go get -u github.com/knq/chromedp
```
For API based scrapping engine : 
```sh
$ go get github.com/Jeffail/gabs
```

### How to run

Run in the usual way for chromedp engine:
```sh
$ go run scraper_chromedp.go
```
Run in the usual way for API based engine:
```sh
$ go run scraper_api.go
```

### Todos

 - Write more scraping functions via API
 - API integration for POST

License
----

MIT

   [Go]: <https://golang.org/>
   [Go Installation guide]: <https://golang.org/doc/install>
   [Greyhound Bet]: <http://greyhoundbet.racingpost.com/>
   [Chromedp]: <https://github.com/knq/chromedp>

