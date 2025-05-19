package scheduler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charleshuang3/cronjobs/internal/bot/bottesting"
	"github.com/charleshuang3/cronjobs/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestExecuteJobSuccess(t *testing.T) {
	s := New(nil)

	duration, output, err := s.executeJob(config.JobConfig{
		Program:       "echo",
		Args:          []string{"hello"},
		CaptureOutput: true, // explicitly capture output for this test
	})

	assert.Equal(t, "hello\n", output)
	assert.Nil(t, err)
	assert.LessOrEqual(t, duration, 1*time.Second)
}

func TestExecuteJobFailed(t *testing.T) {
	tempDir := t.TempDir()
	defer os.RemoveAll(tempDir)

	s := New(nil)

	duration, output, err := s.executeJob(config.JobConfig{
		Program:       "rm",
		Args:          []string{filepath.Join(tempDir, "nonexistent-file")},
		CaptureOutput: true, // explicitly capture output for this test
	})

	assert.True(t, strings.HasPrefix(output, "rm: cannot remove"))
	assert.NotNil(t, err)
	assert.LessOrEqual(t, duration, 1*time.Second)
}

func TestExecuteJobNoCapture_NotEffect(t *testing.T) {
	s := New(nil)

	duration, output, err := s.executeJob(config.JobConfig{
		Program:       "echo",
		Args:          []string{"hello"},
		CaptureOutput: false, // explicitly avoid capturing output
	})

	assert.Equal(t, "hello\n", output) // Output should be empty as capture is disabled
	assert.Nil(t, err)
	assert.LessOrEqual(t, duration, 1*time.Second)
}

func TestRunJobByName(t *testing.T) {
	stubBot := &bottesting.StubBot{}
	s := New(stubBot)
	job := config.JobConfig{
		Name:          "test_job",
		Program:       "echo",
		Args:          []string{"hello"},
		Cron:          "@monthly",
		CaptureOutput: true,
	}

	err := s.AddJob(job)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		err := s.RunJobByName("test_job")
		assert.NoError(t, err)
	})

	t.Run("running", func(t *testing.T) {
		s.runningLock.Lock()
		s.running["test_job"] = time.Now()
		s.runningLock.Unlock()
		err := s.RunJobByName("test_job")
		assert.ErrorContains(t, err, "is already running")
	})

	t.Run("not found", func(t *testing.T) {
		err := s.RunJobByName("nonexistent_job")
		assert.ErrorContains(t, err, "not found")
	})
}

func TestExecuteJobAndSendNotification(t *testing.T) {
	stubBot := &bottesting.StubBot{}
	s := New(stubBot)

	job := config.JobConfig{
		Name:    "test_job",
		Program: "echo",
		Args:    []string{"hello"},
	}

	s.executeJobAndSendNotification(job)

	assert.NoError(t, stubBot.JobExecuteResult.Error)
	assert.Equal(t, "hello\n", stubBot.JobExecuteResult.Output)
}
