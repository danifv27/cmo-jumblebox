package pipeline

import (
	"fmt"
	"net/url"

	apipe "fry.org/qft/jumble/internal/application/pipeline"
	"fry.org/qft/jumble/internal/infrastructure/pipeline/splunk"
	"github.com/speijnik/go-errortree"
)

// Parse the uri string and returns the proper implementation
// Available uris:
func Parse[S apipe.Messager](URI string) (apipe.Piper[S], error) {
	var err, rcerror error
	var u *url.URL

	u, err = url.Parse(URI)
	if err != nil {
		rcerror = errortree.Add(rcerror, "pipeline.Parse", err)
		return nil, rcerror
	}
	if u.Scheme != "pipeline" {
		rcerror = errortree.Add(rcerror, "pipeline.Parse", fmt.Errorf("invalid scheme %s", URI))
		return nil, rcerror
	}
	switch u.Opaque {
	case "splunk":
		return splunk.NewSplunkPipe[S](), nil
	}

	return nil, errortree.Add(rcerror, "pipeline.Parse", fmt.Errorf("unsupported pipeline implementation %q", u.Opaque))
}
