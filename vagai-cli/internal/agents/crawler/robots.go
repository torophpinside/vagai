package crawler

import (
	"bufio"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type RobotsRule struct {
	Allow    []string
	Disallow []string
	CrawlDelay *time.Duration
}

type RobotsParser struct {
	rules    map[string]*RobotsRule
	userAgents []string
}

func NewRobotsParser(userAgents []string) *RobotsParser {
	return &RobotsParser{
		rules:      make(map[string]*RobotsRule),
		userAgents: userAgents,
	}
}

func (rp *RobotsParser) FetchAndParse(robotsURL string, client *http.Client) error {
	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", rp.userAgents[0])

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil
	}

	scanner := bufio.NewScanner(resp.Body)
	var currentUA string
	var inTargetUA bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "user-agent":
			currentUA = strings.ToLower(value)
			inTargetUA = rp.matchesUserAgent(currentUA)
		case "disallow":
			if inTargetUA && value != "" {
				rp.ensureRule(currentUA).Disallow = append(rp.ensureRule(currentUA).Disallow, value)
			}
		case "allow":
			if inTargetUA && value != "" {
				rp.ensureRule(currentUA).Allow = append(rp.ensureRule(currentUA).Allow, value)
			}
		case "crawl-delay":
			if inTargetUA {
				if delay, err := time.ParseDuration(value + "s"); err == nil {
					rp.ensureRule(currentUA).CrawlDelay = &delay
				}
			}
		}
	}

	return scanner.Err()
}

func (rp *RobotsParser) matchesUserAgent(ua string) bool {
	ua = strings.ToLower(ua)
	if ua == "*" {
		return true
	}
	for _, agent := range rp.userAgents {
		agentLower := strings.ToLower(agent)
		if strings.Contains(agentLower, ua) || strings.Contains(ua, agentLower) {
			return true
		}
	}
	return false
}

func (rp *RobotsParser) ensureRule(ua string) *RobotsRule {
	if rp.rules[ua] == nil {
		rp.rules[ua] = &RobotsRule{}
	}
	return rp.rules[ua]
}

func (rp *RobotsParser) IsAllowed(targetURL string) bool {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return true
	}
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	for _, rule := range rp.rules {
		for _, disallow := range rule.Disallow {
			if rp.matchPath(path, disallow) {
				for _, allow := range rule.Allow {
					if rp.matchPath(path, allow) {
						return true
					}
				}
				return false
			}
		}
	}
	return true
}

func (rp *RobotsParser) matchPath(path, pattern string) bool {
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = "^" + regexp.QuoteMeta(pattern) + "$"
	pattern = strings.ReplaceAll(pattern, "\\*\\.", ".*")
	matched, _ := regexp.MatchString(pattern, path)
	return matched
}

func (rp *RobotsParser) GetCrawlDelay() *time.Duration {
	for _, rule := range rp.rules {
		if rule.CrawlDelay != nil {
			return rule.CrawlDelay
		}
	}
	return nil
}