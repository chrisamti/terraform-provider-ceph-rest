package ceph

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"time"
)

// MetaData implements struct to receive task from type rbd delete or create and is used in exceptions.
type MetaData struct {
	PoolName  string  `json:"pool_name"`
	Namespace *string `json:"namespace"`
	ImageName string  `json:"image_name"`
	ImageSpec string  `json:"image_spec"`
}

// Exception implements struct returned on http 400 responses.
type Exception struct {
	Detail    string `json:"detail"`
	Code      string `json:"code"`
	Component string `json:"component"`
	Status    int    `json:"status"`
	Task      struct {
		Name     string   `json:"name"`
		MetaData MetaData `json:"metadata"`
	} `json:"task"`
}

// Task implements a task struct returned for task finished.
type Task struct {
	Name string `json:"name"`
	// MetaData MetaData `json:"metadata"`
	MetaData  MetaData    `json:"metadata"`
	BeginTime time.Time   `json:"begin_time"`
	EndTime   time.Time   `json:"end_time"`
	Duration  float64     `json:"duration"`
	Progress  int         `json:"progress"`
	Success   bool        `json:"success"`
	RetValue  interface{} `json:"ret_value"`
	Exception Exception   `json:"exception"`
}

// Tasks implements the struct returned by /api/task.
type Tasks struct {
	ExecutingTasks []Task `json:"executing_tasks"`
	FinishedTasks  []Task `json:"finished_tasks"`
}

// GetTask get tasks (https://docs.ceph.com/en/latest/mgr/ceph_api/#get--api-task)
func (c *Client) GetTask() (int, Tasks, error) {
	var resp *resty.Response

	var err error
	var t Tasks

	client := *c.Session.Client

	resp, err = client.
		SetRetryCount(10).
		SetRetryWaitTime(10 * time.Second).
		R().
		SetHeaders(defaultHeaderJson).
		SetResult(&t).
		Get(c.Session.Server.getURL("task"))

	if !resp.IsSuccess() {
		return resp.StatusCode(), Tasks{}, fmt.Errorf("%v", resp.RawResponse)
	}

	c.Logger.Debugf("%d tasks executing, %d tasks finished", len(t.ExecutingTasks), len(t.FinishedTasks))
	// fmt.Printf("%s\n", resp.Body())
	return resp.StatusCode(), t, err
}

func (c *Client) WaitForTaskIsDone(workTask Task) (Task, error) {
	// var status int
	var finishedTask Task
	var maxAttempts = 600
	// var tasks Tasks
	var isFinished bool
	// var err error

	// check if workTask is still processed
	for {
		_, tasks, err := c.GetTask() // get workTask also does retries...
		if err != nil {
			return finishedTask, err

		}
		if !taskIsStillExecuting(workTask, tasks) {
			break
		}
		if workTask.MetaData.ImageSpec != "" {
			c.Logger.Debugf("still executing: %s %s", workTask.Name, workTask.MetaData.ImageSpec)

		} else {
			c.Logger.Debugf("still executing: %s %s\n", workTask.Name,
				PathJoin(workTask.MetaData.PoolName, workTask.MetaData.Namespace, workTask.MetaData.ImageName))
		}

		time.Sleep(5 * time.Second) // wait 5 seconds...
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		_, tasks, err := c.GetTask() // get workTask also does retries...

		if err != nil {
			return finishedTask, err
		}

		// look into executing workTask
		if isFinished, finishedTask = taskIsFinished(workTask, tasks); isFinished {
			c.Logger.Debugf("finished: %s %s", workTask.Name, workTask.MetaData.ImageSpec)
			return finishedTask, nil
		}

		c.Logger.Debugf("still not done : %s %s %d", workTask.Name, workTask.MetaData.ImageSpec, attempt)
		time.Sleep(5 * time.Second) // wait 5 seconds...
	}

	return finishedTask, nil
}

func taskIsStillExecuting(task Task, tasks Tasks) bool {
	// var imageSpec = fmt.Sprintf("%s/%s", st.MetaData.PoolName, st.MetaData.ImageName)
	for _, et := range tasks.ExecutingTasks {
		if matchTask(et, task) {
			return true
		}
	}
	return false
}

func taskIsFinished(task Task, tasks Tasks) (bool, Task) {
	// var imageSpec = fmt.Sprintf("%s/%s", st.MetaData., st.MetaData.ImageName)
	for _, ft := range tasks.FinishedTasks {
		if matchTask(ft, task) {
			return true, ft
		}
	}

	return false, Task{}
}

func matchTask(finishedTask, workTask Task) bool {
	if finishedTask.Name == workTask.Name &&
		finishedTask.MetaData.ImageSpec == workTask.MetaData.ImageSpec &&
		finishedTask.MetaData.Namespace == workTask.MetaData.Namespace &&
		finishedTask.MetaData.ImageName == workTask.MetaData.ImageName &&
		finishedTask.MetaData.PoolName == workTask.MetaData.PoolName {
		return true
	}

	return false
}
