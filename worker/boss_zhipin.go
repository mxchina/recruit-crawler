package worker

import (
	"20181022/recruit-crawler/types"
	"20181022/recruit-crawler/utils"
	"bufio"
	bytes2 "bytes"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const pagesBoss = 5
const infoSourceBoss = "boss直聘"

func getUrl(i int) string {
	return "https://www.zhipin.com/c101040100/?query=%E8%BF%90%E7%BB%B4&period=5&page=" + strconv.Itoa(i) + "&ka=page-" + strconv.Itoa(i)
}

type BossZhiPin struct{}

func (BossZhiPin) Fetch(wg *sync.WaitGroup) <-chan *[]byte {
	from := make(chan *[]byte)
	for i := 0; i < pagesBoss; i++ {
		wg.Add(1)
		go func(i int) {
			resp, err := http.Get(getUrl(i))
			if err != nil {
				log.Printf("fetch err：%s", err.Error())
				return
			}
			if resp.StatusCode != 200 {
				log.Printf("status code error: %s", resp.Status)
				return
			}
			defer resp.Body.Close()

			bodyReader := bufio.NewReader(resp.Body)
			e := utils.DetermineEncoding(bodyReader)
			utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())
			bytes, err := ioutil.ReadAll(utf8Reader)
			if err != nil {
				log.Println("fetch err：", err)
				wg.Done()
				return
			}

			from <- &bytes
		}(i + 1)
	}
	return from
}

func (BossZhiPin) Parse(wg *sync.WaitGroup, from <-chan *[]byte, to chan<- types.JobInfoDone) {
	go func() {
		for bytes := range from {
			// Load the HTML document
			doc, err := goquery.NewDocumentFromReader(bytes2.NewReader(*bytes))
			if err != nil {
				log.Println("NewDocumentFromReader error：%s", err.Error())
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

				job := types.JobInfoDone{
					Wg: wg,
					JobInfo: types.JobInfo{
						InfoSource:  infoSourceBoss,
						CreateDate:  createDate,
						UpdateDate:  "2019-1-1",
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

func filterCreateDate(s string) string {
	if s == "发布于昨天" {
		now := time.Now()
		d, _ := time.ParseDuration("-24h")
		d1 := now.Add(d)
		return d1.Format("2006-01-02 15:04:05")
	}
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
