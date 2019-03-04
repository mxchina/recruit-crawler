package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var Chart = ChartAlarm{
	ID:     "wwf51b2091e9e0d51d",
	Secret: "LxTHd5MOZU4H6bYijzTqnv_a1qBB9Luq7p3u7f4cT5I",
	AppID:  "1000002",
}

func main() {
	chart := ChartAlarm{
		ID:     "wwf51b2091e9e0d51d",
		Secret: "LxTHd5MOZU4H6bYijzTqnv_a1qBB9Luq7p3u7f4cT5I",
		AppID:  "1000002",
	}

	// .CompanyName, rest.JobName, rest.Salary, rest.SalaryMin, rest.SalaryMax, rest.PositionURL createDate updateDate
	content := fmt.Sprintf("City：%s\nCompanyName：%s\nJobName：%s\nSalary：%s\nPositionURL：%s\ncreateDate：%s\nupdateDate：%s\n",
		"重庆-北碚区",
		"深圳市德瑞信息技术有限公司",
		"IDC运维/机房运维工程师",
		"5-9千/月",
		"https://jobs.51job.com/chongqing-bbq/110726606.html?s=01&t=0",
		"2019-0-0",
		"2019-03-03")

	err := chart.SendText(content)
	if err != nil {
		log.Println("send position failed：", err)
		log.Println("-------------------------------")
		log.Printf("content\n：%s\n", content)
		log.Println("-------------------------------")
	} else {
		log.Println("send position success！")
		log.Println("-------------------------------")
		log.Printf("content\n：%s\n", content)
		log.Println("-------------------------------")
	}
}

// 全局变量保存token
var token string

// ID: "wwf51b2091e9e0d51d",
// Secret: "LxTHd5MOZU4H6bYijzTqnv_a1qBB9Luq7p3u7f4cT5I",
// AppID:  "1000002",
type ChartAlarm struct {
	ID     string
	AppID  string
	Secret string
}

func (c ChartAlarm) getSendUrl() string {
	if token == "" {
		token = c.getNewToken()
	}
	log.Println("token：", token)
	return fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
}

func (c ChartAlarm) getTokenUrl() string {
	return fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", c.ID, c.Secret)
}

func (c ChartAlarm) getNewToken() string {
	resp, err := http.Get(c.getTokenUrl())
	if err != nil {
		log.Println("getNewToken error：", err)
		return ""
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(bufio.NewReader(resp.Body))
	if err != nil {
		log.Println("getNewToken error：", err)
		return ""
	}
	errcode := jsoniter.Get(bytes, "errcode").ToInt()
	if errcode != 0 {
		log.Println("getNewToken error：", fmt.Sprintf("getToken errcode：%d\n", errcode))
		return ""
	}
	return jsoniter.Get(bytes, "access_token").ToString()
}

type myLoad struct {
	ToUser  string `json:"touser"`
	Toparty string `json:"toparty"`
	Totag   string `json:"totag"`
	Msgtype string `json:"msgtype"`
	Agentid string `json:"agentid"`
	Text    Text   `json:"text"`
	Safe    int    `json:"safe"`
}

type Text struct {
	Content string `json:"content"`
}

func (c ChartAlarm) SendText(content string) error {
	myload := myLoad{
		ToUser:  "@all",
		Toparty: "",
		Totag:   "",
		Msgtype: "text",
		Agentid: c.AppID,
		Text: Text{
			Content: content,
		},
		Safe: 0,
	}

	sendUrl := c.getSendUrl()
	bytesData, err := json.Marshal(myload)

	log.Println("bytesData", string(bytesData))
	if err != nil {
		return err
	}
	reader := bytes.NewReader(bytesData)

	request, err := http.NewRequest("POST", sendUrl, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	errCode := jsoniter.Get(data, "errcode").ToInt()

	if errCode == 0 {
		return nil
	} else {
		return errors.New("errCode：" + strconv.Itoa(errCode))
	}
}
