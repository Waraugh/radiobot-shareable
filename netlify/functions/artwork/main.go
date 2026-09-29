package main

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type mbResponse struct {
	Recordings []struct {
		Score        int    `json:"score"`
		Title        string `json:"title"`
		ArtistCredit []struct {
			Name   string `json:"name"`
			Artist struct {
				Name string `json:"name"`
			} `json:"artist"`
		} `json:"artist-credit"`
		Releases []struct {
			ReleaseGroup struct {
				ID string `json:"id"`
			} `json:"release-group"`
		} `json:"releases"`
	} `json:"recordings"`
}

func norm(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(s)), " "))
}
func response(status int, p interface{}) *events.APIGatewayProxyResponse {
	b, _ := json.Marshal(p)
	return &events.APIGatewayProxyResponse{StatusCode: status, Headers: map[string]string{"Content-Type": "application/json; charset=utf-8", "Cache-Control": "public, max-age=86400, s-maxage=604800", "Access-Control-Allow-Origin": "*"}, Body: string(b)}
}
func handler(r events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	artist := strings.TrimSpace(r.QueryStringParameters["artist"])
	track := strings.TrimSpace(r.QueryStringParameters["track"])
	if artist == "" || track == "" {
		return response(400, map[string]string{"error": "artist and track are required"}), nil
	}
	q := `recording:"` + strings.ReplaceAll(track, `"`, ``) + `" AND artist:"` + strings.ReplaceAll(artist, `"`, ``) + `"`
	u := "https://musicbrainz.org/ws/2/recording/?fmt=json&limit=5&query=" + url.QueryEscape(q)
	c := &http.Client{Timeout: 6 * time.Second}
	req, _ := http.NewRequest("GET", u, nil)
	req.Header.Set("User-Agent", "RicanRadio/1.0 (https://rican-radio.netlify.app)")
	req.Header.Set("Accept", "application/json")
	res, e := c.Do(req)
	if e != nil {
		return response(502, map[string]string{"error": "artwork lookup failed"}), nil
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return response(502, map[string]string{"error": "metadata service unavailable"}), nil
	}
	var mb mbResponse
	if json.NewDecoder(res.Body).Decode(&mb) != nil {
		return response(502, map[string]string{"error": "invalid metadata response"}), nil
	}
	wt, wa := norm(track), norm(artist)
	for _, rec := range mb.Recordings {
		if rec.Score < 90 || norm(rec.Title) != wt || len(rec.Releases) == 0 {
			continue
		}
		match := false
		for _, ac := range rec.ArtistCredit {
			if norm(ac.Name) == wa || norm(ac.Artist.Name) == wa {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		for _, rel := range rec.Releases {
			if rel.ReleaseGroup.ID != "" {
				return response(200, map[string]string{"artwork": "https://coverartarchive.org/release-group/" + rel.ReleaseGroup.ID + "/front-500"}), nil
			}
		}
	}
	return response(404, map[string]string{"error": "no confident artwork match"}), nil
}
func main() { lambda.Start(handler) }
