package go_cron

import (
	"context"
	"cron/alerter"
	"cron/config"
	"cron/handler"
	logger "cron/logger"
	"fmt"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"log/slog"
	"runtime"
	"time"
)

type JobHandler func(context.Context) error

type jobDetail struct {
	cronSpec string
	handler  JobHandler
	jobName  string
	maxRetry int
}

type CronServer struct {
	cron       *cron.Cron
	cronLogger cron.Logger
	//alerter    alerter.Alerter
	jobDetails []jobDetail
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewCronServer() *CronServer {
	cLogger := logger.NewCronLogger(slog.Default())

	c := cron.New(
		cron.WithLocation(time.Local),
		cron.WithLogger(cLogger),
		cron.WithChain(
			cron.Recover(cLogger),
		),
	)
	svc := handler.NewHandler()

	jobDetails := []jobDetail{
		{"@every 1s", svc.Expire1Request, "任务1过时", 0},
		{"@every 1s", svc.Expire2Response, "任务2过时", 0},
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &CronServer{
		cron:       c,
		cronLogger: cLogger,
		jobDetails: jobDetails,
		//alerter:    alerter,
		ctx:    ctx,
		cancel: cancel,
	}
}
func (cs *CronServer) Start() error {
	slog.Info("cron server start")
	for _, job := range cs.jobDetails {
		if err := cs.addWrappedJob(job.cronSpec, job.handler, job.jobName, job.maxRetry); err != nil {
			slog.Error("register cron job failed", "jobName", job.jobName, "error", err)
			return err
		}
	}
	slog.Info("starting cron server...")
	cs.cron.Run()
	return nil
}
func (cs *CronServer) Stop() error {
	slog.Error("stopping cron server...")
	cs.cancel()

	stopCtx := cs.cron.Stop()
	<-stopCtx.Done()
	slog.Error("cron server stopped")
	return nil
}

func (cs *CronServer) addWrappedJob(cronSpec string, handler JobHandler, jobName string, maxRetry int) error {
	job := cron.NewChain(
		cron.DelayIfStillRunning(cs.cronLogger),
	).Then(cron.FuncJob(func() {
		ctx := logger.AddAttrsToCtx(cs.ctx, "jobName", jobName)
		ctx = logger.AddAttrsToCtx(ctx, "jobID", uuid.NewString())

		// recover
		defer func() {
			if r := recover(); r != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				slog.ErrorContext(ctx, fmt.Sprintf("cron job panic: %v, stack: %v", err, string(buf)))

				// 上报错误信息
				//reportData := map[string]any{
				//	"ts":           time.Now().Format(time.RFC3339),
				//	"server":       "cron server",
				//	"cron.JobName": jobName,
				//	"msg":          "cron job panic",
				//	"error":        err.Error(),
				//}
				//go cs.alerter.Alert(reportData)

			}
		}()

		// run job with retry
		retried := 0
		for {
			// 关闭server，不再启动或重试
			select {
			case <-ctx.Done():
				return
			default:
			}

			err := handler(ctx)
			if err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("run job %v failed, maxRetry: %v, retried: %v, err: %v", jobName, maxRetry, retried, err))
				// 上报错误信息
				//reportData := map[string]any{
				//	"ts":           time.Now().Format(time.RFC3339),
				//	"server":       "cron server",
				//	"cron.JobName": jobName,
				//	"msg":          "run cron job failed",
				//	"error":        err.Error(),
				//}
				//go cs.alerter.Alert(reportData)

				if retried < maxRetry {
					retried++
					ctx = logger.AddAttrsToCtx(ctx, "jobID", uuid.NewString())
					continue
				}
			}

			return
		}
	}))
	_, err := cs.cron.AddJob(cronSpec, job)
	if err != nil {
		return err
	}
	return nil
}

func NewAlerter(c *config.Config) alerter.Alerter {
	if c.Alert.Type == "feishu" {
		return alerter.NewFeishuAlerter(c.Alert.Feishu.WebhookURL, c.Alert.Feishu.SignSecret)
	}
	return alerter.NewNoopAlerter()
}
