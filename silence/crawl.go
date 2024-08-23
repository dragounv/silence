package silence

import "fmt"

type Crawl struct {
	ID int
	SeedsFile string
}

func NewCrawl(id int, timestamp string) *Crawl {
	seedsFile := fmt.Sprintf("seeds-%s-%03d.txt",timestamp, id)
	return &Crawl{id, seedsFile}
}