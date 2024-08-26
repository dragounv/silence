package silence

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"math"
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
		MaxIterations:  20,
		MaxLines:       64_000,
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

	err = job.runCrawls(app)
	if err != nil {
		return
	}

	return nil
}

func (job *Job) initCrawls(app *App) error {
	if job.MaxIterations < 1 {
		err := fmt.Errorf("MaxIterations is set to less than one")
		app.Log.Error(
			"MaxIterations must be bigger than 0",
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

	if job.MaxLines < 1 {
		err := fmt.Errorf("MaxLines is set to less than one")
		app.Log.Error(
			"MaxLines must be bigger than 0",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}
	iterations := int(math.Ceil(float64(lines) / float64(job.MaxLines)))
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
	const timestampFormat = "20060102150405"
	timestamp := time.Now().Format(timestampFormat)
	for i := 0; i < cap(crawls); i++ {
		crawls = append(crawls, NewCrawl(i, timestamp, SeedsDirectory))
	}

	app.Log.Debug("", slog.Int("lines", lines), slog.Int("iterations", len(crawls)))

	err = job.createSeedFiles(crawls)
	if err != nil {
		app.Log.Error(
			"failed to create seed files for individual harvests",
			slog.String(ErrorKey, err.Error()),
		)
	}

	job.crawls = crawls

	app.Log.Debug("Crawls initialized")
	return nil
}

// Counts lines in file. It may ignore last line if it doesn't end with newline
// but that is not important, becouse the information is only used to
// determine number of harvests.
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

func (job *Job) createSeedFiles(crawls []*Crawl) error {
	linesPerFile := job.MaxLines
	seedsOrigin, err := os.Open(job.SeedsPath)
	if err != nil {
		return err
	}
	defer seedsOrigin.Close()

	seedsScanner := bufio.NewScanner(seedsOrigin)

	for _, crawl := range crawls {
		seedsBatch, err := os.Create(crawl.SeedsFile)
		if err != nil {
			err = fmt.Errorf("failed to create file %s with error: %w", crawl.SeedsFile, err)
			return err
		}
		defer seedsBatch.Close()

		// Only owner and group can write
		err = seedsBatch.Chmod(0664)
		if err != nil {
			err = fmt.Errorf("failed to change permissions to file %s with error: %w", crawl.SeedsFile, err)
			return err
		}

		err = copyLines(seedsScanner, seedsBatch, linesPerFile)
		if err != nil {
			err = fmt.Errorf("failed to copyLines from %s to %s with error: %w", job.SeedsPath, crawl.SeedsFile, err)
			return err
		}
	}

	return nil
}

func copyLines(linesIn *bufio.Scanner, linesOut io.Writer, linesPerFile int) error {
	writer := bufio.NewWriter(linesOut)
	i := 0
	// Scan until linesPerFile, scanner hits EOF or scanner encouters error.
	for i < linesPerFile && linesIn.Scan() {
		line := linesIn.Bytes()
		_, err := writer.Write(line)
		if err != nil {
			return err
		}
		err = writer.WriteByte('\n')
		if err != nil {
			return err
		}
		i++
	}
	if linesIn.Err() != nil {
		return linesIn.Err()
	}
	return writer.Flush()
}

func (job *Job) runCrawls(app *App) error {
	for _, crawl := range job.crawls {
		app.Log.Info(
			fmt.Sprintf("starting crawl %d", crawl.ID),
		)
		err := crawl.Run(job, app)
		if err != nil {
			app.Log.Error(
				fmt.Sprintf("error when processing crawl %d", crawl.ID),
				slog.String(ErrorKey, err.Error()),
				slog.Int("id", crawl.ID),
			)
		}
	}

	
}