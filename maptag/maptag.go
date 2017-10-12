package maptag

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
)

type pluginConfig struct {
	maptype  string
	cmd      string
	reftype  string
	refname  string
	refgroup string
	regex    string
	replace  string
	args     []string
	ttl      time.Duration
}

type Plugin struct {
	initialized bool
	cachetime   time.Time
	mapping     map[string][]string
	config      *pluginConfig
	re          *regexp.Regexp
}

func NewPlugin() *Plugin {
	mp := Plugin{
		mapping:     make(map[string][]string),
		initialized: false,
		cachetime:   time.Unix(0, 0),
	}
	return &mp
}

func (p *Plugin) Process(mts []plugin.Metric, cfg plugin.Config) ([]plugin.Metric, error) {
	var err error

	if !p.initialized {
		p.config, err = getConfig(cfg)
		if err != nil {
			return mts, err
		}
		p.initialized = true
	}

	p.re, err = regexp.Compile(p.config.regex)
	if err != nil {
		return nil, err
	}

	// cycle all metrics
	for _, mt := range mts {
		switch p.config.maptype {
		case "newtag":

			p.getMapping()
			idx := -1
			switch p.config.reftype {

			// looking for tag value
			case "tag":
				// if metric has a tag "refname"
				if m_tagval, ok := mt.Tags[p.config.refname]; ok {
					// lookup of tag value in mapping "refgroup"
					idx = getValueIndex(p.mapping[p.config.refgroup], m_tagval)
				}

				// looking for namespace element name
			case "ns_name":
				// cycle metric namespace elements
				for _, nse := range mt.Namespace {
					// if namespace element name is "refname"
					if nse.Name == p.config.refname {
						// lookup of namespace element value in mapping "refgroup"
						idx = getValueIndex(p.mapping[p.config.refgroup], nse.Value)
					}
				}

				// looking for namespace element value
			case "ns_value":
				// Lookup of "refname" in metric namespace elements values
				nsi := getValueIndex(mt.Namespace.Strings(), p.config.refname)
				// if such namespace element is found
				if nsi >= 0 {
					// lookup of namespace element value in mapping "refgroup"
					idx = getValueIndex(p.mapping[p.config.refgroup], mt.Namespace[nsi].Value)
				}

				// Incorrect reftype value
			default:
				return nil, fmt.Errorf("Incorrect reftype value: %v", p.config.reftype)
			}

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
		case "replacetag":
			if m_tagval, ok := mt.Tags[p.config.refname]; ok {
				// lookup of tag value in mapping "refgroup"
				new_tagval := p.re.ReplaceAllString(m_tagval, p.config.replace)
				mt.Tags[p.config.refname] = new_tagval
			}

		default:
			return nil, fmt.Errorf("Incorrect maptype value: %v", p.config.maptype)
		}
	}

	return mts, nil
}

func (p *Plugin) getMapping() error {
	if time.Since(p.cachetime) > p.config.ttl {
		output, err := getCmdStdout(p.config.cmd, p.config.args)
		if err != nil {
			return err
		}

		p.mapping, err = getMappings(output, p.re)
		if err != nil {
			return err
		}
		p.cachetime = time.Now()
	}
	return nil
}

func (p *Plugin) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{"*"}, "maptype", true)
	policy.AddNewStringRule([]string{"*"}, "cmd", false)
	policy.AddNewStringRule([]string{"*"}, "reftype", false)
	policy.AddNewStringRule([]string{"*"}, "refname", true)
	policy.AddNewStringRule([]string{"*"}, "refgroup", false)
	policy.AddNewStringRule([]string{"*"}, "regex", true)
	policy.AddNewStringRule([]string{"*"}, "replace", false)
	policy.AddNewIntRule([]string{"*"}, "ttl", false, plugin.SetDefaultInt(180))
	return *policy, nil
}

func getConfig(cfg plugin.Config) (*pluginConfig, error) {
	var err error
	errs := []error{}
	mpc := pluginConfig{}
	mpc.maptype, err = cfg.GetString("maptype")
	if err != nil {
		errs = append(errs, fmt.Errorf(err.Error()+" maptype"))
	}

	if mpc.maptype == "newtag" {
		mpc.cmd, err = cfg.GetString("cmd")
		if err != nil {
			errs = append(errs, fmt.Errorf(err.Error()+" cmd"))
		}

		mpc.reftype, err = cfg.GetString("reftype")
		if err != nil {
			errs = append(errs, fmt.Errorf(err.Error()+" reftype"))
		}

		mpc.refgroup, err = cfg.GetString("refgroup")
		if err != nil {
			errs = append(errs, fmt.Errorf(err.Error()+" refgroup"))
		}
	}

	if mpc.maptype == "replacetag" {
		mpc.replace, err = cfg.GetString("replace")
		if err != nil {
			errs = append(errs, fmt.Errorf(err.Error()+" replace"))
		}
	}

	mpc.refname, err = cfg.GetString("refname")
	if err != nil {
		errs = append(errs, fmt.Errorf(err.Error()+" refname"))
	}

	mpc.regex, err = cfg.GetString("regex")
	if err != nil {
		errs = append(errs, fmt.Errorf(err.Error()+" regex"))
	}

	ittl, err := cfg.GetInt("ttl")
	if err != nil {
		errs = append(errs, fmt.Errorf(err.Error()+" ttl"))
	}
	mpc.ttl = time.Duration(ittl) * time.Minute

	mpc.args, err = getConfigArgs(cfg)
	if err != nil {
		errs = append(errs, err)
	}

	errorstr := ""
	var errout error
	if len(errs) > 0 {
		for _, err := range errs {
			errorstr += err.Error() + "\n"
		}
		errout = fmt.Errorf(errorstr)
	}

	return &mpc, errout
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
		if (v == val) || (v == ("/dev/" + val)) {
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
