package main

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
}

var baseUrl string = "https://kr.indeed.com/jobs?q=python&limit=50"

func main() {
	totalPages := getPages()

	e := make(chan []extractedJob)

	var results []extractedJob

	for i := 0; i < totalPages; i++ {
		go getPage(i, e)
	}

	for i := 0; i < totalPages; i++ {
		results = append(results, <-e...)
	}

	writeJobs(results)

	fmt.Println("Done, extracted")

}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")

	checkErr(err)

	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"ID", "Title", "Location", "Salary"}

	wErr := w.Write(headers)

	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.location, job.salary}

		w.Write(jobSlice)
	}

}

func getPage(page int, e chan<- []extractedJob) {
	var jobs []extractedJob

	url := baseUrl + "&start=" + strconv.Itoa(page*50)

	fmt.Println("Requesting url : ", url)
	res, err := http.Get(url)

	checkErr(err)
	checkCode(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)

	doc.Find(".job_seen_beacon").Each(func(i int, s *goquery.Selection) {
		jt := s.Find(".jobTitle").Find("a")

		id, _ := jt.Attr("data-jk")
		title, _ := jt.Find("span").Attr("title")

		location := s.Find(".companyLocation").Text()

		salary := strings.Replace(s.Find(".attribute_snippet").Text(), "\"", "", 1)
		jobs = append(jobs, extractedJob{id: id, title: title, location: location, salary: salary})
	})

	e <- jobs

}

func getPages() int {
	var page int

	res, err := http.Get(baseUrl)

	checkErr(err)
	checkCode(res)

	// because of memory leak
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)

	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		page = s.Find("a").Length()
	})

	return page
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Error to Get ", baseUrl)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error to Get ", baseUrl)
	}
}
