package silence

import (
	"fmt"
	"path"
)

type Crawl struct {
	ID int
	SeedsFile string
}

func NewCrawl(id int, timestamp string, directory string) *Crawl {
	seedsFile := fmt.Sprintf("seeds-%s-%03d.txt", timestamp, id)
	seedsFile = path.Join(directory, seedsFile)
	return &Crawl{id, seedsFile}
}

func (crawl *Crawl) String() string {
	return fmt.Sprintf("id:%d seeds:%s", crawl.ID, crawl.SeedsFile)
}

func (crawl *Crawl) Run(job *Job, app *App) error {
	
}