package main

import (
	"fmt"
	"strings"

	"github.com/gocolly/colly"
)

func parseIssuePage(repoOwner string, repoName string, issueNum string) (*IssuePageInfos, error) {
	title := ""
	ownerID := ""
	repoID := ""
	hasTitleChanged := false
	isClosedAsComplete := false

	c := colly.NewCollector(
		colly.AllowedDomains("github.com"),
	)
	url := fmt.Sprintf("https://github.com/%s/%s/issues/%s", repoOwner, repoName, issueNum)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		fmt.Printf("Response %s: %d bytes\n", r.Request.URL, len(r.Body))
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error %s: %v\n", r.Request.URL, err)
	})

	// extract text
	c.OnHTML("bdi.js-issue-title", func(e *colly.HTMLElement) {
		title = strings.TrimSpace(e.Text)
	})

	// extract owner ID
	c.OnHTML("meta[name='octolytics-dimension-user_id']", func(e *colly.HTMLElement) {
		for _, attribute := range e.DOM.Nodes[0].Attr {
			if attribute.Key == "content" {
				ownerID = attribute.Val
				break
			}
		}
	})

	// extract repo ID
	c.OnHTML("meta[name='octolytics-dimension-repository_id']", func(e *colly.HTMLElement) {
		for _, attribute := range e.DOM.Nodes[0].Attr {
			if attribute.Key == "content" {
				repoID = attribute.Val
				break
			}
		}
	})

	// check if the title has been changed
	c.OnHTML("div.js-discussion", func(e1 *colly.HTMLElement) {
		e1.ForEach("div.js-timeline-item", func(_ int, e2 *colly.HTMLElement) {
			e2.ForEach("[class='TimelineItem js-targetable-element']", func(_ int, e3 *colly.HTMLElement) {
				if strings.Contains(e3.Text, "changed the title") {
					hasTitleChanged = true
				}
			})
		})
	})

	// check if the issue has been closed as complete
	c.OnHTML("div.gh-header-meta > div > span", func(e *colly.HTMLElement) {
		for _, attribute := range e.DOM.Nodes[0].Attr {
			if attribute.Key == "title" && attribute.Val == "Status: Closed" {
				isClosedAsComplete = true
				break
			}
		}
	})

	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	return &IssuePageInfos{title, repoOwner, repoName, ownerID, repoID, hasTitleChanged, isClosedAsComplete}, nil
}
