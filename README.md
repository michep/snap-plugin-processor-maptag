# Snap processor plugin - maptag
Snap plugin intended to add tags based on lookup

It's used in the [Snap framework](http://github.com:intelsdi-x/snap).

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license-and-authors)
6. [Acknowledgements](#acknowledgements)

## Getting Started
### System Requirements 
* [golang 1.7+](https://golang.org/dl/) (needed only for building)

### Operating systems
All OSs currently supported by snap:
* Linux/amd64
* Darwin/amd64

### Installation
#### To build the plugin binary:
Fork https://github.com/michep/snap-plugin-processor-maptag

Clone repo

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-processor-maptag
```

Build the plugin by running make within the cloned repo:
```
$ go build .
```
This builds the plugin.

### Configuration and Usage
* Set up the [Snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)

## Documentation

The intention of this plugin is to add tags to metrics based on regexp lookup over particular command or script execution output.

The plugin can be configured by following parameters:
- `cmd` - command to execute. Required parameter.
- `arg[0..9]` - arguments to command. Optional parameters.
- `regex` - regular expression to process command output. Use golang [regexp syntax](https://github.com/google/re2/wiki/Syntax). This expression should contain named capturing groups for lookup and tags adding. Required parameter.
- `refgroup` - named capturing group in `regex` that will be used to lookup values. Required parameter.
- `reftype` - which data use to lookup - tag values, metric static namespace element value or dynamic namespace element name; proper values are `tag`, `ns_value` and `ns_name`. Required parameter.
- `refname` - metric `reftype`, which value will be searched in `refgroup` results. Required parameter.
- `ttl` - plugin cache time-to-live, in minutes. After cache was created it will not be updated till this time period expire. Optional parameter, default value - 180m (3h).

Notice: Special characters in regular expressions needs to be escaped.


### Examples
In this example we run iostat collector, maptag processor and file publisher to write data into file.

Documentation for Snap collector cpu plugin can be found [here](https://github.com/intelsdi-x/snap-plugin-collector-iostat).
Documentation for Snap file publisher plugin can be found [here](https://github.com/intelsdi-x/snap-plugin-publisher-file).

In one terminal window, open the snap daemon with log level 1 (`-l 1`) and disabled plugin signing check (`-t 0`):
```
$ snapteld -t 0 -l 1
```

In another terminal window:

Download and load collector, processor and publisher plugins
```
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-collector-iostat/latest/linux/x86_64/snap-plugin-collector-iostat
$ wget http://snap.ci.snap-telemetry.io/plugins/snap-plugin-publisher-file/latest/linux/x86_64/snap-plugin-publisher-file
$ wget https://github.com/michep/snap-plugin-processor-maptag/releases/download/1/snap-plugin-processor-maptag_linux_x86_64

$ chmod 755 snap-plugin-*
$ snaptel plugin load snap-plugin-collector-iostat
$ snaptel plugin load snap-plugin-publisher-file
$ snaptel plugin load snap-plugin-processor-maptag_linux_x86_64
```

See available metrics for your system
```
$ snaptel metric list
```

Create a task file - see examplary task manifests in [examples/tasks](examples/tasks/):

```yaml

---
version: 1
schedule:
  type: "simple"
  interval: "10s"
workflow:
  collect:
    metrics:
      /intel/iostat/device/*: {}
    process:
    - plugin_name: maptag
      config:
        cmd: /bin/sh
        arg0: -c
        arg1: ls -l /dev/disk/by-uuid/ 
        regex: '(?P<uuid>\w{8}-\w+-\w+-\w+-\w+) -> (\.\.\/)+(?P<re_dev>\S+)'
        refgroup: re_dev
        reftype: tag
        refname: dev
        ttl: 20
    publish:
          publish:
            - plugin_name: "file"
              config:
                file: "/tmp/maptag-processog.log"
```

Start task:
```
$ snaptel task create -t task.yaml
```

This data is published to a file `/tmp/maptag-processog.log`

To stop task:
```
$ snaptel task stop <task id>
```

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an issue and/or submit a pull request.

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements
* Author: [Mike Chepaykin](https://github.com/michep/)
