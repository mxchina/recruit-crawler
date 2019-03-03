package main

import (
	"bufio"
	bytes2 "bytes"
	"fmt"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/PuerkitoBio/goquery"
)

func NewUrl(i int) string {
	var url =  "https://search.51job.com/list/060000,000000,0000,00,1,99,%25E8%25BF%2590%25E7%25BB%25B4,2,"+strconv.Itoa(i)+".html?lang=c&stype=1&postchannel=0000&workyear=99&cotype=99&degreefrom=99&jobterm=99&companysize=99&lonlat=0%2C0&radius=-1&ord_field=0&confirmdate=9&fromType=5&dibiaoid=0&address=&line=&specialarea=00&from=&welfare="
	return url
}

func ExampleScrape() {
	// Request the HTML page.
	res, err := http.Get(NewUrl(1))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	bodyReader := bufio.NewReader(res.Body)

	// transform encode any to utf-8
	e := determineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())
	bytes, err := ioutil.ReadAll(utf8Reader)

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(bytes2.NewReader(bytes))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(doc.Html())
	fmt.Println("---------------------------")
	// Find the review items
	doc.Find("#resultList.dw_table .el:not(.title)").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		jobName, exists := s.Find(".t1 > span a").Attr("title")
		if !exists {
			jobName = ""
		}
		positionURL, exists := s.Find(".t1 > span a").Attr("href")
		if !exists {
			positionURL = ""
		}

		companyName, exists := s.Find(".t2 > a").Attr("title")
		if !exists {
			companyName = ""
		}

		city := s.Find(".t3").Text()

		salary := s.Find(".t4").Text()
		//salaryMin, salaryMax := getSalaryInfo(salary)

		updateDate := s.Find(".t5").Text()

		fmt.Printf("Review %d: %s - %s - %s - %s - %s - %s\n", i, jobName, positionURL,
			companyName, city, salary, updateDate)
	})
}

func main() {
	ExampleScrape()
}

func determineEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)
	if err != nil {
		log.Printf("Fetcher error: %v", err)
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}