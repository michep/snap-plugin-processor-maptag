package maptag

import (
	"testing"
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	. "github.com/smartystreets/goconvey/convey"
	"runtime"
)

func TestNewPlugin(t *testing.T) {
	Convey("Creating new plugin", t, func() {
		plg := NewPlugin()
		So(plg, ShouldNotBeNil)
		So(plg.mapping, ShouldNotBeNil)
	})
}

func TestPlugin_GetConfigPolicy(t *testing.T) {
	plg := NewPlugin()
	Convey("Getting config policy", t, func() {
		So(func() { plg.GetConfigPolicy() }, ShouldNotPanic)

		config, err := plg.GetConfigPolicy()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
	})
}

func TestPlugin_Process(t *testing.T) {
	plg := NewPlugin()

	Convey("Processing metrics with reftype=tag", t, func() {
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_tag())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldContainKey, "newtag")
		So(mts[0].Tags["newtag"], ShouldEqual, "somevalue")
		So(mts[1].Tags, ShouldNotContainKey, "newtag")
	})

	Convey("Processing metrics with reftype=ns_value", t, func() {
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_value())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldContainKey, "newtag")
		So(mts[0].Tags["newtag"], ShouldEqual, "somevalue")
		So(mts[1].Tags, ShouldContainKey, "newtag")
		So(mts[1].Tags["newtag"], ShouldEqual, "somevalue")
	})

	Convey("Processing metrics with reftype=ns_name", t, func() {
		mts, err := plg.Process(mockMetrics(), mockPluginConfig_name())
		So(err, ShouldBeNil)
		So(mts, ShouldHaveLength, 2)
		So(mts[0].Tags, ShouldNotContainKey, "newtag")
		So(mts[1].Tags, ShouldContainKey, "newtag")
		So(mts[1].Tags["newtag"], ShouldEqual, "somevalue")
	})
}

func mockMetrics() []plugin.Metric {
	metrics := []plugin.Metric{}
	// first
	mt := plugin.Metric{}
	mt.Data = 3.1415926
	mt.Timestamp = time.Now()
	mt.Namespace = plugin.NewNamespace("test", "static", "namespace", "pi")
	mt.Tags = map[string]string{"tagone": "valueone", "tagtwo": "valuetwo"}
	metrics = append(metrics, mt)

	// second
	mt = plugin.Metric{}
	mt.Data = 2.7182818
	mt.Timestamp = time.Now()
	mt.Namespace = plugin.NewNamespace("test")
	mt.Namespace = mt.Namespace.AddDynamicElement("dynamic", "dynamic namespace element")
	mt.Namespace = mt.Namespace.AddStaticElements("namespace", "e")
	mt.Namespace[1].Value = "valuedynamic"
	mt.Tags = map[string]string{"tagone": "anothervalueone", "tagtwo": "valuetwo"}
	metrics = append(metrics, mt)

	return metrics
}

func mockPluginConfig_tag() plugin.Config {
	cfg := plugin.Config{
		"cmd":      "?",
		"arg0":     "?",
		"arg1":     "echo valueone somevalue",
		"regex":    "(?P<first>\\S+)\\s+(?P<newtag>\\S+)",
		"reftype":  "tag",
		"refname":  "tagone",
		"refgroup": "first",
	}
	addCmdToConfig(&cfg)
	return cfg
}

func mockPluginConfig_value() plugin.Config {
	cfg := &plugin.Config{
		"cmd":      "?",
		"arg0":     "?",
		"arg1":     "echo namespace somevalue",
		"regex":    "(?P<first>\\S+)\\s+(?P<newtag>\\S+)",
		"reftype":  "ns_value",
		"refname":  "namespace",
		"refgroup": "first",
	}
	addCmdToConfig(cfg)
	return *cfg
}

func mockPluginConfig_name() plugin.Config {
	cfg := plugin.Config{
		"cmd":      "?",
		"arg0":     "?",
		"arg1":     "echo valuedynamic somevalue",
		"regex":    "(?P<first>\\S+)\\s+(?P<newtag>\\S+)",
		"reftype":  "ns_name",
		"refname":  "dynamic",
		"refgroup": "first",
	}
	addCmdToConfig(&cfg)
	return cfg
}

func addCmdToConfig(cfg *plugin.Config) {
	switch runtime.GOOS {
	case "windows":
		(*cfg)["cmd"] = "cmd.exe"
		(*cfg)["arg0"] = "/C"
	case "linux":
		(*cfg)["cmd"] = "/bin/sh"
		(*cfg)["arg0"] = "-c"
	}
}
