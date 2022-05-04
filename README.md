# Dogstatsd Local
> A local implementation of the dogstatsd protocol (<=v1.2) from [Datadog](https://www.datadog.com)

## Why?

[Datadog](https://www.datadog.com) is great for production application metric aggregation. This project was inspired by the need to inspect and debug metrics _before_ sending them to `datadog`.

`dogstatsd-local` is a small program which understands the `dogstatsd` and `statsd` protocols. It listens on a local UDP server and writes metrics, events and service checks per the [dogstatsd protocol](https://docs.datadoghq.com/developers/dogstatsd/datagram_shell/) to `stdout` in user configurable formats.

This can be helpful for _debugging_ metrics themselves, and to prevent polluting datadog with noisy metrics from a development environment. **dogstatsd-local** can also be used to pipe metrics as json to other processes for further processing.

## Usage

### Build Manually

This is a go application with no external dependencies. Building should be as simple as running `go build` in the source directory.

Once compiled, the `dogstatsd-local` binary can be run directly:
```bash
$ ./dogstatsd-local -port 8126
```

### Docker

```bash
$ docker run -p 8125:8125/udp anujdas/dogstatsd-local
```

## Sample Formats

### Raw (no formatting)

When writing a metric such as:

```bash
$ printf "namespace.metric:1:2|c|@1|#tag1,tag2:value" | nc -cu localhost 8125
```

Running **dogstatsd-local** with the `-format raw` flag will output the plain udp packet:

```bash
$ docker run -p 8125:8125/udp anujdas/dogstatsd-local -format raw
namespace.metric:1:2|c|@1|#tag1,tag2:value

```

### Human

When writing a metric such as:

```bash
$ printf "namespace.metric:1:2|c|@1|#tag1,tag2:value" | nc -cu localhost 8125
```

Running **dogstatsd-local** with the `-format human` flag will output a human readable metric:

```bash
$ docker run -p 8125:8125/udp anujdas/dogstatsd-local -format human
metric:counter|namespace.metric|1.00,2.00  tag1 tag2:value

```

### JSON

When writing a metric such as:
```bash
$ printf "namespace.metric:1:2|c|@1|#tag1,tag2:value|c:c1" | nc -cu localhost 8125
```

Running **dogstatsd-local** with the `-format json` flag will output json:

```bash
$ docker run -p 8125:8125/udp anujdas/dogstatsd-local -format json
{"name":"namespace.metric","type":"counter","values":[1,2],"sample_rate":1,"tags":["tag1","tag2:value","container_id":"c1"]}
```

**dogstatsd-local** can be piped to any process that understands json via stdin. For example, to pretty print the name and first value with [jq](https://stedolan.github.io/jq/):

```bash
$ docker run -p 8125:8125/udp anujdas/dogstatsd-local -format json | jq ".name,.values[0]"
"namespace.metric"
1
```
