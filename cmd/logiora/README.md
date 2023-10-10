# Logiora - Log Parsing and IP Whitelisting Utility

Logiora is a Go-based command-line utility designed for parsing log files, extracting relevant information, and performing IP whitelisting checks. It provides a flexible and customizable approach to process logs and filter out IP addresses based on predefined criteria.

The name is constructed to capture the essence of log analysis by combining elements that relate to the activity of examining logs for valuable information. "Logiora" emphasizes the act of exploration.

* "Logi": Derived from "logs" or "log," which directly refers to the records or data entries that are typically analyzed for insights.
* "ora": This part of the name is reminiscent of "explora" or "exploration." It suggests the act of delving into and exploring the logs for valuable information and understanding.

## Features

*Log Parsing:* Logiora can parse log entries using custom log formats. You can specify the log format to be parsed as a command-line argument.

*IP Whitelisting:* Logiora allows you to check if IP addresses extracted from log entries match a whitelist of IP ranges provided in CIDR notation.

*Flexible Configuration:* You can configure Logiora's behavior using various command-line options, including specifying the log format, whitelist ranges, and output format.

*Output Formats:* Logiora supports multiple output formats, including JSON, Excel, and plain text, to suit your reporting and analysis needs.

*Pipeline Processing:* Logiora uses a pipeline-based approach to efficiently process log entries and perform IP whitelisting checks in parallel.

# Versioning

We use [SemVer](http://semver.org/) for versioning. For the versions available, see the [tags on this repository](https://gitea.casamau.synology.me/fry/cmo-jumbleBox.git). 

# Acknowledgement

* [logrus](https://github.com/sirupsen/logrus)
* [pipelines](https://github.com/splunk/pipelines)
