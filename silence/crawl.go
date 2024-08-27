package silence

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
)

const (
	AcceptHeaderKey = "Accept"
	AcceptXML = "application/xml"
)

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
	
	err := crawl.pingHeritrix(app)
	if err != nil {
		return err
	}

	return nil
}

func (crawl *Crawl) pingHeritrix(app *App) error {
	response, err := crawl.request(http.MethodGet, "/engine", http.NoBody)
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

func (crawl *Crawl) request(method string, path string, body io.Reader) (*http.Response ,error) {
	address, err := url.Parse(crawl.Job.CrawlerAddress)
	if err != nil {
		return nil, err
	}
	address.Scheme = "https://"
	address.Path = path

	request, err := http.NewRequest(method, address.String(), body)
	if err != nil {
		return nil, err
	}

	request.Header.Set(AcceptHeaderKey, AcceptXML)
	response, err := crawl.Job.client.Do(request)
	if err != nil {
		return nil, err
	}

	// defer response.Body.Close()
	// app.Log.Info(response.Status)
	// _, err = io.Copy(os.Stdout, response.Body)
	// if err != nil {
	// 	app.Log.Error(
	// 		"error when reading request body",
	// 		slog.String(ErrorKey, err.Error()),
	// 	)
	// 	return err
	// }

	return response, nil
}