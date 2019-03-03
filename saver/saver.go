package saver

import (
	"20181022/crawler/types"
	"20181022/crawler/utils"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"log"
)

type Saver struct {
	Db *sqlx.DB
	Ch chan types.JobInfoDone
}

func NewSaver(db *sqlx.DB, ch chan types.JobInfoDone) *Saver {
	return &Saver{Db: db, Ch: ch}
}

var (
//Db, err = sqlx.Open("mysql", "root:Mx560205@tcp(132.232.29.36:3306)/zhaopin")
// 定义全局save channel，所有的需要存储的数据都往这里发
//SaverChannel = CreateSaver()

)

func (s *Saver) Run() {

	go func() {
		//执行sql语句，切记这里的占位符是？
		for job := range s.Ch {
			// TODO 这里起goroutine?
			// 插入数据前，要保证不插入重复数据，这里先把数据查出来再程序中去重，因为去重后的数据再程序后续中还要用
			// 这里定义一下，怎么判断数据是否重复，以下3点同时满足则视为重复，不插入数据库
			// 1 CompanyName重复
			// 2 JobName重复
			// 3 Salary重复
			go filterAndSave(job, s.Db)
		}
	}()

	//test，在新建一个网站的worker时，可以先用下面的代码测试
	//go func() {
	//	count := 0
	//	for rest := range s.Ch {
	//		count ++
	//		fmt.Printf("count：%d\n公司名：%s\n工作名：%s\n工资：%s\n工资下限：%3f\n工资上限：%3f\nurl：%s\n",
	//			count, rest.CompanyName, rest.JobName, rest.Salary, rest.SalaryMin, rest.SalaryMax, rest.PositionURL)
	//		fmt.Println("----------------------------------------------------------------------------------")
	//		rest.Wg.Done()
	//	}
	//}()
}

func filterAndSave(job types.JobInfoDone, db *sqlx.DB) {
	rows := db.QueryRow("SELECT count(1) FROM job_info WHERE company_name=? AND job_name=? AND salary=?", job.CompanyName, job.JobName, job.Salary)
	count := 0
	err := rows.Scan(&count)
	if err != nil {
		log.Println("err--->", err)
	}
	// 如果找不到，那么说明没有和当前数据重复，可以继续插入该数据
	if count == 0 {
		// 自己补充需要过滤的或者需要执行的其他动作
		// TODO 看怎么实现每一个网站传入专门的func在saver中做自己的事，而不是现在这样统一一种处理方式
		if job.CompanyName == "中移物联网有限公司" {
			job.Wg.Done()
			return
		}
		// 插入数据库前判断，如果工资大于10K，则发起告警。具体告警方式，可以用微信企业号或者其他？
		if job.SalaryMax >= 10000 {
			// 发送微信告警
			log.Println("准备发送微信告警")
			content := fmt.Sprintf("来源：%s\nCity：%s\nCompanyName：%s\nJobName：%s\nSalary：%s\nPositionURL：%s\ncreateDate：%s\nupdateDate：%s\n",
				job.InfoSource,
				job.City,
				job.CompanyName,
				job.JobName,
				job.Salary,
				job.PositionURL,
				job.CreateDate,
				job.UpdateDate)

			err := utils.Chart.SendText(content)
			log.Println("to send chart_alarm failed：", err.Error())
			log.Println("failed content：", content)
		}

		log.Println("准备插入新数据")
		_, err := db.Exec("INSERT INTO job_info(info_source,create_date,update_date,city,position_url,salary,salary_min,salary_max,company_name,job_name)VALUES (?,?,?,?,?,?,?,?,?,?)",
			job.InfoSource,
			job.CreateDate,
			job.UpdateDate,
			job.City,
			job.PositionURL,
			job.Salary,
			job.SalaryMin,
			job.SalaryMax,
			job.CompanyName,
			job.JobName, )
		if err != nil {
			log.Println("insert failed,", err)
		}
		job.Wg.Done()

		//通过LastInsertId可以获取插入数据的id
		//userId, err := result.LastInsertId()
		//通过RowsAffected可以获取受影响的行数
		//rowCount, err := result.RowsAffected()
		//fmt.Println("user_id:", userId)
		//fmt.Println("rowCount:", rowCount)
	} else {
		job.Wg.Done()
	}

}
