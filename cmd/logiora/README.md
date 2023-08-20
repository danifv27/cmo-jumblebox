# Logiora

## Overview

This Prometheus exporter is designed to extract valuable service level indicator (SLI) metrics related to latency and availability from nginx logs. By tailing log files and sorting requests into API-specific buckets, this exporter calculates SLIs for each API, providing insights into performance and reliability.

The name is constructed to capture the essence of log analysis by combining elements that relate to the activity of examining logs for valuable information. "Logiora" emphasizes the act of exploration.

* "Logi": Derived from "logs" or "log," which directly refers to the records or data entries that are typically analyzed for insights.
* "ora": This part of the name is reminiscent of "explora" or "exploration." It suggests the act of delving into and exploring the logs for valuable information and understanding.

# Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://gitea.casamau.synology.me/fry/cmo-jumbleBox.git). 

# Acknowledgement

* [logrus](https://github.com/sirupsen/logrus)
* [pipelines](https://github.com/splunk/pipelines)