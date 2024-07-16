package commediatests

import (
	"fmt"
	"sync"

	"github.com/oneclickvirt/CommonMediaTests/commediatests/netflix"
	"github.com/oneclickvirt/CommonMediaTests/commediatests/website"
	. "github.com/oneclickvirt/defaultset"
)

func MediaTests(language string) string {
	var (
		res0, res1, res2, result string
		err0, err1, err2         error
		wg                       sync.WaitGroup
	)
	switch language {
	case "en":
		wg.Add(3)
		func() {
			defer wg.Done()
			res1, err1 = website.YoutubeCheck("en")
			if err1 != nil && EnableLoger {
				InitLogger()
				defer Logger.Sync()
				Logger.Info(err1.Error())
			}
		}()
		func() {
			defer wg.Done()
			res0, err0 = netflix.Netflix("en")
			if err0 != nil && EnableLoger {
				InitLogger()
				defer Logger.Sync()
				Logger.Info(err0.Error())
			}
		}()
		func() {
			defer wg.Done()
			res2, err2 = website.Disneyplus("en")
			if err2 != nil && EnableLoger {
				InitLogger()
				defer Logger.Sync()
				Logger.Info(err2.Error())
			}
		}()
	case "zh":
		wg.Add(3)
		func() {
			defer wg.Done()
			res1, err1 = website.YoutubeCheck("zh")
			if err1 != nil && EnableLoger {
				InitLogger()
				defer Logger.Sync()
				Logger.Info(err1.Error())
			}
		}()
		func() {
			defer wg.Done()
			res0, err0 = netflix.Netflix("zh")
			if err0 != nil && EnableLoger {
				InitLogger()
				defer Logger.Sync()
				Logger.Info(err0.Error())
			}
		}()
		func() {
			defer wg.Done()
			res2, err2 = website.Disneyplus("zh")
			if err2 != nil && EnableLoger {
				InitLogger()
				defer Logger.Sync()
				Logger.Info(err2.Error())
			}
		}()
	default:
		fmt.Println("不支持的语言参数")
		return ""
	}
	wg.Wait()
	result += White("----------------Netflix-----------------\n")
	result += res0
	result += White("----------------Youtube-----------------\n")
	result += res1
	result += White("---------------DisneyPlus---------------\n")
	result += res2
	return result
}
