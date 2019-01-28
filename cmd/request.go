package cmd

import (
	"fmt"
	"sync"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/vbauerster/mpb"
	"github.com/vbauerster/mpb/decor"
)

const (
	parallelTask = 20
)

type task struct {
	egoscale.AsyncCommand
	string
}

type taskStatus struct {
	id        int
	jobStatus egoscale.JobStatusType
}

type taskResponse struct {
	resp interface{}
	error
}

func asyncTasks(tasks []task) []taskResponse {

	//init results
	responses := make([]taskResponse, len(tasks))

	//create task Progress
	taskBars := make([]*mpb.Bar, len(tasks))
	maximum := 1 << 30
	var taskWG sync.WaitGroup
	p := mpb.New(
		mpb.WithWaitGroup(&taskWG),
		mpb.WithContext(gContext),
	)
	taskWG.Add(len(tasks))

	var workerWG sync.WaitGroup
	workerWG.Add(len(tasks))
	workerSem := make(chan int, parallelTask)

	max := 50 * time.Millisecond

	//exec task and init bars
	for i, task := range tasks {
		c := make(chan taskStatus)
		go execTask(task, i, c, &responses[i], workerSem, &workerWG)
		taskBars[i] = p.AddSpinner(int64(maximum),
			mpb.SpinnerOnLeft,
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(task.string),
			),
		)

		//listen for bar progress
		go func(channel chan taskStatus, idx int) {
			defer taskWG.Done()
			defer close(channel)

			start := time.Now()

			// for select + sleep
			for {
				select {
				case status := <-channel:
					if status.jobStatus != egoscale.Pending {
						taskBars[idx].IncrBy(maximum, time.Since(start))
						return
					}
				case <-time.After(max):
					// do nothing
				}
				taskBars[idx].IncrBy(1, time.Since(start))
			}
		}(c, i)
	}

	workerWG.Wait()
	p.Wait()

	return responses
}

func execTask(task task, id int, c chan taskStatus, resp *taskResponse, sem chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	sem <- 1

	response := cs.Response(task.AsyncCommand)
	var errorReq error
	cs.AsyncRequestWithContext(gContext, task.AsyncCommand, func(jobResult *egoscale.AsyncJobResult, err error) bool {
		if err != nil {
			errorReq = err
			return false
		}

		if jobResult.JobStatus != egoscale.Pending {
			if errR := jobResult.Result(response); errR != nil {
				errorReq = errR
			}
			return false
		}

		c <- taskStatus{id, egoscale.Pending}
		return true
	})

	if errorReq == nil {
		(*resp).resp = response
		c <- taskStatus{id, egoscale.Success}
	} else {
		c <- taskStatus{id, egoscale.Failure}
		(*resp).error = fmt.Errorf("failure %s: %s", task.string, errorReq)
	}

	<-sem
}

// filterErrors return all task with an error
func filterErrors(tasks []taskResponse) []error {
	var r []error
	for _, task := range tasks {
		if task.error != nil {
			r = append(r, task.error)
		}
	}
	return r
}

// asyncRequest if no response expected send nil
func asyncRequest(cmd egoscale.AsyncCommand, msg string) (interface{}, error) {
	response := cs.Response(cmd)

	fmt.Print(msg)
	var errorReq error
	cs.AsyncRequestWithContext(gContext, cmd, func(jobResult *egoscale.AsyncJobResult, err error) bool {

		fmt.Print(".")

		if err != nil {
			errorReq = err
			return false
		}

		if jobResult.JobStatus == egoscale.Pending {
			return true
		}

		if errR := jobResult.Result(response); errR != nil {
			errorReq = errR
			return false
		}

		fmt.Println(" success.")
		return false
	})

	if errorReq != nil {
		fmt.Println(" failure!")
	}

	return response, errorReq
}
