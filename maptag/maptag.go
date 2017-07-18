package maptag

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type pluginConfig struct {
	cmd      string
	reftag   string
	refgroup string
	regex    string
	args     []string
	ttl      time.Duration
}

type Plugin struct {
	initialized bool
	cachetime   time.Time
	mapping     map[string][]string
	config      *pluginConfig
}

func NewPlugin() *Plugin {
	mp := Plugin{
		mapping:     make(map[string][]string),
		initialized: false,
	}
	return &mp
}

func (p *Plugin) Process(mts []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	var err error

	p.config, err = getConfig(cfg)
	if err != nil {
		return mts, err
	}

	output, err := getCmdStdout(p.config.cmd, p.config.args)
	if err != nil {
		return mts, err
	}

	re, err := regexp.Compile(p.config.regex)
	if err != nil {
		return mts, err
	}

	p.mapping, err = getMappings(output, re)
	if err != nil {
		return mts, err
	}
	// cycle all metrics
	for _, mt := range mts {
		// if metric has a tag "reftag"
		if m_tagval, ok := mt.Tags[p.config.reftag]; ok {
			// lookup value of a tag "reftag" in mapping "refgroup"
			idx := getValueIndex(p.mapping[p.config.refgroup], m_tagval)
			// if found
			if idx >= 0 {
				// cycle all groups in mapping
				for grname, grvalues := range p.mapping {
					// if group is NOT a "refgroup"
					if grname != p.config.refgroup {
						// add new tag whith group name and value to metric
						mt.Tags[grname] = grvalues[idx]
					}
				}
			}
		}
	}

	return mts, nil
}

func (p *Plugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{""}, "cmd", true)
	policy.AddNewStringRule([]string{""}, "reftag", true)
	policy.AddNewStringRule([]string{""}, "refgroup", true)
	policy.AddNewStringRule([]string{""}, "regex", true)
	return *policy, nil
}

func getConfig(cfg plugin.Config) (*pluginConfig, error) {
	var err error
	mpc := pluginConfig{}
	mpc.cmd, _ = cfg.GetString("cmd")
	mpc.reftag, _ = cfg.GetString("reftag")
	mpc.refgroup, _ = cfg.GetString("refgroup")
	mpc.regex, _ = cfg.GetString("regex")

	mpc.args, err = getConfigArgs(cfg)
	if err != nil {
		return nil, err
	}

	return &mpc, nil
}

func getConfigArgs(cfg plugin.Config) ([]string, error) {
	args := []string{}
	for i := 0; i < 10; i++ {
		key := "arg" + strconv.Itoa(i)
		if val, err := cfg.GetString(key); err == nil {
			args = append(args, val)
		}
	}
	return args, nil
}

func getCmdStdout(cmd string, args []string) (string, error) {
	output_b, err := exec.Command(cmd, args...).Output()
	output := string(output_b)
	if err != nil {
		return "", err
	}
	return output, nil
}

func getValueIndex(arr []string, val string) int {
	for idx, v := range arr {
		if v == val {
			return idx
		}
	}
	return -1
}

func getMappings(output string, re *regexp.Regexp) (map[string][]string, error) {
	mapping := make(map[string][]string)
	groupids := make(map[string]int)
	for idx, name := range re.SubexpNames() {
		if name != "" {
			groupids[name] = idx
		}
	}
	for _, line := range strings.Split(output, "\n") {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}
		for grname, grid := range groupids {
			if _, ok := mapping[grname]; !ok {
				mapping[grname] = []string{}
			}
			mapping[grname] = append(mapping[grname], matches[grid])
		}
	}
	return mapping, nil
}
