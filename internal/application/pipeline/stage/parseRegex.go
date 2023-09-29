package stage

import (
	"fmt"
	"regexp"
	"strings"

	isplunk "fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
)

// RegexParse is a log record parser. Use specific constructors to initialize it.
type RegexParse struct {
	format string
	regexp *regexp.Regexp
}

// Log line by line parsing.
// exp  will be transformed to regular expression and used for parsing
// For example:
//
//	`$remote_addr [$time_local] "$request"`
//
// It should contain variables in form `$name`. The regular expression will be created using this string format representation
//
//	`^(?P<remote_addr>[^ ]+) \[(?P<time_local>[^]]+)\] "(?P<request>[^"]+)"$`
func NewRegexParse(exp string) *RegexParse {
	// First split up multiple concatenated fields with placeholder
	placeholder := " _PLACEHOLDER___ "
	preparedFormat := exp
	concatenatedRe := regexp.MustCompile(`[A-Za-z0-9_]\$[A-Za-z0-9_]`)
	for concatenatedRe.MatchString(preparedFormat) {
		preparedFormat = regexp.MustCompile(`([A-Za-z0-9_])\$([A-Za-z0-9_]+)(\\?([^$\\A-Za-z0-9_]))`).ReplaceAllString(preparedFormat, fmt.Sprintf("${1}${3}%s$$${2}${3}", placeholder))
		// fmt.Printf("[DBG]preparedFormat: ^%v\n", strings.Trim(preparedFormat, " "))
	}
	// Second replace each fileds to regexp grouping
	quotedFormat := regexp.QuoteMeta(preparedFormat + " ")
	re := regexp.MustCompile(`\\\$([A-Za-z0-9_]+)(?:\\\$[A-Za-z0-9_])*(\\?([^$\\A-Za-z0-9_]))`).ReplaceAllString(quotedFormat, "(?P<$1>[^$3]*)$2")
	// Finally remove placeholder
	re = regexp.MustCompile(fmt.Sprintf(".%s", placeholder)).ReplaceAllString(re, "")

	// fmt.Printf("[DBG]regexp: ^%v\n", strings.Trim(re, " "))
	return &RegexParse{
		format: exp,
		regexp: regexp.MustCompile(fmt.Sprintf("^%v", strings.Trim(re, " "))),
	}
}

func (p *RegexParse) Do(input isplunk.SplunkPipeMsg) []isplunk.SplunkPipeMsg {
	var outMsgs []isplunk.SplunkPipeMsg

	if val, ok := input.Get("entry").(string); ok {
		re := p.regexp
		fields := re.FindStringSubmatch(val)
		if len(fields) > 0 {
			// fmt.Printf("[DBG]regexp fields: %v", fields)
			outMsg := isplunk.NewSplunkMessage("parsed.fields", nil)
			m := make(map[string]string)
			for i, name := range re.SubexpNames() {
				if i == 0 {
					continue
				}
				m[name] = fields[i]
			}
			outMsg.Add("fields", m)

			return append(outMsgs, outMsg)
		}
	}

	return outMsgs
}
