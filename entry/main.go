package main

import (
	"20181022/recruit-crawler/saver"
	"20181022/recruit-crawler/types"
	"20181022/recruit-crawler/utils"
	"20181022/recruit-crawler/worker"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"sync"
)

// create Db and saver
var (
	Db    = initDB()
	ch = make(chan types.JobInfoDone)
	Saver = saver.NewSaver(Db, ch)
	wg = &sync.WaitGroup{}
)

func main() {
	defer Db.Close()

	// 启动saver，返回一个channel接收parser送过来的数据，将接收到的数据存往数据库
	Saver.Run()

	//启动智联worker
	zhiLian := worker.ZhiLian{}
	worker.RunWorker(zhiLian, wg, Saver.Ch)

	//启动51job worker
	job51 := worker.Job51{}
	worker.RunWorker(job51, wg, Saver.Ch)

	wg.Wait()
}

func initDB() *sqlx.DB {
	Db, err := sqlx.Open("mysql", "root:Mx560205@tcp(132.232.29.36:3306)/zhaopin")
	if err != nil {
		// 先微信告警
		content:= "招聘爬虫，数据库132.232.29.36连接失败！"
		utils.Chart.SendText(content)
		log.Fatal("connect to mysql 132.232.29.36 failed,", err)
	}
	fmt.Println("connect to mysql success")
	Db.SetMaxOpenConns(150)
	Db.SetMaxIdleConns(50)
	return Db
}
