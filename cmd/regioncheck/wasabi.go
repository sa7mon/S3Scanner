package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
)

type nextData struct {
	Props struct {
		PageProps struct {
			Story struct {
				Name    string `json:"name"`
				Content struct {
					Body []struct {
						ID      string `json:"id"`
						UID     string `json:"_uid"`
						Title   string `json:"Title"`
						Width   string `json:"width"`
						Columns []struct {
							Components []struct {
								UID     string `json:"_uid"`
								Buttons []struct {
									UID  string `json:"_uid"`
									Icon string `json:"icon"`
									Link struct {
										ID        string `json:"id"`
										URL       string `json:"url"`
										Linktype  string `json:"linktype"`
										Fieldtype string `json:"fieldtype"`
										CachedURL string `json:"cached_url"`
										Story     struct {
											Name          string `json:"name"`
											ID            int    `json:"id"`
											UUID          string `json:"uuid"`
											Slug          string `json:"slug"`
											URL           string `json:"url"`
											FullSlug      string `json:"full_slug"`
											StopResolving bool   `json:"_stopResolving"`
										} `json:"story"`
									} `json:"link"`
									Type      string `json:"type"`
									Color     string `json:"color"`
									Label     string `json:"label"`
									Component string `json:"component"`
									IconAlign string `json:"iconAlign"`
								} `json:"buttons"`
								Component     string `json:"component"`
								ParagraphText struct {
									Type    string `json:"type"`
									Content []struct {
										Type    string `json:"type"`
										Content []struct {
											Text string `json:"text"`
											Type string `json:"type"`
										} `json:"content"`
									} `json:"content"`
								} `json:"paragraphText"`
							} `json:"components"`
							CustomColor struct {
								Value  string `json:"value"`
								Plugin string `json:"plugin"`
							} `json:"customColor"`
							VerticalAlign      string `json:"verticalAlign"`
							BackgroundColor    string `json:"backgroundColor"`
							BackgroundPosition string `json:"backgroundPosition"`
						} `json:"columns"`
						Component string `json:"component"`
					} `json:"body"`
				} `json:"content"`
				Slug       string `json:"slug"`
				Alternates []struct {
					ID        int    `json:"id"`
					Name      string `json:"name"`
					Slug      string `json:"slug"`
					Published bool   `json:"published"`
					FullSlug  string `json:"full_slug"`
					IsFolder  bool   `json:"is_folder"`
					ParentID  int    `json:"parent_id"`
				} `json:"alternates"`
				DefaultFullSlug any `json:"default_full_slug"`
				TranslatedSlugs any `json:"translated_slugs"`
			} `json:"story"`
			Key     int    `json:"key"`
			Locale  string `json:"locale"`
			Name    string `json:"name"`
			Sidebar bool   `json:"sidebar"`
		} `json:"pageProps"`
		NSsg bool `json:"__N_SSG"`
	} `json:"props"`
	Page  string `json:"page"`
	Query struct {
		Slug []string `json:"slug"`
	} `json:"query"`
}

func getRegionsWasabi() ([]string, error) {
	requestURL := "https://wasabi.com/company/storage-regions"
	res, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	regions := []string{}
	doc.Find("main h3~p:has(b)").Contents().Each(func(i int, t *goquery.Selection) {
		if goquery.NodeName(t) == "#text" {
			rangeNames := strings.Split(t.Text(), "&")
			for _, r := range rangeNames {
				regions = append(regions, strings.TrimSpace(r))
			}
		}
	})
	return regions, nil
}
