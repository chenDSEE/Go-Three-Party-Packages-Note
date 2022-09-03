package main

import (
	"fmt"
	"main/cron"
	"time"
)

func main() {
	// NOTE: default cron.Cron not support second. Add cron.SecondOptional to support second
	// 之所以要在 new 的时候指定不同的参数，是 cron package 对 api 向前兼容
	// 通过 cron.Option 的指定，让 cron.ScheduleParser 知晓相应时间描述字符串的解析方式
	var logWrapper cron.JobWrapper = func(job cron.Job) cron.Job {
		return cron.FuncJob(func() {
			fmt.Printf("one cron-job active\n")
			job.Run()
		})
	}

    c := cron.New(cron.WithSeconds(), cron.WithChain(logWrapper))

	c.AddFunc("* * * * * *", func() {
		fmt.Printf("%v --> 1 second pass...\n", time.Now().Format(time.Stamp))
	})

	//c.AddFunc("@every 10s", func() {
	//	fmt.Printf("%v --> 10 seconds pass...\n", time.Now().Format(time.Stamp))
	//})

	//c.AddFunc("*/2 * * * * *", func() {
	//	fmt.Printf("%v --> 2 seconds pass...\n", time.Now().Format(time.Stamp))
	//})

	//c.AddFunc("0 * * * * *", func() {
	//	fmt.Printf("%v --> 1 minute pass...\n", time.Now().Format(time.Stamp))
	//})

	fmt.Printf("%v, Starting...\n", time.Now().Format(time.Stamp))
	c.Start()

	select{}

	c.Stop()
	fmt.Println("Stoping...")
}
