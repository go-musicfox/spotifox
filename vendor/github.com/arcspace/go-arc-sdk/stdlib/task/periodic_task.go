package task

import (
	"github.com/arcspace/go-arc-sdk/stdlib/utils"
)

type PeriodicTask struct {
	Context
	ticker  utils.Ticker
	mailbox *utils.Mailbox
	taskFn  func(ctx Context)
}

func NewPeriodicTask(name string, ticker utils.Ticker, taskFn func(ctx Context)) *PeriodicTask {
	return &PeriodicTask{
		ticker:  ticker,
		mailbox: utils.NewMailbox(1),
		taskFn:  taskFn,
	}
}

func (task *PeriodicTask) OnContextStarted(ctx Context, parent Context) error {
	task.ticker.Start()

	task.Context.Go("ticker", func(ctx Context) {
		for {
			select {
			case <-ctx.Done():
				return

			case <-task.ticker.Notify():
				task.Enqueue()

			case <-task.mailbox.Notify():
				x := task.mailbox.Retrieve()
				if x != nil {
					task.taskFn(ctx)
				}
			}
		}
	})
	return nil
}

func (task *PeriodicTask) Close() error {
	task.ticker.Close()
	return nil
}

func (task *PeriodicTask) Enqueue() {
	task.mailbox.Deliver(struct{}{})
}
