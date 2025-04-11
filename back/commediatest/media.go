package commediatest

import (
	"fmt"
	"github.com/oneclickvirt/CommonMediaTests/commediatests"
)

func ComMediaTest(language string) {
	res := commediatests.MediaTests(language)
	fmt.Printf(res)
}
