package powclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

type Challenge struct {
	RequestID string `json:"request_id"`
	Challenge string `json:"challenge"`
}

type RequestResponse struct {
	Challenge   Challenge `json:"challenge"`
	RequestTime int64     `json:"request_time"`
}

type SubmitRequest struct {
	Challenge   Challenge `json:"challenge"`
	Answer      []string  `json:"answer"`
	RequestTime int64     `json:"request_time"`
}

type SubmitResponse struct {
	Token string `json:"token"`
}

type GetTokenParams struct {
	TimeoutSec  time.Duration
	BaseUrl     string
	RequestPath string
	SubmitPath  string
	UserAgent   string
	SNI         string
	Host        string
	Proxy       *url.URL /** 支持socks5:// http:// **/
}

func NewGetTokenParams() *GetTokenParams {
	return &GetTokenParams{
		TimeoutSec:  5 * time.Second, // 你的默认值
		BaseUrl:     "http://127.0.0.1:55000",
		RequestPath: "/request_challenge",
		SubmitPath:  "/submit_answer",
		UserAgent:   "POW client",
		SNI:         "",
		Host:        "",
		Proxy:       nil,
	}
}

type ChallengeParams struct {
	BaseUrl     string
	RequestPath string
	SubmitPath  string
	UserAgent   string
	Host        string
	Client      *http.Client
}

func RetToken(getTokenParams *GetTokenParams) (string, error) {
	client := &http.Client{
		Timeout: getTokenParams.TimeoutSec,
	}
	if getTokenParams.SNI != "" {
		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					ServerName: getTokenParams.SNI,
				},
			},
		}
	}
	if getTokenParams.Proxy != nil {
		if client.Transport == nil {
			client.Transport = &http.Transport{}
		}
		client.Transport.(*http.Transport).Proxy = http.ProxyURL(getTokenParams.Proxy)
	}
	challengeParams := &ChallengeParams{
		BaseUrl:     getTokenParams.BaseUrl,
		RequestPath: getTokenParams.RequestPath,
		SubmitPath:  getTokenParams.SubmitPath,
		UserAgent:   getTokenParams.UserAgent,
		Host:        getTokenParams.Host,
		Client:      client,
	}
	// Get challenge
	challengeResponse, err := requestChallenge(challengeParams)
	if err != nil {
		return "", err
	}
	//fmt.Println(challengeResponse.Challenge.Challenge)

	// Solve challenge and submit answer
	token, err := submitAnswer(challengeParams, challengeResponse)
	if err != nil {
		return "", err
	}

	return token, nil
}

func requestChallenge(challengeParams *ChallengeParams) (*RequestResponse, error) {
	req, err := http.NewRequest("GET", challengeParams.BaseUrl+challengeParams.RequestPath, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", challengeParams.UserAgent)
	//req.Header.Add("Host", getTokenParams.Host)
	req.Host = challengeParams.Host
	resp, err := challengeParams.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			log.Fatalln("请求次数超限，请稍后再试")
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var challengeResponse RequestResponse
	err = json.NewDecoder(resp.Body).Decode(&challengeResponse)
	if err != nil {
		return nil, err
	}

	return &challengeResponse, nil
}

func submitAnswer(challengeParams *ChallengeParams, challengeResponse *RequestResponse) (string, error) {
	requestTime := challengeResponse.RequestTime
	challenge := challengeResponse.Challenge.Challenge
	requestId := challengeResponse.Challenge.RequestID
	N := new(big.Int)
	N.SetString(challenge, 10)
	factorsList := factors(N)
	if len(factorsList) != 2 {
		return "", errors.New("factors function did not return exactly two factors")
	}
	p1 := factorsList[0]
	p2 := factorsList[1]
	if p1.Cmp(p2) > 0 { // if p1 > p2
		p1, p2 = p2, p1 // swap p1 and p2
	}
	submitRequest := SubmitRequest{
		Challenge:   Challenge{RequestID: requestId},
		Answer:      []string{p1.String(), p2.String()},
		RequestTime: requestTime,
	}
	requestBody, err := json.Marshal(submitRequest)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", challengeParams.BaseUrl+challengeParams.SubmitPath, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("User-Agent", challengeParams.UserAgent)
	//req.Header.Add("Host", getTokenParams.Host)
	req.Host = challengeParams.Host

	resp, err := challengeParams.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusTooManyRequests {
			return "", errors.New("请求次数超限")
		}
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", errors.New(string(bodyBytes))
	}

	var submitResponse SubmitResponse
	err = json.NewDecoder(resp.Body).Decode(&submitResponse)
	if err != nil {
		return "", err
	}

	return submitResponse.Token, nil
}
