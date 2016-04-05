package nginx_log

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type NginxLog struct {
	sync.Mutex
	// Lock for preventing a data race during resource cleanup
	Sources []string

	// track current tails so we can close them in Stop()
	tails []*tail.Tail

	acc telegraf.Accumulator
}

const sampleConfig = `
  ## input files
  sources = ["/var/log/nginx/nginx-error.log"]
`

func (n *NginxLog) SampleConfig() string {
	return sampleConfig
}

func (n *NginxLog) Description() string {
	return "Generic TCP listener"
}

// All the work is done in the Start() function, so this is just a dummy
// function.
func (n *NginxLog) Gather(acc telegraf.Accumulator) error {

	n.acc = acc

	seekinfo := tail.SeekInfo{Whence: os.SEEK_END}
	cfg := tail.Config{Follow: true, ReOpen: true, Logger: tail.DiscardingLogger, Location: &seekinfo}

	for _, file := range n.Sources {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("file not exits '%s'", file)
		}

		go func() {
			t, err := tail.TailFile(file, cfg)
			if err != nil {
				log.Fatalf("ERROR: tail file: %s - %s", file, err)
			}
			n.tails = append(n.tails, t)
			for line := range t.Lines {
				n.parse(line.Text, acc)
				//fmt.Println(line.Text)
			}
		}()
	}

	return nil
}

func (n *NginxLog) parse(logLine string, acc telegraf.Accumulator) {

	patternString := `(?P<host>[^\s]+)\s+(?P<remote_add>[^\s]+)\s+\[(?P<created_time>[^\s\]]+\s+\+\d+)\]`
	patternString += `\s+\"[A-Z]+\s+(?P<request_path>[^\s\?]+)[^\s]*\s+[^"]+"\s+(?P<status_code>\d+)`
	patternString += `\s+(?P<body_size>\d+)\s+"[^\s]+"\s+"[^"]+"\s+(?P<request_time>\d+\.\d+)`
	patternString += `\s+(?P<upstream_time>\d+\.\d+)?`
	myExp := regexp.MustCompile(patternString)
	matchGroups := myExp.FindStringSubmatch(logLine)
	tags := make(map[string]string)
	fields := make(map[string]interface{})

	if matchGroups != nil {
		tags["host"] = matchGroups[1]
		tags["path"] = matchGroups[4]

		toFloat := func(elem string) (float64, error) {
			if elem == "" || elem == "-" {
				return 0, nil
			}
			return strconv.ParseFloat(strings.TrimSpace(elem), 64)
		}
		bodySize, err1 := toFloat(matchGroups[6])
		responseTime, err2 := toFloat(matchGroups[7])
		if err1 != nil {
			log.Fatalf("Error: bodySize fail to int! from:%s, line:%s", matchGroups[6], logLine)
		}
		if err2 != nil {
			log.Fatalf("Error: responseTime fail to int! from:%s, line:%s", matchGroups[7], logLine)
		}
		fields["response_time"] = responseTime
		fields["body_size"] = bodySize
	}
	requestTime, err := time.Parse("02/Jan/2006:15:04:05 -0700", matchGroups[3])
	if err != nil {
		log.Fatalf("Error: requestTime parse fail! from:%s, line:%s", matchGroups[3], logLine)

	}

	acc.AddFields("nginx_log", fields, tags, requestTime)

}

// Start starts the tcp listener service.
func (n *NginxLog) Start(acc telegraf.Accumulator) error {
	n.Lock()
	defer n.Unlock()
	n.acc = acc
	return n.Gather(acc)
}

// Stop cleans up all resources
func (n *NginxLog) Stop() {
	n.Lock()
	defer n.Unlock()
	for _, t := range n.tails {
		t.Cleanup()
		t.Stop()
	}

}

func init() {
	inputs.Add("nginx_log", func() telegraf.Input {
		return &NginxLog{}
	})
}
