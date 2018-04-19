package main

import (
	"net/http"
	"log"
	"github.com/PuerkitoBio/goquery"
	"strings"
	"encoding/json"
	"fmt"
)

type offer struct {
	Text string    `json:"text"`
	Img  []string  `json:"img"`
	Info otherInfo `json:"info"`
}

type otherInfo struct {
	UrlUser  string `json:"url_user"`
	TimePost string `json:"time_post"`
	UserName string `json:"user_name"`
	UrlPost  string `json:"url_post"`
}

type offers struct {
	WallItem   []offer `json:"wall_item"`
	LastIdPost string  `json:"last_id_post"`
}

func findJpg(iteration int, doc *goquery.Document) []string {
	var urls []string
	doc.Find(".thumbs_map.fill").Each(func(i int, s *goquery.Selection) {
		if iteration == i {
			s.Find("div").Each(func(_ int, s *goquery.Selection) {
				urlWithGrabage, ok := s.Attr("data-src_big")
				if ok {
					urlClear := strings.Split(urlWithGrabage, "|")[0]
					urls = append(urls, urlClear)
				}
			})
		}
	})
	return urls
}

func getLastIdPost(doc *goquery.Document) string {
	var postId string
	doc.Find(".wall_item").Each(func(i int, s *goquery.Selection) {

		if i != 0 {
			return
		}
		postId, _ = s.Find("a").Attr("name")
		postId = strings.Split(postId, "-")[1]

	})
	return postId
}

func getIdPost(number int, doc *goquery.Document) string {
	var postId string
	doc.Find(".wall_item").Each(func(i int, s *goquery.Selection) {

		if i != number {
			return
		}

		postId, _ = s.Find("a").Attr("name")
		postId = strings.Split(postId, "-")[1]

	})
	return postId
}

func getOtherInfo(number int, doc *goquery.Document) otherInfo {
	var info otherInfo
	doc.Find(".wall_item").Each(func(i int, s *goquery.Selection) {
		if number != i {
			return
		}
		s.Find(".wi_head").Each(func(_ int, s *goquery.Selection) {

			s.Find("a").Each(func(i int, s *goquery.Selection) {
				switch i {
				case 0:
					info.UrlUser, _ = s.Attr("href")
					info.UrlUser = "vk.com" + info.UrlUser
				case 2:
					info.UrlPost, _ = s.Attr("href")
					info.UrlPost = "vk.com" + info.UrlPost
				}

			})
			info.TimePost = s.Find(".wi_date").Text()
			info.UserName = s.Find(".pi_author").Text()
		})
	})

	return info
}

func (of offers) fetchInfo(last string) offers {
	doc := scarpe()
	flagLast := false
	doc.Find(".pi_text").Each(func(i int, wi_body *goquery.Selection) {
		if last == getIdPost(i, doc) {
			flagLast = true
		}
		if flagLast {
			return
		}
		of.WallItem = append(of.WallItem, offer{Text: wi_body.Text(), Img: findJpg(i, doc), Info: getOtherInfo(i, doc)})
	})

	of.LastIdPost = getLastIdPost(doc)
	return of
}

func main() {

	http.HandleFunc("/", hanlder)
	http.ListenAndServe(":8080", nil)
}

func scarpe() *goquery.Document {
	res, err := http.Get("https://vk.com/otdadimka")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	//	io.Copy(os.Stdout, res.Body)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return doc
}

func hanlder(w http.ResponseWriter, r *http.Request) {
	var o offers
	last := r.FormValue("last_post")
	fmt.Println(last)
	o = o.fetchInfo(last)
	byteJS, _ := json.Marshal(o)
	fmt.Fprintln(w, string(byteJS))
}
