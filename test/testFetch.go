package main

import (
	"20181022/recruit-crawler/utils"
	"bufio"
	bytes2 "bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	resp, err := http.Get(getUrl(1))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyReader := bufio.NewReader(resp.Body)
	e := utils.DetermineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())
	bytes, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%s\n", string(bytes))

	doc, err := goquery.NewDocumentFromReader(bytes2.NewReader(bytes))
	if err != nil {
		log.Fatal(err)
	}

	items := doc.Find(".job-list ul li")

	items.Each(func(i int, s *goquery.Selection) {
		jobName := s.Find(".info-primary .job-title").Text()

		positionURL, exists := s.Find(".info-primary a").Attr("href")
		if !exists {
			positionURL = ""
		}
		positionURL = filterUrl(positionURL)

		companyName := s.Find(".info-company a").Text()

		city, err := s.Find(".info-primary p").Html()
		city = filterCity(city, err)

		salary := s.Find(".info-primary .red").Text()
		salaryMin, salaryMax := utils.GetSalaryInfo(salary)

		createDate := s.Find(".info-publis p").Text()
		createDate = filterCreateDate(createDate)

		fmt.Println("----------------------------------------")
		fmt.Printf("count：%d\nInfoSource：%s\nCreateDate：%s\nUpdateDate：%s\nCity：%s\nPositionURL：%s\nSalary：%s\nSalaryMin：%f\nSalaryMax：%f\nJobName：%s\nCompanyName：%s\n",
			i,
			"boss直聘",
			createDate,
			"2019-1-1",
			city,
			positionURL,
			salary,
			salaryMin,
			salaryMax,
			jobName,
			companyName)
		fmt.Println("----------------------------------------")
		//job := types.JobInfoDone{
		//	Wg: wg,
		//	JobInfo: types.JobInfo{
		//		InfoSource:  "boss直聘",
		//		CreateDate:  createDate,
		//		UpdateDate:  "2019-1-1",
		//		City:        city,
		//		PositionURL: positionURL,
		//		Salary:      salary,
		//		SalaryMin:   salaryMin,
		//		SalaryMax:   salaryMax,
		//		JobName:     jobName,
		//		CompanyName: companyName,
		//	},
		//}
	})
}

func filterCity(s string, err error) string {
	// <p>重庆 渝北区 光电园<em class="vline"></em>1-3年<em class="vline"></em>本科</p>
	if err != nil {
		return "未找到"
	}
	RE := regexp.MustCompile(`([\s\S]*?)<em class[\s\S]*`)
	match := RE.FindStringSubmatch(s)
	if len(match) > 1 {
		return strings.Replace(strings.TrimSpace(match[1]), " ", "-", -1)
	}
	return "未找到"
}

func getUrl(i int) string {
	return "https://www.zhipin.com/c101040100/?query=%E8%BF%90%E7%BB%B4&page=1&ka=page-" + strconv.Itoa(i)
}

func filterCreateDate(s string) string {
	// 发布于03月02日
	s = strings.Replace(s, "发布于", "", -1) // 03月02日
	s = strings.Replace(s, "月", "-", -1)
	s = strings.Replace(s, "日", "", -1)
	s = "2019-" + s // 2019-3-02
	return s
}

func filterUrl(s string) string {
	return "https://www.zhipin.com" + s
}
