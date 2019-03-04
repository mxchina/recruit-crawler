package main

import (
	"20181022/recruit-crawler/saver"
	"20181022/recruit-crawler/types"
	"20181022/recruit-crawler/utils"
	"20181022/recruit-crawler/worker"
	"flag"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"sync"
	"time"
)

func init() {
	logInit()
}

// create Db and saver
var (
	Db    = initDB()
	ch    = make(chan types.JobInfoDone)
	Saver = saver.NewSaver(Db, ch)
	wg    = &sync.WaitGroup{}
)

func main() {
	defer Db.Close()

	// 启动saver，返回一个channel接收parser送过来的数据，将接收到的数据存往数据库
	Saver.Run()

	// 启动智联worker
	zhiLian := worker.ZhiLian{}
	worker.RunWorker(zhiLian, wg, Saver.Ch)

	// 启动51job worker
	job51 := worker.Job51{}
	worker.RunWorker(job51, wg, Saver.Ch)

	// 启动boss直聘 worker
	bossZhiPin := worker.BossZhiPin{}
	worker.RunWorker(bossZhiPin, wg, Saver.Ch)

	// 启动...

	wg.Wait()
}

func initDB() *sqlx.DB {
	Db, err := sqlx.Open("mysql", saver.MySQL_DSN)
	if err != nil {
		// 先微信告警
		content := "招聘爬虫，数据库132.232.29.36连接失败！"
		utils.Chart.SendText(content)
		log.Fatal("connect to mysql 132.232.29.36 failed,", err)
	}
	log.Println("connect to mysql success")
	Db.SetMaxOpenConns(150)
	Db.SetMaxIdleConns(50)
	return Db
}

func logInit() {
	dateStr := time.Now().Format(`2006-01-02-15`)
	var (
		logDir  = "./"
		logfile = flag.String("logfile",
			logDir+"go-"+dateStr+".log",
			"set logfile and default is "+logDir+"/go-年-月-日-时.log")
	)
	flag.Parse()
	exist, err := PathExists(logDir)
	if err != nil {
		log.Fatal("判断日志目录是否存在，失败：" + err.Error())
	}
	if !exist {
		if err := os.Mkdir(logDir, os.ModePerm); err != nil {
			log.Fatal("日志目录创建失败：" + err.Error())
		}
	}
	file, err := os.OpenFile(*logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm|os.ModeTemporary)
	if err != nil {
		log.Fatal("日志文件创建失败：" + err.Error())
	}

	log.SetOutput(file)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
