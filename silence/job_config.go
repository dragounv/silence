package silence

import (
	"fmt"
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
	id        int
}

func (jc *JobConfig) CrawlName() string {
	const crawlType = "Topics"
	const delimiter = "-"
	timestamp := time.Now().Format(time.DateOnly)
	id := fmt.Sprintf("Part%d", jc.id)
	return strings.Join([]string{crawlType, timestamp, jc.CrawlNameSuffix, id}, delimiter)
}

func (jc *JobConfig) SeedsFile() string {
	return jc.seedsFile
}

func (jc *JobConfig) Copy() *JobConfig {
	newStruct := *jc
	return &newStruct
}
