package nginx_log

import (
	"fmt"
	"os"
	"sync"

	"github.com/hpcloud/tail"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

type NginxLog struct {
	// Lock for preventing a data race during resource cleanup
	sourceFiles []string
	cleanup     sync.Mutex
	wg          sync.WaitGroup

	// track current tails so we can close them in Stop()
	tails []tail.Tail

	acc telegraf.Accumulator
}

const sampleConfig = `
  ## input files
  source_files = ["/var/log/nginx/nginx-error.log"]
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
	var wg sync.WaitGroup

	for file := range n.sourceFiles {
		if _, err := os.Start(file); os.IsNotExist(err) {
			return fmt.Errorf("file not exits '%s'", file)
		}

		wg.Add(1)
		t, err := tail.TailFile(file, tail.Config{Follow: true})
		t.tails = append(t.tails, t)

		go func(t *tail.Tail) {
			for line := range t.Lines {
				fmt.Println(line.Text)
			}
			defer wg.Done()
		}(t)
	}

	wg.Wait()

	return nil
}

// Start starts the tcp listener service.
func (n *NginxLog) Start(acc telegraf.Accumulator) error {

	t.acc = acc
	return n.Gather(acc)
}

// Stop cleans up all resources
func (n *NginxLog) Stop() {
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
