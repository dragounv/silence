package silence

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"text/template"
	"time"
)

const (
	AcceptHeaderKey = "Accept"
	AcceptXML       = "application/xml"
)

const ActionKey = "action"

type Crawl struct {
	ID        int
	SeedsFile string
	Job       *Job
	endpoint  string
}

func NewCrawl(id int, timestamp string, directory string, job *Job) *Crawl {
	seedsFile := fmt.Sprintf("seeds-%s-%03d.txt", timestamp, id)
	seedsFile = path.Join(directory, seedsFile)
	endpoint := "engine/job/" + job.Name
	return &Crawl{id, seedsFile, job, endpoint}
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

	err = crawl.createCrawlBeans()
	if err != nil {
		app.Log.Error(
			fmt.Sprintf("failed to create %s", CrawlerBeansName),
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	err = crawl.build()
	if err != nil {
		app.Log.Error(
			"build failed",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	err = crawl.launch()
	if err != nil {
		app.Log.Error(
			"launch failed",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	err = crawl.unpause()
	if err != nil {
		app.Log.Error(
			"unpause failed",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	// TODO: Add monitoring
	err = crawl.await(app)
	if err != nil {
		app.Log.Error(
			"await failed",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	err = crawl.terminate()
	if err != nil {
		app.Log.Error(
			"terminate failed",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	err = crawl.teardown()
	if err != nil {
		app.Log.Error(
			"teardown failed",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

	// ---
	// TODO: Wait for crawl to teardown
	err = crawl.awaitTeardown()
	if err != nil {
		app.Log.Error(
			"error when waiting for teardown to finish",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}
	
	app.Log.Debug(
		fmt.Sprintf("cleaning crawl %d", crawl.ID),
	)
	err = crawl.clean()
	if err != nil {
		app.Log.Error(
			"clean failed",
			slog.String(ErrorKey, err.Error()),
		)
		return err
	}

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

func (crawl *Crawl) request(method string, path string, values url.Values) (*http.Response, error) {
	// net/url is really bad, I should just use string manipulation instead
	addr := crawl.Job.CrawlerAddress
	if !strings.HasPrefix(addr, "http") {
		addr = "https://" + addr
	}

	// slog.Info(addr)
	address, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	address.Path = path
	// slog.Info(address.String(), "path", path)

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
	config.crawlType = crawl.Job.Name

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

func (crawl *Crawl) doAction(action string) error {
	data := url.Values{}
	data.Add(ActionKey, action)
	response, err := crawl.request(http.MethodPost, crawl.endpoint, data)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		err = fmt.Errorf("non ok status code recieved (%s)", response.Status)
		return err
	}

	// body, err := io.ReadAll(response.Body)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println(string(body))

	// --

	time.Sleep(5 * time.Second)
	return nil
}

func (crawl *Crawl) build() error {
	const action = "build"
	return crawl.doAction(action)
}

func (crawl *Crawl) launch() error {
	const action = "launch"
	return crawl.doAction(action)
}

func (crawl *Crawl) unpause() error {
	const action = "unpause"
	return crawl.doAction(action)
}

func (crawl *Crawl) terminate() error {
	const action = "terminate"
	return crawl.doAction(action)
}

func (crawl *Crawl) teardown() error {
	const action = "teardown"
	return crawl.doAction(action)
}

func (crawl *Crawl) await(app *App) error {
	app.Log.Info(
		"waiting for crawl to finish",
		slog.Int("max_wait_s", crawl.Job.MaxWaitSeconds),
	)

	maxDuration := time.Duration(crawl.Job.MaxWaitSeconds) * time.Second
	done := time.After(maxDuration)

	for {
		response, err := crawl.request(http.MethodGet, crawl.endpoint, nil)
		if err != nil {
			app.Log.Error(
				"error when checking crawl status",
				slog.String(ErrorKey, err.Error()),
			)
			return err
		}

		if response.StatusCode != 200 {
			err = fmt.Errorf("response returned code %s", response.Status)
			app.Log.Error(
				"error when checking crawl status",
				slog.String(ErrorKey, err.Error()),
				slog.Int(ReturnStatusKey, response.StatusCode),
			)
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			app.Log.Error(
				"error when checking crawl status",
				slog.String(ErrorKey, err.Error()),
			)
			return err
		}

		crawlResponse := new(CrawlResponse)

		err = xml.Unmarshal(body, crawlResponse)
		if err != nil {
			app.Log.Error(
				"error when checking crawl status",
				slog.String(ErrorKey, err.Error()),
			)
			return err
		}

		app.Log.Info(
			"crawl status",
			slog.String("state", crawlResponse.ControllerState),
			slog.String("exit_status", crawlResponse.ExitStatus),
			slog.String("exit_desc", crawlResponse.ExitDescription),
		)

		if crawlResponse.ControllerState == "FINISHED" {
			app.Log.Info("finished, terminating")
			return nil
		}

		response.Body.Close()

		select {
		case <-done:
			{
				app.Log.Warn("crawl did not finish before timeout, terminating")
				return nil
			}
		default:
			// Do not block here
		}

		time.Sleep(1 * time.Minute)
	}
}

func (crawl *Crawl) awaitTeardown() error {
	const timeout = 2 * time.Hour
	done := time.After(timeout)
	for {
		response, err := crawl.request(http.MethodGet, crawl.endpoint, nil)
		if err != nil {
			return err
		}

		if response.StatusCode != 200 {
			err = fmt.Errorf("response returned code %s", response.Status)
			return err
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		crawlResponse := new(CrawlResponse)

		err = xml.Unmarshal(body, crawlResponse)
		if err != nil {
			return err
		}

		if crawlResponse.IsRunning == "false" &&
		crawlResponse.IsLaunchable == "true" {
			return nil
		}

		select {
		case <-done:
			{
				err = fmt.Errorf("the crawl was not tear down before timeout, it is likely frozen and must be shut down by operator")
				return err
			}
		default:
			// Do not block here
		}
	}
}

type CrawlResponse struct {
	XMLName         xml.Name `xml:"job"`
	ControllerState string   `xml:"crawlControllerState"`
	ExitStatus      string   `xml:"crawlExitStatus"`
	ExitDescription string   `xml:"statusDescription"`
	Actions         []string `xml:"availableActions>value"`
	IsRunning       string   `xml:"isRunning"`
	IsLaunchable    string   `xml:"isLaunchable"`
}
