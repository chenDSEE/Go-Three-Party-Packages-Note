# cron(v3.0.1)

> https://github.com/robfig/cron/tree/v3.0.1



## How to run demo
```bash
go run main.go 
```

- 要手动设置 `cron.SecondOptional` 才能启动对于 second 这个粒度的支持
```go
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

```


## Why *cron* package
- 定时任务的时间指定比 Linux 下的 crontab 更为精细，可以精确到秒
- 比起标准库的 time package，cron package 通过与 crontab 相似的时间、周期描述语言，让任务拥有更灵活的指定、执行周期
  - 可以让定时任务，周期性的在业务地方运行，而且能够很灵活的进行调整



## struct and interface

### Cron struct

#### 初始化

```go
// Cron keeps track of any number of entries, invoking the associated func as
// specified by the schedule. It may be started, stopped, and the entries may
// be inspected while running.
type Cron struct {
	entries   []*Entry // 全部 cron-job 都会在这里
	chain     Chain
	stop      chan struct{}

	// cron.Cron.Start() 之后，必须通过 channel 新增 cron-job
	add       chan *Entry
	// cron.Cron.Start() 之后，必须通过 channel 删除 cron-job
	remove    chan EntryID
	snapshot  chan chan []Entry
	running   bool
	logger    Logger
	runningMu sync.Mutex
	location  *time.Location
	parser    ScheduleParser // 如何解析时间描述字符串
	nextID    EntryID
	// 确保退出时，正在运行的 cron-job 能够完成。而不是做了一半，就直接退出了
	jobWaiter sync.WaitGroup
}


// 通过注入可变参数 Option 的方式来进行初始化，Option 很显然会是一个函数变量，专门用来修改刚刚生成的 Cron 中的部分参数
func New(opts ...Option) *Cron {
	c := &Cron{
		entries:   nil,
		chain:     NewChain(),
		add:       make(chan *Entry),
		stop:      make(chan struct{}),
		snapshot:  make(chan chan []Entry),
		remove:    make(chan EntryID),
		running:   false,
		runningMu: sync.Mutex{},
		logger:    DefaultLogger,
		location:  time.Local,
		parser:    standardParser,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option represents a modification to the default behavior of a Cron.
type Option func(*Cron)

func WithSeconds() Option {
	return WithParser(NewParser(
		Second | Minute | Hour | Dom | Month | Dow | Descriptor,
	))
}
```



#### 加入定时任务

```go
// FuncJob is a wrapper that turns a func() into a cron.Job
type FuncJob func()

// AddFunc adds a func to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque ID is returned that can be used to later remove it.
// 最后都是使用 type FuncJob func() 将 cmd 传递给 cron.AddJob()
func (c *Cron) AddFunc(spec string, cmd func()) (EntryID, error) {
	return c.AddJob(spec, FuncJob(cmd))
}

// AddJob adds a Job to the Cron to be run on the given schedule.
// The spec is parsed using the time zone of this Cron instance as the default.
// An opaque ID is returned that can be used to later remove it.
func (c *Cron) AddJob(spec string, cmd Job) (EntryID, error) {
	// cron.WithSecond() 的话，就是调用：cron.Parser.Parse()
	// 然后将时间描述字符串解析为由 cron.SpecSchedule struct impl 的 Schedule interface
	// 这样 cron.Cron 能够通过 SpecSchedule.Next() 来询问这个 job 下一次触发是什么时候，
	// 从而将所有的定时 job 进行排序
	schedule, err := c.parser.Parse(spec)
	if err != nil {
		return 0, err
	}
	return c.Schedule(schedule, cmd), nil
}
```



#### 启动并执行周期任务

- `cron.Cron` 一旦启动之后，都是通过 channel 来完成增删 cron-job 操作的
  - 否则将会被 `c.runningMu.Lock()` block 住

```go
// Start the cron scheduler in its own goroutine, or no-op if already started.
// 创建一个新的 goroutine 来监控定期任务的执行
func (c *Cron) Start() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		// fast-path 的函数短一点，有利于被 inline
		// 减少函数调用的次数
		return
	}
	c.running = true
	go c.run() // 启动一个新的 goroutine 作为监控
}

// Run the cron scheduler, or no-op if already running.
// 在当前 goroutine 中监控定期任务的执行
func (c *Cron) Run() {
	c.runningMu.Lock()
	if c.running {
		c.runningMu.Unlock()
		return
	}
	c.running = true
	c.runningMu.Unlock()
	c.run()
}

// run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
// 总是被 c.runningMu.Lock() 保护着
func (c *Cron) run() {

	// Figure out the next activation times for each entry.
	/* 确认当前时间，以及每一个 entry 下一次激活的时间 */
	now := c.now() // 获取当前时间
	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now) // update 每一个 entry 下一次执行的时间点
		c.logger.Info("schedule", "now", now, "entry", entry.ID, "next", entry.Next)
	}

	/* 无限循环 to handler cron-job */
	// 只有当退出信号到来，才会退出这个循环
	// 当 cron.Cron 运行之后，新的操作都是通过 channel 来完成的
	//
	// 做了什么：
	// for {
	//   0. 从 cron.Cron.entries 中取出最快被激活的 cron-job，并为它创建一个定时器进行监控
	//   1. 最近的 cron-job 需要被激活(case now = <-timer.C:)
	//   2. 新增 cron-job(case newEntry := <-c.add:)
	//   3. 相应 Entries() 的 deep copy 请求()
	//   4. 退出信号(casa <-c.stop:)
	//   5. 删除 cron-job(case id := <-c.remove:)
	// }
	for {
		// Determine the next entry to run.
		// 把所有定时任务，从最近到最远的顺序排列，并把顺序存储在 cron.Cron.entries 中
		sort.Sort(byTime(c.entries))

		var timer *time.Timer
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			....
		} else {
			// 创建一个 timer，并设置为最先被激活的任务的时间差
			// now - next-job-active-time
			timer = time.NewTimer(c.entries[0].Next.Sub(now))
		}

		........
			select {
			case now = <-timer.C:
				.......
				// 1. 最近的 cron-job 需要被激活(case now = <-timer.C:)
				// Run every entry whose next time was less than now
				for _, e := range c.entries {
					// 因为 c.entries 是已经按着触发顺序，由近到远排列好了，所以可以按顺序遍历
					// 把可以触发的 cron-job 也触发掉
					if e.Next.After(now) || e.Next.IsZero() {
						break
					}
					c.startJob(e.WrappedJob)
					.....
				}

			case newEntry := <-c.add:
				// 2. 新增 cron-job(case newEntry := <-c.add:)
				timer.Stop()
				now = c.now()  // for next for-loop round, 因为 timer 要重新设定了
				newEntry.Next = newEntry.Schedule.Next(now)
				c.entries = append(c.entries, newEntry) // 安全，因为由 c.runningMu.Lock() 保护
				c.logger.Info("added", "now", now, "entry", newEntry.ID, "next", newEntry.Next)

			case replyChan := <-c.snapshot:
				// 3. 相应 Entries() 的 deep copy 请求()
				replyChan <- c.entrySnapshot()
				continue

			case <-c.stop:
				// 4. 退出信号(casa <-c.stop:)
				timer.Stop()
				c.logger.Info("stop")
				return

			case id := <-c.remove:
				// 5. 删除 cron-job(case id := <-c.remove:)
				timer.Stop()
				now = c.now() // for next for-loop round, 因为 timer 要重新设定了
				c.removeEntry(id) // 遍历，O(n) 复杂度
				c.logger.Info("removed", "entry", id)
			}
		.......
	}
}

// startJob runs the given job in a new goroutine.
// 利用 sync.WaitGroup 来确保退出时，正在运行的 cron-job 能够全部执行完成
func (c *Cron) startJob(j Job) {
	c.jobWaiter.Add(1)
	go func() {
		defer c.jobWaiter.Done()
		j.Run() // 调用相应的 callback
	}()
}

```









### Schedule interface, SpecSchedule struct

- `Schedule` interface 的核心是：告诉 `cron.Cron`，某一个 callback、`cron.Entry`、`cron.Job` 下一个被激活的时刻是什么时候

```go
// Schedule describes a job's duty cycle.
// Schedule interface de 唯一能力是：告诉调用者，
// 下一个时刻是什么时候，然后调用者根据这个触发时刻对不同的 job 进行排序
type Schedule interface {
	// Next returns the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}


// SpecSchedule specifies a duty cycle (to the second granularity), based on a
// traditional crontab specification. It is computed initially and stored as bit sets.
type SpecSchedule struct {
	// 虽然实现 Schedule interface 只需要不停告知下一个运行时刻
	// 但是不把相应的 Second, Minute, Hour, Dom, Month, Dow 保存起来
	// 是做不到周期性计算的
	Second, Minute, Hour, Dom, Month, Dow uint64

	// Override location for this schedule.
	Location *time.Location
}
```



### Entry struct

- 每一个定时 cron-job 最后都会被封装为 `cron.Entry` 然后在放进 `cron.Cron.entries` 进行管理

```go
// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// ID is the cron-assigned ID of this entry, which may be used to look up a
	// snapshot or remove it.
	ID EntryID

	// Schedule on which this job should be run.
    // 回答 cron.Cron 下一次触发的时间点是？
    // 总是通过 Entry.Schedule.Next() 来更新
	Schedule Schedule

	Next time.Time
	Prev time.Time

	// Q&A(DONE): 这样 wrapped 起来的目的是什么？
	// 链式调用，
	// c := cron.New(cron.WithChain(logWrapper))
	WrappedJob Job
	Job Job
}
```









