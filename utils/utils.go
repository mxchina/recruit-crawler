package utils

import (
	"bufio"
	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func GetSalaryInfo(s string) (min float64, max float64) {
	// s可能的结果
	// 1 6K-8K
	// 2 薪资面议
	if s == "薪资面议" {
		return 0, 0
	}
	splits := strings.Split(s, "-")

	if len(splits) != 2 {
		return 0, 0
	}

	min = StrToFloat64(splits[0][:len(splits[0])-1], 3)
	max = StrToFloat64(splits[1][:len(splits[1])-1], 3)

	return min, max
}

func DetermineEncoding(r *bufio.Reader) encoding.Encoding {
	bytes, err := r.Peek(1024)
	if err != nil {
		log.Printf("Fetcher error: %v", err)
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(bytes, "")
	return e
}

func Get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("fetch err：", err)
		return []byte{}, err
	}
	defer resp.Body.Close()

	bodyReader := bufio.NewReader(resp.Body)
	e := DetermineEncoding(bodyReader)
	utf8Reader := transform.NewReader(bodyReader, e.NewDecoder())

	bytes, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		return []byte{}, err
	}
	return bytes, nil
}
