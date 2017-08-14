package filesystem

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/maja42/AniScraper/utils"

	"github.com/PuerkitoBio/goquery"
)

var log utils.Logger // TODO: delete this variable

// WebClient is used as a helper utility for fetching webcontent
type WebClient interface {
	GetDocument(pageUrl *url.URL) (*goquery.Document, error)
}

type webClient struct {
	client *http.Client
}

func NewWebClient() WebClient {
	return &webClient{
		client: &http.Client{},
	}
}

func (webcli *webClient) GetDocument(pageUrl *url.URL) (*goquery.Document, error) {
	request := &http.Request{
		Method: "GET",
		URL:    pageUrl,
	}
	webcli.addRequestHeaders(request)
	log.Debugf("Fetching %q", pageUrl.String())

	resp, err := webcli.client.Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Server returned with status %d", resp.Status)
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	return doc, err
}

func (webcli *webClient) addRequestHeaders(request *http.Request) {
	/*
	   In order to use the Web-API, the following headers need to be correct:
	       Content-Type
	       X-Requested-With
	       User-Agent
	       Cookie
	    Additionally, Post-Requests might have parameters (like the parameter 'v') that have to be correct.

	    If one or more of these parameters are incorrect, the server either responds without a body, or with
	        "External access is forbidden! If you’re interested please submit a request for an API interface."

	    Until the API access has been granted, all requests will use the Web-API and HTML parsing,
	    which does not require any special permissions. However, this interface is subject to change and
	    will require permanent adaption of the application.
	*/

	// request.Header.Add("Accept", "application/json, text/javascript, */*; q=0.01")
	// request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// request.Header.Add("X-Requested-With", "XMLHttpRequest")

	// request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.49 Safari/537.36")
	// request.Header.Set("Cookie", "session_database=4ab8e138a63ef1eb648004b8a0ad9805e9fdfe22%7E577a6e1b6ec7b4-23295555")
}
