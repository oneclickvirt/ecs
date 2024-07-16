package jp

import (
	"fmt"
	"github.com/oneclickvirt/UnlockTests/model"
	"github.com/oneclickvirt/UnlockTests/utils"
	"net/http"
	"strings"
)

// VideoMarket
// www.videomarket.jp 仅 ipv4 且 post 请求
func VideoMarket(c *http.Client) model.Result {
	name := "VideoMarket"
	hostname := "videomarket.jp"
	if c == nil {
		return model.Result{Name: name}
	}
	headers := map[string]string{
		"authority":     "www.videomarket.jp",
		"authorization": "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ4MTkxNTkyMDIsImlhdCI6MTY2NTU1OTIwMiwiaXNzIjoiaHR0cHM6Ly9hdXRoLnZpZGVvbWFya2V0LmpwIiwic3ViIjoiY2ZjZDIwODQ5NWQ1NjVlZjY2ZTdkZmY5Zjk4NzY0ZGEiLCJ1c2VyX3R5cGUiOjAsInNpdGVfdHlwZSI6MiwiY2xpZW50X2lkIjoiYmVkNDdkOTFiMDVhYjgzMGM4YzBhYmFiYjQwNTg5MTFhY2E5NTdkMDBkMTUzNjA2MjI3NzNhOTQ0Y2RlNzRhNSIsInZtaWQiOjB9.Tq18RCxpVz1oV1lja52uRmF0nT6Oa0QsZMTVlPfANwb-RrcSn7PwE9vh7GdNIBc0ydDxRoUMuhStz_Kbu8KxvAh25eafFh7hf0DDqWKKU4ayPMueaR12t74SjFIRC7Cla1NR4uRn3_mgJfZFqOkIf6L5OR9LOVIBhrQPkhbMyqwZyh_kxTH7ToJIQoINb036ftqcF1KfR8ndtBlkrrWWnDpfkmE7-fJQHh92oKKd9l98W5awuEQo0MFspIdSNgt3gLi9t1RRKPDISGlzJkwMLPkHIUlWWZaAmnEkwSeZCPj_WJaqUqBATYKhi3yJZNGlHsScQ_KgAopxlsI6-c88Gps8i6yHvPVYw3hQ9XYq9gVL_SpyW9dKKSPE9MY6I19JHLBXuFXi5OJccqtQzTnKm_ZQM3EcKt5s0cNlXm9RMt0fNdRTQdJ53noD9o-b6hUIxDcHScJ_-30Emiv-55g5Sq9t5KPWO6o0Ggokkj42zin69MxCiUSHXk5FgeY8rX76yGBeLPLPIaaRPXEC1Jeo1VO56xNnQpyX_WHqHWDKhmOh1qSzaxiAiC5POMsTfwGr19TwXHUldYXxuNMIfeAaPZmNTzR5J6XdenFkLnrssVzXdThdlqHpfguLFvHnXTCAm0ZhFIJmacMNw1IxGmCQfkM4HtgKB9ZnWm6P0jIISdg",
		"cookie":        "_gid=GA1.2.1853799793.1706147718; VM_REGIST_BANNER_REF_LINK=%2Ftitle%2F292072%2FA292072001999H01; __ulfpc=202401250957239984; _im_vid=01HMZ5C5GNNC6VWSPKD3E4W7YP; __td_signed=true; _td_global=0d11678b-5151-473e-b3a8-4f4d780f26a6; __juicer_sesid_9i3nsdfP_=d36a2e17-011",
	}
	url := "https://www.videomarket.jp/graphql"
	resp, body, err := utils.PostJson(c, url, `{"operationName": "repPacksOnTab","variables": {"fullTitleId": "292072","groupType": "SINGLE_CHOICE","page": {"current": 1,"size": 20}},"query": "query repPacksOnTab($fullTitleId: String!, $groupType: GroupType!, $page: PageInput!) {\n  repPacksOnTab(fullTitleId: $fullTitleId, groupType: $groupType, page: $page) {\n    repFullPackId\n    groupType\n    packName\n    fullTitleId\n    titleName\n    storyImageUrl16x9\n    playTime\n    subtitleDubType\n    outlines\n    courseIds\n    price\n    discountRate\n    couponPrice\n    couponDiscountRate\n    rentalDays\n    viewDays\n    deliveryExpiredAt\n    salesType\n    counter {\n      currentPage\n      currentResult\n      totalPages\n      totalResults\n      __typename\n    }\n    undiscountedPrice\n    packs {\n      undiscountedPrice\n      canPurchase\n      fullPackId\n      subGroupType\n      fullTitleId\n      qualityConsentType\n      courseIds\n      price\n      discountRate\n      couponPrice\n      couponDiscountRate\n      rentalDays\n      viewDays\n      deliveryExpiredAt\n      salesType\n      extId\n      stories {\n        fullStoryId\n        subtitleDubType\n        encodeVersion\n        isDownloadable\n        isBonusMaterial\n        fileSize\n        __typename\n      }\n      __typename\n    }\n    status {\n      hasBeenPlayed\n      isCourseRegistered\n      isEstPurchased\n      isNowPlaying\n      isPlayable\n      isRented\n      playExpiredAt\n      playableQualityType\n      rentalExpiredAt\n      __typename\n    }\n    __typename\n  }\n}\n"}`,
		headers,
	)
	if err != nil {
		return model.Result{Name: name, Status: model.StatusNetworkErr, Err: err}
	}
	defer resp.Body.Close()
	if strings.Contains(body, "OverseasAccess") {
		return model.Result{Name: name, Status: model.StatusNo}
	}
	if strings.Contains(body, "292072") {
		result1, result2, result3 := utils.CheckDNS(hostname)
		unlockType := utils.GetUnlockType(result1, result2, result3)
		return model.Result{Name: name, Status: model.StatusYes, UnlockType: unlockType}
	}
	return model.Result{Name: name, Status: model.StatusUnexpected,
		Err: fmt.Errorf("get www.videomarket.jp failed with code: %d", resp.StatusCode)}
}
