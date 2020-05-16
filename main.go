package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jtguibas/cinema"
)

// Response Instagram fetch type
type Response struct {
	Graphql struct {
		Hashtag struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Media struct {
				Count    int `json:"count"`
				PageInfo struct {
					HasNextPage bool   `json:"has_next_page"`
					EndCursor   string `json:"end_cursor"`
				} `json:"page_info"`
				Edges []struct {
					Node struct {
						TypeName  string `json:"__typename"`
						ShortCode string `json:"shortcode"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"edge_hashtag_to_media"`
		} `json:"hashtag"`
	} `json:"graphql"`
}

// Info type
type Info struct {
	Graphql struct {
		ShortCodeMedia struct {
			VideoURL   string `json:"video_url"`
			DisplayURL string `json:"display_url"`
		} `json:"shortcode_media"`
	} `json:"graphql"`
}

// PageInfo type
type PageInfo struct {
	HasNextPage bool   `json:"has_next_page"`
	EndCursor   string `json:"end_cursor"`
	Count       int
}

var config PageInfo

func main() {
	for {
		response := fetchInstagram("thailand", config.EndCursor)
		fmt.Println("COUNT => ", config.Count)
		if !response {
			break
		}
	}
}

func fetchInstagram(TagName string, MaxID string) bool {
	url := "https://www.instagram.com/explore/tags/" + TagName + "/?__a=1"
	if MaxID != "" {
		url += "&max_id=" + MaxID
	}
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var v Response
	json.Unmarshal([]byte(body), &v)
	for i, Edge := range v.Graphql.Hashtag.Media.Edges {
		if Edge.Node.TypeName == "GraphVideo" {
			fmt.Println(i, Edge.Node.ShortCode)
			info := fetchInfo(Edge.Node.ShortCode)
			if info {
				config.Count++
			}
		}
	}
	config.EndCursor = v.Graphql.Hashtag.Media.PageInfo.EndCursor
	config.HasNextPage = v.Graphql.Hashtag.Media.PageInfo.HasNextPage
	return v.Graphql.Hashtag.Media.PageInfo.HasNextPage
}

func fetchInfo(ShortCode string) bool {
	url := "https://www.instagram.com/p/" + ShortCode + "?__a=1"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	var v Info
	json.Unmarshal([]byte(body), &v)
	os.MkdirAll("temp", os.ModePerm)
	os.MkdirAll("thumnail", os.ModePerm)
	os.MkdirAll("trim", os.ModePerm)
	// Download Video and trim
	if err := DownloadFile("temp/"+ShortCode+".mp4", v.Graphql.ShortCodeMedia.VideoURL); err != nil {
		panic(err)
	} else {
		video, err := cinema.Load("temp/" + ShortCode + ".mp4")
		check(err)
		video.Trim(0*time.Second, 1*time.Second)
		video.Render("trim/" + ShortCode + ".mp4")
		os.RemoveAll("temp")
	}
	// Download Thumnail
	if err := DownloadFile("thumnail/"+ShortCode+".jpg", v.Graphql.ShortCodeMedia.DisplayURL); err != nil {
		panic(err)
	}
	return true
}

// DownloadFile functions
func DownloadFile(filepath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
