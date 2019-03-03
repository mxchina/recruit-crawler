package worker

import (
	"20181022/recruit-crawler/types"
	"20181022/recruit-crawler/utils"
	"bufio"
	"github.com/json-iterator/go"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
)


// 智联招聘通过ajax获取数据，get方法。参数中有个start，start/90表示第几页，每页90条数据
// 智联招聘每一页是90个职位
const (
	infoSource   = "智联"
	itemsPerPage = 90
)

type ZhiLian struct{}

// 参数wg必须要在main中add过，才能在main中wait，然后逐级传递，这里Add后，传递到parse，parse需要先接住这里的add（调用done）。
// 然后parse需要传递到saver，所以parse也需要add，parse的add由saver来done。一对一
func (z ZhiLian) Fetch(wg *sync.WaitGroup) <-chan *[]byte {
	out := make(chan *[]byte)
	// 10 是一个预估值，实际应该只有7,8页左右，如果没到10页，下面判断了ContentLength不正常也会return
	for i := 0; i < itemsPerPage*10; i += itemsPerPage {
		wg.Add(1)
		go func(i int) {
			url := "https://fe-api.zhaopin.com/c/i/sou?start=" + strconv.Itoa(i) + "&pageSize=90&cityId=551&workExperience=-1&education=-1&companyType=-1&employmentType=-1&jobWelfareTag=-1&kw=%E8%BF%90%E7%BB%B4&kt=3&_v=0.07826219&x-zp-page-request-id=133a5213903d4e5fbfc601595283bdd9-1551362993438-539926"
			resp, err := http.Get(url)
			if err != nil {
				log.Println("fetch error：", err)
			}
			if resp.ContentLength < 1000 {
				// 这里如果ContentLength小于1000表示，已经没有数据了，return掉，也不会再发数据往parse了。
				// 但是在开goroutine之前wg.add过，所以这里要接住。
				wg.Done()
				return
			}
			defer resp.Body.Close()

			bodyReader := bufio.NewReader(resp.Body)

			// transform encode any to utf-8
			e := utils.DetermineEncoding(bodyReader)
			utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())
			bytes, err := ioutil.ReadAll(bufio.NewReader(utf8Reader))
			if err != nil {
				log.Println("fetch err：", err)
				wg.Done()
				return
			}
			out <- &bytes

		}(i)
	}
	return out
}

func (z ZhiLian) Parse(wg *sync.WaitGroup, from <-chan *[]byte, to chan<- types.JobInfoDone) {
	go func() {
		for reader := range from {
			// 拿到结果集，这里是个[]的jsoniter.Any类型。后续还要继续Unmarshal
			results := jsoniter.Get(*reader, "data").Get("results")
			for i := 0; i < results.Size(); i ++ {
				// 这是单个的网页原始的json对象字符串
				item := results.Get(i)
				// 解析出item中的有用字段，然后发往saver存储
				//go parseAndSend(wg, item, to)
				jobInfoDone := types.CreateJobInfoDone(wg, item, infoSource)
				// 将该jobInfo 存到数据库，发往负责save的channel
				wg.Add(1)
				to <- *jobInfoDone
			}
			// 这里接住Fetch的wg.add
			wg.Done()
		}
	}()


}

//func parseAndSend(wg *sync.WaitGroup, item jsoniter.Any, toSaver chan types.JobInfoDone) {
//
//	jobInfoDone := types.CreateJobInfoDone(wg, item, infoSource)
//	// 将该jobInfo 存到数据库，发往负责save的channel
//	toSaver <- *jobInfoDone
//
//}
