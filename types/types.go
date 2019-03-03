package types

import (
	"20181022/recruit-crawler/utils"
	"github.com/json-iterator/go"
	"sync"
)

type JobInfo struct {
	InfoSource  string
	CreateDate  string
	UpdateDate  string
	City        string
	PositionURL string
	Salary      string
	SalaryMin   float64
	SalaryMax   float64
	CompanyName string
	JobName     string
}

type JobInfoDone struct {
	Wg *sync.WaitGroup
	JobInfo
}

func CreateJobInfoDone(wg *sync.WaitGroup, item jsoniter.Any, infoSource string) *JobInfoDone {
	var (
		createDate           = item.Get("createDate").ToString()
		updateDate           = item.Get("updateDate").ToString()
		city                 = item.Get("city").Get("display").ToString()
		positionURL          = item.Get("positionURL").ToString()
		salary               = item.Get("salary").ToString()
		salaryMin, salaryMax = utils.GetSalaryInfo(salary)
		jobName              = item.Get("jobName").ToString()
		companyName          = item.Get("company").Get("name").ToString()
	)

	return &JobInfoDone{
		Wg: wg,
		JobInfo: JobInfo{
			InfoSource:  infoSource,
			CreateDate:  createDate,
			UpdateDate:  updateDate,
			City:        city,
			PositionURL: positionURL,
			Salary:      salary,
			SalaryMin:   salaryMin,
			SalaryMax:   salaryMax,
			JobName:     jobName,
			CompanyName: companyName,
		},
	}
}