package commediatest

import (
	"fmt"
	"github.com/oneclickvirt/CommonMediaTests/commediatests"
)

func Media() {
	res := commediatests.MediaTests("zh")
	fmt.Printf(res)
}
