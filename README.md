# BunnyCDN Exporter for Prometheus

This is a simple server that scrapes BunnyCDN stats and exports them via HTTP for
Prometheus consumption.

## Getting Started

To run it:

```bash
./bunnycdn_exporter [flags]
```

Help on flags:

```bash
./bunnycdn_exporter --help
```

For more information check the [source code documentation][gdocs]. All of the
core developers are accessible via the Prometheus Developers [mailinglist][].

[gdocs]: http://godoc.org/github.com/permutive/bunnycdn_exporter
[mailinglist]: https://groups.google.com/forum/?fromgroups#!forum/prometheus-developers

## Usage

### Direct

```bash
bunnycdn_exporter --bunnycdn.api-key="<API_KEY>"
```

Or by using an environment variable for setting the API key:

```bash
export BUNNYCDN_API_KEY="<API_KEY>"
bunnycdn_exporter"
```

### Docker

[![Docker Pulls](https://img.shields.io/docker/pulls/permutive/bunnycdn-exporter.svg?maxAge=604800)][hub]

To run the bunnycdn exporter as a Docker container, run:

```bash
docker run -p 9584:9584 permutive/bunnycdn-exporter --bunnycdn.api-key="<API_KEY>"
```

alternatively, the API key can be passed as an environment variable:
```bash
docker run -p 9584:9584 -e BUNNYCDN_API_KEY="<API_KEY>" permutive/bunnycdn-exporter"
```

[hub]: https://hub.docker.com/r/permutive/bunnycdn-exporter/

## Development

[![Go Report Card](https://goreportcard.com/badge/github.com/prometheus/bunnycdn_exporter)][goreportcard]
[![Code Climate](https://codeclimate.com/github/prometheus/bunnycdn_exporter/badges/gpa.svg)][codeclimate]

[goreportcard]: https://goreportcard.com/report/github.com/permutive/bunnycdn_exporter
[codeclimate]: https://codeclimate.com/github/permutive/bunnycdn_exporter

### Building

```bash
make build
```

### Testing

[![Build Status](https://travis-ci.org/permutive/bunnycdn_exporter.png?branch=master)][travisci]

```bash
make test
```

[travisci]: https://travis-ci.org/prometheus/bunnycdn_exporter

## License

Apache License 2.0, see [LICENSE](https://github.com/prometheus/bunnycdn_exporter/blob/master/LICENSE).
