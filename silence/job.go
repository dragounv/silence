package silence

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Job struct {
	configPath string

	Name           string
	TemplatePath   string
	SeedsPath      string
	CrawlerAddress string
	MaxLines       int
	MaxIterations  int
	Config         *JobConfig

	client *http.Client
	crawls []*Crawl
}

const DefaultJobConfigPath = "job.json"

const SeedsDirectory = "seeds_dir"

func NewJob(app *App, path string) (*Job, error) {
	job := DefaultJob(path)

	data, err := os.ReadFile(job.configPath)
	if err != nil {
		app.Log.Error(
			"cannot read job file",
			slog.String(ErrorKey, err.Error()),
		)
		return job, err
	}

	err = json.Unmarshal(data, job)
	if err != nil {
		app.Log.Error(
			"failed to unmarshal job file",
			slog.String(ErrorKey, err.Error()),
		)
		return job, err
	}

	job.initClient()

	return job, nil
}

func DefaultJob(path string) *Job {
	return &Job{
		configPath:     path,
		SeedsPath:      "seeds.txt",
		TemplatePath:   "crawler-beans.template",
		CrawlerAddress: "localhost:7778",
		MaxIterations: 20,
		MaxLines: 64_000,
		Config:         new(JobConfig),
	}
}

func (job *Job) initClient() {
	// Heritrix uses self signed certificates, this is fine for now
	heritrixTransport := http.DefaultTransport.(*http.Transport).Clone()
	heritrixTransport.TLSClientConfig.InsecureSkipVerify = true
	job.client = &http.Client{Transport: heritrixTransport}
}

func (job *Job) run(app *App) error {
	err := job.initCrawls(app)
	if err != nil {
		return err
	}
}

func (job *Job) initCrawls(app *App) error {
	if job.MaxIterations < 1 {
		err := fmt.Errorf("error: max_iterations set to less than one")
		app.Log.Error(
			"maax_iterations must be bigger than 0",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	lines, err := countLines(job.SeedsPath)
	if err != nil {
		app.Log.Error(
			fmt.Sprintf("failed to get count of lines in file %s", job.SeedsPath),
			slog.String(ErrorKey, err.Error()),
			slog.Int("lines", lines),
		)
		return err
	}

	iterations := lines / job.MaxLines
	if iterations > job.MaxIterations {
		err := fmt.Errorf("too many iterations needed")
		app.Log.Error(
			fmt.Sprintf("number of iterations (%d) is bigger than max_iterations (%d)", iterations, job.MaxIterations),
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	// Create seeds directory
	err = os.Mkdir(SeedsDirectory, 0755)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		app.Log.Error(
			fmt.Sprintf("failed to create directory %s", SeedsDirectory),
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	crawls := make([]*Crawl, 0, iterations)
	timestamp := time.Now().Format("20060102150405")
	for i := 0; i < cap(crawls); i++ {
		crawls = append(crawls, NewCrawl(i, timestamp))
	}

	// TODO: Test that I can reach this part.
}

// Counts lines in file. It may ignore last line if it doesn't end with newline
// but that is not important for our usecase.
func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	const bufferSize = 1 << 20
	buffer := make([]byte, bufferSize)
	sum := 0
	newline := []byte{'\n'}
	for {
		readBytes, err := file.Read(buffer)
		if err == io.EOF {
			buffer = buffer[:readBytes]
			sum += bytes.Count(buffer, newline)
			break
		}
		if err != nil {
			return sum, err
		}
		sum += bytes.Count(buffer, newline)
	}

	return sum, nil
}
