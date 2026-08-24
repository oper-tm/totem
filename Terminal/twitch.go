package main

import (
	"errors"
	"encoding/json"
	"net/url"
)

var (
	// csjd any
	pldjd any
)

func TgetSideNavList() (error, error, []any){
	p := map[string]interface{}{
		"operationName": "SideNav",
		"variables": map[string]interface{}{
			"input": map[string]interface{}{
				"recommendationContext": map[string]interface{}{
					"platform": "web",
					"clientApp": "twilight",
					"channelName": nil,
					"categorySlug": nil,
					"lastChannelName": nil,
					"lastCategorySlug": nil,
					"pageviewContent": nil,
					"pageviewContentType": nil,
					"pageviewLocation": nil,
					"pageviewMedium": nil,
					"previousPageviewContent": nil,
					"previousPageviewContentType": nil,
					"previousPageviewLocation": nil,
					"previousPageviewMedium": nil,
				},
			},
			"creatorAnniversariesFeature": false,
    		"withFreeformTags": false,
		},
		"extensions": map[string]interface{}{
			"persistedQuery": map[string]interface{}{
				"version": 1,
				"sha256Hash": "40418288329fcecbbf422c5a7cbcc5a937f5670550a9d3b246b8327ff1903ba1",
			},
		},
	}
	h := map[string]interface{}{
		"Client-ID": "kimne78kx3ncx6brgo4mv6wki5h1ko",
	}

	errw, err, sl_Data := getFilePost("https://gql.twitch.tv/gql", "test.json", p, h, false)
	if err != nil {
		return errw, err, []any("err")
	}

	err = json.Unmarshal([]byte(sl_Data), &csjd)
	if err != nil {
		return errors.New("tgml:1"), err, []any("err")
	}

	cs_m := csjd.(map[string]any)
	cs_data := cs_m["data"].(map[string]any)
	cs_sideNav := cs_data["sideNav"].(map[string]any)
	cs_sections := cs_sideNav["sections"].(map[string]any)
	cs_edges := cs_sections["edges"].([]any)[0].(map[string]any)
	cs_node := cs_edges["node"].(map[string]any)
	cs_content := cs_node["content"].(map[string]any)
	cs_cedges := cs_content["edges"].([]any)
	return nil, nil, cs_cedges
}

func TgetStreamM3U8(name string) (error, error, string) {
	p := map[string]interface{}{
		"operationName": "PlaybackAccessToken",
		"query": `query PlaybackAccessToken($login: String!) {
			streamPlaybackAccessToken(
			channelName: $login,
			params: {
				platform: "web",
				playerBackend: "mediaplayer",
				playerType: "site"
			}
			) {
			value
			signature
			}
		}`,
		"variables": map[string]interface{}{
			"login": name,
		},
	}
	h := map[string]interface{}{
		"Client-ID": "kimne78kx3ncx6brgo4mv6wki5h1ko",
	}

	errw, err, sl_Data := getFilePost("https://gql.twitch.tv/gql", "test.json", p, h, false)
	if err != nil {
		return errw, err, "err"
	}

	err = json.Unmarshal([]byte(sl_Data), &pldjd)
	if err != nil {
		return errors.New("tgsm:1"), err, "err"
	}

	m := pldjd.(map[string]any)
	data := m["data"].(map[string]any)
	token := data["streamPlaybackAccessToken"].(map[string]any)

	value := token["value"].(string)
	signature := token["signature"].(string)

	encodedValue := url.QueryEscape(value)
	return nil, nil, "https://usher.ttvnw.net/api/v2/channel/hls/" + name + ".m3u8?acmb=eyJBcHBWZXJzaW9uIjoiNDlhNTk2N2UtOTE1NC00MDk0LThjMzYtZDkzZGUxY2NmMWI2IiwiQ2xpZW50QXBwIjoidHdpbGlnaHQiLCJVUkwiOiJodHRwczovL3R3aXRjaC50di9qeW54emkifQ%3D%3D&allow_source=true&browser_family=edge&browser_version=149.0&cdm=wv&enable_score=true&fast_bread=true&include_unavailable=true&lang=en&os_name=Windows&os_version=NT%2010.0&p=1332241&platform=web&play_session_id=6acf1a7845b1479191e9fbe1be24ec25&player_backend=mediaplayer&player_version=1.54.0-rc.1&playlist_include_framerate=true&reassignments_supported=true&sig=" + signature + "&supported_codecs=av1,h264&token=" + encodedValue + "&transcode_mode=cbr_v1"
}