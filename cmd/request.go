package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/exoscale/egoscale"
	"github.com/hashicorp/go-multierror"
	"github.com/vbauerster/mpb/v4"
	"github.com/vbauerster/mpb/v4/decor"
)

const (
	parallelTask = 20
)

type task struct {
	egoscale.Command
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
	progress := mpb.NewWithContext(gContext,
		mpb.WithWaitGroup(&taskWG),
		mpb.ContainerOptOnCond(mpb.WithOutput(nil), func() bool { return gQuiet }),
	)

	taskWG.Add(len(tasks))

	var workerWG sync.WaitGroup
	workerWG.Add(len(tasks))
	workerSem := make(chan int, parallelTask)

	max := 50 * time.Millisecond

	//exec task and init bars
	for i, task := range tasks {
		c := make(chan taskStatus)
		switch cmd := task.Command.(type) {
		case egoscale.AsyncCommand:
			go execTask(cmd, task.string, i, c, &responses[i], workerSem, &workerWG)
		default:
			go execSyncTask(task, i, c, &responses[i], workerSem, &workerWG)
		}
		taskBars[i] = progress.AddSpinner(int64(maximum),
			mpb.SpinnerOnLeft,
			mpb.PrependDecorators(
				// simple name decorator
				decor.Name(task.string),
			),
			mpb.AppendDecorators(decor.OnComplete(decor.Elapsed(decor.ET_STYLE_GO), "done")),
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
	progress.Wait()

	return responses
}

func execTask(task egoscale.AsyncCommand, message string, id int, c chan taskStatus, resp *taskResponse, sem chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	sem <- 1

	response := cs.Response(task)
	var errorReq error
	cs.AsyncRequestWithContext(gContext, task, func(jobResult *egoscale.AsyncJobResult, err error) bool {
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
		(*resp).error = fmt.Errorf("failure %s: %s", message, errorReq)
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

func execSyncTask(task task, id int, c chan taskStatus, resp *taskResponse, sem chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	sem <- 1

	_, ok := cs.Response(task.Command).(*egoscale.BooleanResponse)
	if ok {
		if err := cs.BooleanRequestWithContext(gContext, task.Command); err != nil {
			c <- taskStatus{id, egoscale.Failure}
			(*resp).error = fmt.Errorf("failure %s: %s", task.string, err)
			return
		}
		c <- taskStatus{id, egoscale.Success}
		<-sem
		return
	}

	result, err := cs.RequestWithContext(gContext, task.Command)
	if err != nil {
		c <- taskStatus{id, egoscale.Failure}
		(*resp).error = fmt.Errorf("failure %s: %s", task.string, err)
		return
	}
	(*resp).resp = result
	c <- taskStatus{id, egoscale.Success}
	<-sem
}

// asyncRequest if no response expected send nil
func asyncRequest(cmd egoscale.AsyncCommand, msg string) (interface{}, error) {
	response := cs.Response(cmd)

	if !gQuiet {
		fmt.Fprint(os.Stderr, msg)
	}
	var errorReq error
	cs.AsyncRequestWithContext(gContext, cmd, func(jobResult *egoscale.AsyncJobResult, err error) bool {
		if !gQuiet {
			fmt.Fprint(os.Stderr, ".")
		}

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

		if !gQuiet {
			fmt.Fprintln(os.Stderr, " success")
		}

		return false
	})

	if errorReq != nil && !gQuiet {
		fmt.Fprintln(os.Stderr, " failure")
	}

	return response, errorReq
}

// forEachZone executes the function f for each specified zone, and return a multierror.Error containing all
// errors that may have occurred during execution.
func forEachZone(zones []string, f func(zone string) error) error {
	var meg = new(multierror.Group)

	for _, zone := range zones {
		zone := zone
		meg.Go(func() error {
			return f(zone)
		})
	}

	return meg.Wait().ErrorOrNil()
}
