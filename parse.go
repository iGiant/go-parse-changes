package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"time"
	"strings"
	"github.com/go-ini/ini"
	"github.com/PuerkitoBio/goquery"
	"mylibs/slkclient"
)

type iniParam struct {
	file, url, selectors, attr, head, text string
	start, end, number int
}

const
	ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
		 "Chrome/73.0.3683.46 Safari/537.36 OPR/60.0.3255.8 (Edition beta)"

func parsing(ip iniParam) (string, error) {
	var (
		result []string
		item string
	)

	req, err := http.NewRequest("GET", ip.url, nil)
	if err != nil {
		return "", fmt.Errorf("Error reading request: %v", err)
	}
	req.Header.Set("User-Agent", ua)

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code error: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading body: %v", err)
	}
	doc.Find(ip.selectors).Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		if ip.attr == "text" {
			item = s.Text()
		} else {
			item, _ = s.Attr(ip.attr)
		}
		result = append(result, item)

	})
	if len(result) == 0 {
		return "", fmt.Errorf("No items found")
	}
	if ip.number < 0 {
		ip.number = len(result) + ip.number
	}
	if ip.start != -1 && ip.end != -1 && ip.end > ip.start {
		return result[ip.number][ip.start:ip.end], nil
	}
	return result[ip.number], nil
}

func getIni(filename string) (ip iniParam, err error) {
	cfg, err := ini.Load(filename)
    if err != nil {
        return
	}

	ip.file = cfg.Section("main").Key("file").String()
	ip.url = cfg.Section("main").Key("url").String()
	ip.selectors = cfg.Section("main").Key("selectors").String()
	ip.selectors = strings.ReplaceAll(ip.selectors, "№", "#")
	ip.attr = cfg.Section("main").Key("attr").String()
	ip.start, err = cfg.Section("main").Key("start").Int()
	if err != nil {
        return
	}
	ip.end, err = cfg.Section("main").Key("end").Int()
	if err != nil {
        return
	}
	ip.number, err = cfg.Section("main").Key("number").Int()
	if err != nil {
        return
	}
	ip.head = cfg.Section("main").Key("head").String()
	ip.text = cfg.Section("main").Key("text").String()
	return
}

func getOldText(filename string) (result string, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	result = string(content)
	return
}

func main() {
	filenames, err := ioutil.ReadDir(".")
	if err == nil {
		var (
			ip iniParam
			newtext, oldtext string
			err error
		)
		for _, filename := range filenames {
			if strings.HasSuffix(filename.Name(), ".ini") {
				ip, err = getIni(filename.Name())
				if err != nil {
					continue
				}
				newtext, err = parsing(ip)
				if err != nil {
					continue
				}
				oldtext, _ = getOldText(ip.file)
				if newtext != oldtext {
					slkclient.SendToSlack(ip.head, fmt.Sprintf(ip.text, newtext), "", "", "")
					ioutil.WriteFile(ip.file, []byte(newtext), 0666)
				}
			}
	}
}
}
// go build -ldflags "-H windowsgui" <file.go>