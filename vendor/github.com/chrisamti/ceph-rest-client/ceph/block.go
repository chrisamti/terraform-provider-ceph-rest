package ceph

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-resty/resty/v2"
)

// ErrImageSpecIsEmpty is returned if param imageSpec is empty.
var ErrImageSpecIsEmpty = errors.New("param imageSpec can not be empty")

// ErrPoolNameIsEmpty is returned if param poolName is empty.
var ErrPoolNameIsEmpty = errors.New("param poolName can not be empty")

// ErrImageNameIsEmpty is returned if param imageName is empty.
var ErrImageNameIsEmpty = errors.New("param imageName can not be empty")

// ErrCreateImageAlreadyExists is return if image to be created already exists.
var ErrCreateImageAlreadyExists = errors.New("RBD image already exists (error creating image)")

// ErrEditImageAlreadyExists is return if image to be renamed already exists.
var ErrEditImageAlreadyExists = errors.New("RBD image already exists (error renaming image)")

const (
	RBDImageAlreadyExists = "17"
)

// RBD implements struct returned from GET /api/block/image/{image_spec}
// --> https://docs.ceph.com/en/latest/mgr/ceph_api/#get--api-block-image-image_spec.
type RBD struct {
	Size            int64         `json:"size"`
	ObjSize         int           `json:"obj_size"`
	NumObjs         int           `json:"num_objs"`
	Order           int           `json:"order"`
	BlockNamePrefix string        `json:"block_name_prefix"`
	Name            string        `json:"name"`
	UniqueID        string        `json:"unique_id"`
	ID              string        `json:"id"`
	ImageFormat     int           `json:"image_format"`
	PoolName        string        `json:"pool_name"`
	Namespace       interface{}   `json:"namespace"`
	Features        int           `json:"features"`
	FeaturesName    []string      `json:"features_name"`
	Timestamp       time.Time     `json:"timestamp"`
	StripeCount     int           `json:"stripe_count"`
	StripeUnit      int           `json:"stripe_unit"`
	DataPool        interface{}   `json:"data_pool"`
	Parent          interface{}   `json:"parent"`
	Snapshots       []interface{} `json:"snapshots"`
	TotalDiskUsage  int           `json:"total_disk_usage"`
	DiskUsage       int           `json:"disk_usage"`
	Configuration   []struct {
		Name   string `json:"name"`
		Value  string `json:"value"`
		Source int    `json:"source"`
	} `json:"configuration"`
}

// RBDList implements struct received from GET /api/block/image.
// --> https://docs.ceph.com/en/latest/mgr/ceph_api/#get--api-block-image.
type RBDList []struct {
	Status   int    `json:"status"`
	Value    []RBD  `json:"value"`
	PoolName string `json:"pool_name"`
}

// RBDCreate implements struct send to ceph for rbd image creation on POST /api/block/image.
// --> https://docs.ceph.com/en/latest/mgr/ceph_api/#post--api-block-image
type RBDCreate struct {
	Features      []string    `json:"features"`
	PoolName      string      `json:"pool_name"`
	Namespace     *string     `json:"namespace"`
	Name          string      `json:"name"`
	Size          int         `json:"size"`
	ObjSize       int         `json:"obj_size"`
	StripeUnit    interface{} `json:"stripe_unit"`
	StripeCount   interface{} `json:"stripe_count"`
	DataPool      interface{} `json:"data_pool"`
	Configuration struct{}    `json:"configuration"`
}

// RBDError implements error struct returned.
type RBDError struct {
	Detail    string `json:"detail"`
	Code      string `json:"code"`
	Component string `json:"component"`
	Status    int    `json:"status"`
	Task      struct {
		Name     string `json:"name"`
		Metadata struct {
			PoolName  string      `json:"pool_name"`
			Namespace interface{} `json:"namespace"`
			ImageName string      `json:"image_name"`
		} `json:"metadata"`
	} `json:"task"`
}

// RBDUpdate implements struct send to ceph for rbd image updates on PUT /api/block/image/{image_spec}.
// --> https://docs.ceph.com/en/latest/mgr/ceph_api/#put--api-block-image-image_spec
type RBDUpdate struct {
	Features      []string `json:"features"`
	Name          string   `json:"name"`
	Size          int64    `json:"size"`
	Configuration struct{} `json:"configuration"`
}

// ListBlockImage gets a list of RBD block images (https://docs.ceph.com/en/latest/mgr/ceph_api/#get--api-block-image)
func (c *Client) ListBlockImage(poolName string) (status int, rbdList RBDList, err error) {
	var resp *resty.Response

	client := *c.Session.Client

	if poolName != "" {
		resp, err = client.R().
			SetHeaders(defaultHeaderJson).
			SetQueryParam("pool_name", poolName).
			SetResult(&rbdList).
			Get(c.Session.Server.getURL("block/image"))
	} else {
		resp, err = client.R().
			SetHeaders(defaultHeaderJson).
			SetResult(&rbdList).
			Get(c.Session.Server.getURL("block/image"))

	}

	if !resp.IsSuccess() {
		return resp.StatusCode(), nil, fmt.Errorf("%v", resp.RawResponse)
	}

	return resp.StatusCode(), rbdList, err

}

// GetBlockImage gets an RBD block image (https://docs.ceph.com/en/latest/mgr/ceph_api/#get--api-block-image-image_spec)
func (c *Client) GetBlockImage(imageSpec string) (status int, rbd RBD, err error) {
	var resp *resty.Response

	if imageSpec == "" {
		return 0, rbd, ErrImageSpecIsEmpty
	}

	client := *c.Session.Client

	resp, err = client.
		R().
		SetHeaders(defaultHeaderJson).
		SetResult(&rbd).
		Get(c.Session.Server.getURL(fmt.Sprintf("block/image/%s", url.QueryEscape(imageSpec))))

	if !resp.IsSuccess() {
		return resp.StatusCode(), rbd, fmt.Errorf("could not get image %v: %v", imageSpec, resp.Error())
	}

	return resp.StatusCode(), rbd, err
}

// CreateBlockImage creates an RBD image (https://docs.ceph.com/en/latest/mgr/ceph_api/#post--api-block-image)
func (c *Client) CreateBlockImage(rbdCreate RBDCreate, counter uint) (status int, err error) {

	if counter > c.MaxIterations {
		return 0, ErrMaxIterationsExceeded
	}

	counter++

	var (
		resp      *resty.Response
		exception Exception
	)

	// create copy of client
	client := *c.Session.Client

	resp, err = client.
		SetRetryCount(10).
		SetRetryWaitTime(10 * time.Second).
		AddRetryCondition(c.retryConditionCheckForAccepted).
		R().
		SetHeaders(defaultHeaderJson).
		SetBody(rbdCreate).
		Post(c.Session.Server.getURL("block/image"))

	if err != nil {
		return 0, err
	}

	if !resp.IsSuccess() {
		if resp.StatusCode() == http.StatusBadRequest {
			err = client.JSONUnmarshal(resp.Body(), &exception)
			if err == nil {
				c.Logger.Debugf("err %s (%s)", exception.Code, exception.Detail)
				if exception.Code == RBDImageAlreadyExists {
					return resp.StatusCode(), ErrEditImageAlreadyExists
				}
			}
		}

		return resp.StatusCode(), fmt.Errorf("could not create image: %v on pool %v: %v ", rbdCreate.Name, rbdCreate.PoolName, resp.Error())
	}

	status = resp.StatusCode()

	// check task state
	switch status {
	case http.StatusCreated, http.StatusAccepted:
		lookForTask := Task{
			Name: "rbd/create",
			MetaData: MetaData{
				PoolName:  rbdCreate.PoolName,
				Namespace: rbdCreate.Namespace,
				ImageName: rbdCreate.Name,
				ImageSpec: "", // ImageSpec is always empty for rbd/create
			},
		}

		lookForTask, err = c.WaitForTaskIsDone(lookForTask)
		if err != nil {
			return 0, err
		}

		if !lookForTask.Success {
			// try to create again
			c.Logger.Debugf("call CreateBlockImage again with counter %d", counter)
			return c.CreateBlockImage(rbdCreate, counter)
		} else {
			status = http.StatusCreated
		}
	}

	return status, nil
}

// UpdateBlockImage updates ceph rbd image (name, size etc al).
// --> https://docs.ceph.com/en/latest/mgr/ceph_api/#put--api-block-image-image_spec
func (c *Client) UpdateBlockImage(poolName string, nameSpace *string, imageName string, rbdUpdate RBDUpdate, counter uint) (status int, err error) {
	if counter > c.MaxIterations {
		return 0, ErrMaxIterationsExceeded
	}

	counter++

	var (
		resp      *resty.Response
		imageSpec string
		exception Exception
	)

	if poolName == "" {
		return 0, ErrPoolNameIsEmpty
	}

	if imageName == "" {
		return 0, ErrImageNameIsEmpty
	}

	// check rbdUpdate
	if rbdUpdate.Name == "" {
		return 0, ErrImageNameIsEmpty
	}

	// todo: check all rbdUpdate values for validity.

	imageSpec = PathJoin(poolName, nameSpace, imageName)

	client := *c.Session.Client

	resp, err = client.
		SetRetryCount(10).
		SetRetryWaitTime(10 * time.Second).
		AddRetryCondition(c.retryConditionCheckForAccepted).
		R().
		SetHeaders(defaultHeaderJson).
		SetBody(rbdUpdate).
		Put(c.Session.Server.getURL(fmt.Sprintf("block/image/%s", url.PathEscape(imageSpec))))

	if !resp.IsSuccess() {
		if resp.StatusCode() == http.StatusBadRequest {
			err = client.JSONUnmarshal(resp.Body(), &exception)
			if err == nil {
				c.Logger.Debugf("err %s (%s)", exception.Code, exception.Detail)
				if exception.Code == "17" {
					return resp.StatusCode(), ErrCreateImageAlreadyExists
				}
			}
		}

		return resp.StatusCode(), fmt.Errorf("%v", resp.RawResponse)
	}

	status = resp.StatusCode()

	switch status {
	case http.StatusAccepted, http.StatusOK, http.StatusBadRequest:
		lookForTask := Task{
			Name: "rbd/edit",
			MetaData: MetaData{
				ImageSpec: imageSpec, // only ImageSpec is needed on delete
			},
		}

		lookForTask, err = c.WaitForTaskIsDone(lookForTask)

		if err != nil {
			return 0, err
		}

		if !lookForTask.Success {
			// try delete again...
			c.Logger.Debugf("calling DeleteBlockImage with counter %d", counter)
			return c.UpdateBlockImage(poolName, nameSpace, imageName, rbdUpdate, counter)
		} else {
			status = http.StatusOK
			err = nil
		}
	}

	return status, err

}

// DeleteBlockImage deletes an RBD image defined with imageSpec
// (https://docs.ceph.com/en/latest/mgr/ceph_api/#delete--api-block-image-image_spec)
func (c *Client) DeleteBlockImage(poolName string, nameSpace *string, imageName string, counter uint) (status int, err error) {

	if counter > c.MaxIterations {
		return 0, ErrMaxIterationsExceeded
	}

	counter++

	var resp *resty.Response
	var imageSpec string

	if poolName == "" {
		return 0, ErrPoolNameIsEmpty
	}

	if imageName == "" {
		return 0, ErrImageNameIsEmpty
	}

	imageSpec = PathJoin(poolName, nameSpace, imageName)

	client := *c.Session.Client

	resp, err = client.
		SetRetryCount(10).
		SetRetryWaitTime(10 * time.Second).
		AddRetryCondition(c.retryConditionCheckForAccepted).
		R().
		SetHeaders(defaultHeaderJson).
		Delete(c.Session.Server.getURL(fmt.Sprintf("block/image/%s", url.PathEscape(imageSpec))))

	if !resp.IsSuccess() {
		return resp.StatusCode(), fmt.Errorf("%v", resp.RawResponse)
	}

	status = resp.StatusCode()

	switch status {
	case http.StatusAccepted, http.StatusNoContent, http.StatusBadRequest:
		lookForTask := Task{
			Name: "rbd/delete",
			MetaData: MetaData{
				ImageSpec: imageSpec, // only ImageSpec is needed on delete
			},
		}

		lookForTask, err = c.WaitForTaskIsDone(lookForTask)

		if err != nil {
			return 0, err
		}

		if !lookForTask.Success {
			// try delete again...
			c.Logger.Debugf("calling DeleteBlockImage with counter %d", counter)
			return c.DeleteBlockImage(poolName, nameSpace, imageName, counter)
		} else {
			status = http.StatusNoContent
			err = nil
		}
	}

	return status, err
}

func (c *Client) retryConditionCheckForAccepted(r *resty.Response, _ error) bool {
	switch r.StatusCode() {
	case http.StatusOK,
		http.StatusNoContent,
		http.StatusNotFound,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusBadRequest:
		// no retry needed
		c.Logger.Debugf("http status: %d --> no retry for %s", r.StatusCode(), r.Request.URL)
		return false
	}

	c.Logger.Debugf("http status: %d --> retry for %s needed", r.StatusCode(), r.Request.URL)
	return true // retry on all other status codes.
}
