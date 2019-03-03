package worker

import (
	"20181022/crawler/types"
	"sync"
)

//type Worker struct {
//	FetchFunc func() <-chan []byte
//	ParseFunc func(wg *sync.WaitGroup, from <-chan []byte, to chan types.JobInfoDone)
//}

type Worker interface {
	// make chan and return it, run a goroutine and send data to this chan
	// wg must add in main, recevied the add should at Parse
	Fetch(wg *sync.WaitGroup) <-chan *[]byte
	// Parse handle the wg from Fetch,
	// the data recevied from chan_from and send to Chan_to
	// before send, should hanlde the wg call wg.Add(1)
	Parse(wg *sync.WaitGroup, from <-chan *[]byte, to chan<- types.JobInfoDone)
}

//func (w Worker) Run(wg *sync.WaitGroup, to chan types.JobInfoDone) {
//	from := w.FetchFunc()
//	w.ParseFunc(wg, from, to)
//}
func RunWorker(w Worker, wg *sync.WaitGroup, to chan<- types.JobInfoDone)  {
	// 启动fetcher，返回一个channel，将网上抓取的数据发往这个channel。parse接到这个channel后从里面取数据
	from := w.Fetch(wg)
	// 启动paeser，将fetch返回的channel传进去，可以从这个channel接收数据，然后解析
	// 将解析好的数据送往to，to是saver接收数据的channel
	w.Parse(wg, from, to)
}
