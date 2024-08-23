package silence

import (
	"strings"
	"time"
)

// Configuration that will be used in crawl-beans template
type JobConfig struct {
	Operator        string
	Description     string
	DataLimit       int
	TimeLimit       int
	DedupDir        string
	ToeThreads      int
	MaxHops         int
	CrawlNameSuffix string

	seedsFile string
}

func (jc *JobConfig) CrawlName() string {
	const crawlType = "Topics"
	const delimiter = "-"
	timestamp := time.Now().Format(time.DateOnly)
	return strings.Join([]string{crawlType, timestamp, jc.CrawlNameSuffix}, delimiter)
}

func (jc *JobConfig) SeedsFile() string {
	return jc.seedsFile
}
