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

const
	ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) " +
		 "Chrome/73.0.3683.46 Safari/537.36 OPR/60.0.3255.8 (Edition beta)"

func parsing(url, selectors, attr string, num int) (string, error) {
	var (
		result []string
		item string
	)

	req, err := http.NewRequest("GET", url, nil)
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
	doc.Find(selectors).Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		if attr == "text" {
			item = s.Text()
		} else {
			item, _ = s.Attr(attr)
		}
		result = append(result, item)

	})
	if len(result) == 0 {
		return "", fmt.Errorf("No items found")
	}
	if num < 0 {
		num = len(result) + num
	}
	// fmt.Println(num, len(result))
	return result[num], nil
}

func getIni(filename string) (file, url, selectors, attr string, 
							  number int, head, text string, err error) {
	cfg, err := ini.Load(filename)
    if err != nil {
        return
	}
	file = cfg.Section("main").Key("file").String()
	url = cfg.Section("main").Key("url").String()
	selectors = cfg.Section("main").Key("selectors").String()
	attr = cfg.Section("main").Key("attr").String()
	number, err = cfg.Section("main").Key("number").Int()
	if err != nil {
        return
	}
	head = cfg.Section("main").Key("head").String()
	text = cfg.Section("main").Key("text").String()
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
			file, url, selectors, attr, head, text, newtext, oldtext string
			number int
			err error
		)
		for _, filename := range filenames {
			if strings.HasSuffix(filename.Name(), ".ini") {
				file, url, selectors, attr, number, head, text, err = getIni(filename.Name())
				if err != nil {
					continue
				}
				newtext, err = parsing(url, selectors, attr, number)
				if err != nil {
					continue
				}
				oldtext, _ = getOldText(file)
				if newtext != oldtext {
					slkclient.SendToSlack(head, fmt.Sprintf(text, newtext), "", "", "")
					ioutil.WriteFile(file, []byte(newtext), 0666)
				}
			}
	}
}
}