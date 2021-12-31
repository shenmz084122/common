package flink

import (
	"fmt"
	"testing"
	"time"
)

var client *Client

func init() {
	config := ClientConfig{
		FlinkRestUrl:  "http://127.0.0.1:8081",
		Timeout:       2000 * time.Millisecond,
		RetryCount:    2,
		QueryInterval: 2000,
	}
	client = NewFlinkClient(config)
}

func Test_ListJobs(t *testing.T) {
	jobs, err := client.listJobs()
	if err != nil {
		t.Error(err)
	}
	for _, job := range jobs {
		fmt.Println(job)
	}
}

func Test_GetJob(t *testing.T) {
	job, err := client.getJobInfoByJobId("d6c331d9ff06d75517ab9946cf884fea")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(job)
}
