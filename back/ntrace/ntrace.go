package ntrace

import (
	"github.com/oneclickvirt/nt3/nt"
)

func TraceRoute3(language, location, checkType string) {
	nt.TraceRoute(language, location, checkType)
}
