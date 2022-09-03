package cron

import (
	"context"
	"sort"
	"sync"
	"time"
)

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

// ScheduleParser is an interface for schedule spec parsers that return a Schedule
// ScheduleParser interface 需要的能力是：将时间描述字符串解析为一个能够不断回答下一次触发时刻（Schedule）的 obj
// 默认使用 cron.Parser
type ScheduleParser interface {
	Parse(spec string) (Schedule, error)
}

// Job is an interface for submitted cron jobs.
type Job interface {
	Run()
}

// Schedule describes a job's duty cycle.
// Schedule interface de 唯一能力是：告诉调用者，
// 下一个时刻是什么时候，然后调用者根据这个触发时刻对不同的 job 进行排序
type Schedule interface {
	// Next returns the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

// EntryID identifies an entry within a Cron instance
type EntryID int

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// ID is the cron-assigned ID of this entry, which may be used to look up a
	// snapshot or remove it.
	ID EntryID

	// Schedule on which this job should be run.
	// 回答 cron.Cron 下一次触发的时间点是？
	// 用来生成下一次运行的时间
	Schedule Schedule

	// Next time the job will run, or the zero time if Cron has not been
	// started or this entry's schedule is unsatisfiable
	// 总是通过 Entry.Schedule.Next() 来更新
	Next time.Time

	// Prev is the last time this job was run, or the zero time if never.
	Prev time.Time

	// WrappedJob is the thing to run when the Schedule is activated.
	// Q&A(DONE): 这样 wrapped 起来的目的是什么？
	// 链式调用，
	// c := cron.New(cron.WithChain(logWrapper))
	WrappedJob Job

	// Job is the thing that was submitted to cron.
	// It is kept around so that user code that needs to get at the job later,
	// e.g. via Entries() can do so.
	Job Job
}

// Valid returns true if this is not the zero entry.
func (e Entry) Valid() bool { return e.ID != 0 }

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

// New returns a new Cron job runner, modified by the given options.
//
// Available Settings
//
//   Time Zone
//     Description: The time zone in which schedules are interpreted
//     Default:     time.Local
//
//   Parser
//     Description: Parser converts cron spec strings into cron.Schedules.
//     Default:     Accepts this spec: https://en.wikipedia.org/wiki/Cron
//
//   Chain
//     Description: Wrap submitted jobs to customize behavior.
//     Default:     A chain that recovers panics and logs them to stderr.
//
// See "cron.With*" to modify the default behavior.
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

// FuncJob is a wrapper that turns a func() into a cron.Job
type FuncJob func()

func (f FuncJob) Run() { f() }

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

// Schedule adds a Job to the Cron to be run on the given schedule.
// The job is wrapped with the configured Chain.
// 对与 cron.Cron 而言，只需要只需要直到一个 Job 的两个特点：下次触发是什么时候 + 触发时要干什么
func (c *Cron) Schedule(schedule Schedule, cmd Job) EntryID {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	c.nextID++
	entry := &Entry{
		ID:         c.nextID,
		Schedule:   schedule, // Schedule interface 的核心是：告诉 cron.Cron，自己这个 entry 下一个被激活的时刻是？
		WrappedJob: c.chain.Then(cmd),
		Job:        cmd,
	}

	// c.running 是在 cron.Start() 的时候被 set 的
	// TODO: 为什么 running 前后是使用不同的 append 方式，为什么要这么做？
	if !c.running {
		// []*Entry, 并发不安全
		c.entries = append(c.entries, entry)
	} else {
		// chan *Entry, 并发安全
		// 而且这是一个 unbuffered channel，会把接收跟发送者都 block 住
		c.add <- entry
	}
	return entry.ID
}

// Entries returns a snapshot of the cron entries.
// 会被 block 住
func (c *Cron) Entries() []Entry {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		replyChan := make(chan []Entry, 1)
		c.snapshot <- replyChan
		return <-replyChan
	}
	return c.entrySnapshot()
}

// Location gets the time zone location
func (c *Cron) Location() *time.Location {
	return c.location
}

// Entry returns a snapshot of the given entry, or nil if it couldn't be found.
func (c *Cron) Entry(id EntryID) Entry {
	for _, entry := range c.Entries() {
		if id == entry.ID {
			return entry
		}
	}
	return Entry{}
}

// Remove an entry from being run in the future.
func (c *Cron) Remove(id EntryID) {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		c.remove <- id
	} else {
		c.removeEntry(id)
	}
}

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
	c.logger.Info("start")

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
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			timer = time.NewTimer(100000 * time.Hour)
		} else {
			// 创建一个 timer，并设置为最先被激活的任务的时间差
			// now - next-job-active-time
			timer = time.NewTimer(c.entries[0].Next.Sub(now))
		}

		for {
			select {
			case now = <-timer.C:
				// 1. 最近的 cron-job 需要被激活(case now = <-timer.C:)
				now = now.In(c.location)
				c.logger.Info("wake", "now", now)

				// Run every entry whose next time was less than now
				for _, e := range c.entries {
					// 因为 c.entries 是已经按着触发顺序，由近到远排列好了，所以可以按顺序遍历
					// 把可以触发的 cron-job 也触发掉
					if e.Next.After(now) || e.Next.IsZero() {
						break
					}
					c.startJob(e.WrappedJob)
					e.Prev = e.Next
					e.Next = e.Schedule.Next(now) // 下一次 for-loop round 重新排序
					c.logger.Info("run", "now", now, "entry", e.ID, "next", e.Next)
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

			break // 必定 break，其实这一层 for-loop 就没啥必要了
		}
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

// now returns current time in c location
func (c *Cron) now() time.Time {
	return time.Now().In(c.location)
}

// Stop stops the cron scheduler if it is running; otherwise it does nothing.
// A context is returned so the caller can wait for running jobs to complete.
func (c *Cron) Stop() context.Context {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	if c.running {
		c.stop <- struct{}{}
		c.running = false
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		c.jobWaiter.Wait()
		cancel()
	}()
	return ctx
}

// entrySnapshot returns a copy of the current cron entry list.(deep copy)
func (c *Cron) entrySnapshot() []Entry {
	var entries = make([]Entry, len(c.entries))
	for i, e := range c.entries {
		entries[i] = *e
	}
	return entries
}

// 直接创建新的 entries
func (c *Cron) removeEntry(id EntryID) {
	var entries []*Entry
	for _, e := range c.entries {
		if e.ID != id {
			entries = append(entries, e)
		}
	}
	c.entries = entries
}
