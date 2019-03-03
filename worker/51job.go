package worker

import (
	"20181022/crawler/types"
	"20181022/crawler/utils"
	"bufio"
	bytes2 "bytes"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// 51job的数据直接在html里，每页50个职位。
// 我搜索的是最近3天的关键字为“运维”的职位，实际差不多10页
// 这里的20是随便取得，只要比实际大就行
const pages = 20
var infoSource51 = "51job"

type Job51 struct{}

func NewUrl(page int) string {
	var url = "https://search.51job.com/list/060000,000000,0000,00,1,99,%25E8%25BF%2590%25E7%25BB%25B4,2," + strconv.Itoa(page) + ".html?lang=c&stype=1&postchannel=0000&workyear=99&cotype=99&degreefrom=99&jobterm=99&companysize=99&lonlat=0%2C0&radius=-1&ord_field=0&confirmdate=9&fromType=5&dibiaoid=0&address=&line=&specialarea=00&from=&welfare="
	return url
}

func (Job51) Fetch(wg *sync.WaitGroup) <-chan *[]byte {
	from := make(chan *[]byte)
	// 每页一个goroutine抓取
	for i := 1; i <= pages; i++ {
		wg.Add(1)
		go func(i int) {
			// Request the HTML page.
			res, err := http.Get(NewUrl(i))
			if err != nil {
				log.Fatal(err)
			}
			defer res.Body.Close()

			if res.StatusCode != 200 {
				log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
			}
			bodyReader := bufio.NewReader(res.Body)

			// transform encode any to utf-8
			e := utils.DetermineEncoding(bodyReader)
			utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

			bytes, err := ioutil.ReadAll(utf8Reader)
			if err != nil {
				log.Println("fetch err：", err)
				wg.Done()
				return
			}

			from <- &bytes
		}(i)
	}
	return from
}

func (Job51) Parse(wg *sync.WaitGroup, from <-chan *[]byte, to chan<- types.JobInfoDone) {
	go func() {
		for bytes := range from {
			// Load the HTML document
			doc, err := goquery.NewDocumentFromReader(bytes2.NewReader(*bytes))
			if err != nil {
				log.Fatal(err)
			}

			items := doc.Find("#resultList.dw_table .el:not(.title)")

			items.Each(func(i int, s *goquery.Selection) {

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
				salaryMin, salaryMax := getSalaryInfo(salary)

				updateDate := s.Find(".t5").Text()

				job := types.JobInfoDone{
					Wg: wg,
					JobInfo: types.JobInfo{
						InfoSource:  infoSource51,
						CreateDate:  "2019-1-1",
						UpdateDate:  "2019-" + updateDate,
						City:        city,
						PositionURL: positionURL,
						Salary:      salary,
						SalaryMin:   salaryMin,
						SalaryMax:   salaryMax,
						JobName:     jobName,
						CompanyName: companyName,
					},
				}

				//send 前wg.add(1)
				wg.Add(1)
				to <- job

			})
			// 接住fetch发过来的add
			wg.Done()
		}
	}()
}

func getSalaryInfo(s string) (min float64, max float64) {
	// 51job上的薪资分为以下几种情况
	//  6-8千/月      0.8-1.2万/月     10-15万/年
	if s == "" {
		return 0, 0
	}
	split := strings.Split(strings.TrimSpace(s), "/")
	r1 := []rune(split[0]) //10-15万
	r2 := []rune(split[1]) //年

	if string(r2) == "年" { //10-15万/年
		if string(r1[len(r1)-1:]) == "万" {
			s1 := string(r1[:len(r1)-1]) // 10-15
			ss := strings.Split(s1, "-") // [10, 15]
			ss0, e := strconv.ParseFloat(ss[0], 64)
			if e != nil {
				ss0 = 0
			}
			ss1, e := strconv.ParseFloat(ss[0], 64)
			if e != nil {
				ss1 = 0
			}
			return ss0 / 1, ss1 / 12
		} else {
			return 0, 0
		}
	}

	if string(r2) == "月" { //  6-8千/月  0.8-1.2万/月
		if string(r1[len(r1)-1:]) == "千" { //  6-8千/月
			s1 := string(r1[:len(r1)-1])            // 6-8
			ss := strings.Split(s1, "-")            // [6, 8]
			ss0, e := strconv.ParseFloat(ss[0], 64) // float 6
			if e != nil {
				ss0 = 0
			}
			ss1, e := strconv.ParseFloat(ss[1], 64) // float 8
			if e != nil {
				ss1 = 0
			}
			return float64(ss0), float64(ss1)

		} else if string(r1[len(r1)-1:]) == "万" { // 0.8-1.2万/月
			s1 := string(r1[:len(r1)-1])            // 0.8-1.2
			ss := strings.Split(s1, "-")            // [0.8, 1.2]
			ss0, e := strconv.ParseFloat(ss[0], 64) // float 0.8
			if e != nil {
				ss0 = 0
			}
			ss1, e := strconv.ParseFloat(ss[1], 64) // float 1.2
			if e != nil {
				ss1 = 0
			}
			return ss0 * 10, ss1 * 10

		} else {
			return 0, 0
		}
	}
	return 0, 0
}