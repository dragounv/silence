package silence

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	AcceptHeaderKey = "Accept"
	AcceptXML = "application/xml"
)

const ActionKey = "action"

type Crawl struct {
	ID int
	SeedsFile string
	Job *Job
}

func NewCrawl(id int, timestamp string, directory string, job *Job) *Crawl {
	seedsFile := fmt.Sprintf("seeds-%s-%03d.txt", timestamp, id)
	seedsFile = path.Join(directory, seedsFile)
	return &Crawl{id, seedsFile, job}
}

func (crawl *Crawl) String() string {
	return fmt.Sprintf("id:%d seeds:%s", crawl.ID, crawl.SeedsFile)
}

func (crawl *Crawl) Run(app *App) error {
	app.Log.Debug(
		fmt.Sprintf("crawl %d is running", crawl.ID),
	)
	
	// err := crawl.pingHeritrix(app)
	// if err != nil {
	// 	return err
	// }

	err := crawl.createCrawlBeans()
	if err != nil {
		app.Log.Error(
			fmt.Sprintf("failed to create %s", CrawlerBeansName),
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	err = crawl.build()

	err = crawl.launch()

	err = crawl.unpause()

	// TODO: Add monitoring

	err = crawl.terminate()

	err = crawl.teardown()

	// ---

	app.Log.Debug(
		fmt.Sprintf("cleaning crawl %d", crawl.ID),
	)
	err = crawl.clean()

	return nil
}

func (crawl *Crawl) pingHeritrix(app *App) error {
	const endpoint = "/engine"
	response, err := crawl.request(http.MethodGet, endpoint, nil)
	if err != nil {
		app.Log.Error(
			"error when pinging heritrix",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}
	defer response.Body.Close()
	
	if response.StatusCode != 200 {
		err = fmt.Errorf("error code recieved")
		app.Log.Error(
			"error code returned from heritrix",
			slog.Int(ReturnStatusKey, response.StatusCode),
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	app.Log.Debug(
		"ping from heritrix",
		slog.Int(ReturnStatusKey, response.StatusCode),
	)

	// Possible addition of checking status of crawls
	return nil
}

func (crawl *Crawl) request(method string, path string, values url.Values) (*http.Response ,error) {
	address, err := url.Parse(crawl.Job.CrawlerAddress)
	if err != nil {
		return nil, err
	}
	address.Scheme = "https://"
	address.Path = path

	var body io.Reader
	if method == http.MethodGet || values == nil {
		body = http.NoBody
	} else {
		body = strings.NewReader(values.Encode())
	} 

	request, err := http.NewRequest(method, address.String(), body)
	if err != nil {
		return nil, err
	}

	request.Header.Set(AcceptHeaderKey, AcceptXML)
	response, err := crawl.Job.client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (crawl *Crawl) createCrawlBeans() error {
	beansTemplate, err := template.ParseFiles(crawl.Job.TemplatePath)
	if err != nil {
		return err
	}

	// Add crawl specific values to config
	// It is importatnt to create new copy for every iteration
	config := crawl.Job.Config.Copy()
	config.seedsFile = crawl.SeedsFile
	config.id = crawl.ID

	crawlerBeansFile, err := os.Create(CrawlerBeansName)
	if err != nil {
		return err
	}
	defer crawlerBeansFile.Close()

	err = beansTemplate.Execute(crawlerBeansFile, config)
	if err != nil {
		return err
	}

	return nil
}

func (crawl *Crawl) clean() error {
	err := os.Remove(crawl.SeedsFile)
	if err != nil {
		return err
	}

	// err = os.Remove(CrawlerBeansName)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func (crawl *Crawl) build() error {
	const endpoint = "engine/job/Topics"
	const action = "build"
	data := url.Values{}
	data.Add(ActionKey, action)
	response, err := crawl.request(http.MethodPost, endpoint, data)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = fmt.Errorf("non ok status code recieved (%s)", response.Status)
		return err
	}

	
}

func (crawl *Crawl) launch() error {
	const endpoint = "engine/job/Topics"
	const action = "launch"

}

func (crawl *Crawl) unpause() error {
	const endpoint = "engine/job/Topics"
	const action = "unpause"

}

func (crawl *Crawl) terminate() error {
	const endpoint = "engine/job/Topics"
	const action = "terminate"
}

func (crawl *Crawl) teardown() error {
	const endpoint = "engine/job/Topics"
	const action = "teardown"
}