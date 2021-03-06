package scrapper

import (
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type extractedJob struct {
	id       string
	title    string
	location string
	salary   string
}

//Scrape Indeed by a term
func Scrape(term string) {

	var baseUrl string = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"

	start := time.Now()

	totalPages := getPages(baseUrl)

	e := make(chan []extractedJob)

	var results []extractedJob

	for i := 0; i < totalPages; i++ {
		go getPage(baseUrl, i, e)
	}

	for i := 0; i < totalPages; i++ {
		results = append(results, <-e...)
	}

	writeJobs(results)

	end := time.Now()
	fmt.Println("Done, extracted Duration : ", end.Second()-start.Second(), "s")

}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")

	checkErr(err)

	w := csv.NewWriter(file)

	defer w.Flush()

	headers := []string{"ID", "Title", "Location", "Salary"}

	wErr := w.Write(headers)

	checkErr(wErr)

	result := [][]string{}
	for _, job := range jobs {
		jobSlice := []string{job.id, job.title, job.location, job.salary}
		result = append(result, jobSlice)
	}

	w.WriteAll(result)

}

func getPage(baseUrl string, page int, e chan<- []extractedJob) {
	var jobs []extractedJob

	url := baseUrl + "&start=" + strconv.Itoa(page*50)

	fmt.Println("Requesting url : ", url)
	res, err := http.Get(url)

	checkErr(err)
	checkCode(res)

	doc, err := goquery.NewDocumentFromReader(res.Body)

	checkErr(err)

	ex := make(chan extractedJob)

	searchCards := doc.Find(".job_seen_beacon")

	searchCards.Each(func(i int, s *goquery.Selection) {
		go extract(s, ex)
	})

	for i := 0; i < searchCards.Length(); i++ {
		jobs = append(jobs, <-ex)
	}

	e <- jobs
}

func extract(s *goquery.Selection, ex chan<- extractedJob) {
	jt := s.Find(".jobTitle").Find("a")

	id, _ := jt.Attr("data-jk")
	title, _ := jt.Find("span").Attr("title")

	location := s.Find(".companyLocation").Text()

	salary := strings.Replace(s.Find(".attribute_snippet").Text(), "\"", "", 1)

	ex <- extractedJob{id: id, title: title, location: location, salary: salary}
}

func getPages(baseUrl string) int {
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
		log.Fatalln("Error")
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error")
	}
}
