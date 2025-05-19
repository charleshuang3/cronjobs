package scheduler

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/charleshuang3/cronjobs/internal/bot"
	"github.com/charleshuang3/cronjobs/internal/config"
	"github.com/robfig/cron/v3"
)

// Scheduler manages and runs cron jobs
type Scheduler struct {
	cron        *cron.Cron
	bot         bot.Bot
	jobs        map[string]jobEntry
	jobsLock    sync.RWMutex
	running     map[string]time.Time // Map to track running jobs and their start times
	runningLock sync.RWMutex
}

type jobEntry struct {
	config   config.JobConfig
	cronID   cron.EntryID
	schedule string
}

// New creates a new Scheduler instance
func New(b bot.Bot) *Scheduler {
	return &Scheduler{
		cron:    cron.New(cron.WithSeconds()),
		bot:     b,
		jobs:    make(map[string]jobEntry),
		running: make(map[string]time.Time),
	}
}

// SetBot sets the bot instance for the scheduler
func (s *Scheduler) SetBot(b bot.Bot) {
	s.bot = b
}

// AddJobs adds all jobs from the configuration
func (s *Scheduler) AddJobs(jobs []config.JobConfig) error {
	if s.bot == nil {
		return fmt.Errorf("bot must be set before adding jobs")
	}

	for _, job := range jobs {
		if err := s.AddJob(job); err != nil {
			return fmt.Errorf("failed to add job %s: %w", job.Name, err)
		}
	}
	return nil
}

// AddJob adds a single job to the scheduler
func (s *Scheduler) AddJob(job config.JobConfig) error {
	if s.bot == nil {
		return fmt.Errorf("bot must be set before adding jobs")
	}

	s.jobsLock.Lock()
	defer s.jobsLock.Unlock()

	if _, exists := s.jobs[job.Name]; exists {
		return fmt.Errorf("job with name %s already exists", job.Name)
	}

	cronID, err := s.cron.AddFunc(job.Cron, func() {
		if s.isJobRunning(job.Name) {
			log.Printf("Job %s is already running, skipping scheduled execution.", job.Name)
			return
		}
		s.executeJobAndSendNotification(job)
	})
	if err != nil {
		return err
	}

	s.jobs[job.Name] = jobEntry{
		config:   job,
		cronID:   cronID,
		schedule: job.Cron,
	}

	return nil
}

// executeJob runs a single job and handles its output
func (s *Scheduler) executeJob(job config.JobConfig) (time.Duration, string, error) {
	s.runningLock.Lock()
	s.running[job.Name] = time.Now()
	s.runningLock.Unlock()

	startTime := time.Now()
	cmd := exec.Command(job.Program, job.Args...)

	output, err := cmd.CombinedOutput()

	duration := time.Since(startTime)

	s.runningLock.Lock()
	delete(s.running, job.Name)
	s.runningLock.Unlock()

	return duration, string(output), err
}

func (s *Scheduler) executeJobAndSendNotification(job config.JobConfig) {
	duration, output, err := s.executeJob(job)

	result := bot.JobExecuteResult{
		JobName:       job.Name,
		Duration:      duration,
		Output:        output,
		Error:         err,
		CaptureOutput: job.CaptureOutput,
	}

	if err := s.bot.SendJobExecuteResult(result); err != nil {
		log.Printf("%v", err)
	}
}

// Start begins running the scheduled jobs
func (s *Scheduler) Start() {
	s.cron.Start()
}

// Stop stops all scheduled jobs
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// ListJobs returns information about all scheduled jobs
func (s *Scheduler) ListJobs() []bot.JobInfo {
	s.jobsLock.RLock()
	defer s.jobsLock.RUnlock()

	jobs := make([]bot.JobInfo, 0, len(s.jobs))
	for name, job := range s.jobs {
		jobs = append(jobs, bot.JobInfo{
			Name:          name,
			Program:       job.config.Program,
			Args:          job.config.Args,
			Cron:          job.schedule,
			CaptureOutput: job.config.CaptureOutput,
		})
	}
	return jobs
}

// RunJobByName runs a job immediately by its name
func (s *Scheduler) RunJobByName(name string) error {
	if s.bot == nil {
		return fmt.Errorf("bot must be set before running jobs")
	}

	s.jobsLock.RLock()
	defer s.jobsLock.RUnlock()

	if job, exists := s.jobs[name]; exists {
		if s.isJobRunning(name) {
			return fmt.Errorf("job with name %s is already running", name)
		}
		go func() {
			s.executeJobAndSendNotification(job.config)
		}()
		return nil
	}
	return fmt.Errorf("job with name %s not found", name)
}

func (s *Scheduler) isJobRunning(name string) bool {
	s.runningLock.RLock()
	defer s.runningLock.RUnlock()
	_, ok := s.running[name]
	return ok
}

// RunningJobs returns a map of currently running jobs with their durations
func (s *Scheduler) RunningJobs() map[string]time.Duration {
	s.runningLock.RLock()
	defer s.runningLock.RUnlock()

	runningJobs := make(map[string]time.Duration)
	for jobName, startTime := range s.running {
		runningJobs[jobName] = time.Since(startTime)
	}
	return runningJobs
}
