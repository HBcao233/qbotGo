package twitter

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/HBcao233/qbotGo/util/data"
	"github.com/tidwall/gjson"
)

var (
	features   string
	csrf_token string
	auth_token string
)

func Init() {
	_features, _ := json.Marshal(map[string]bool{
		"rweb_lists_timeline_redesign_enabled":                                    true,
		"responsive_web_graphql_exclude_directive_enabled":                        true,
		"verified_phone_label_enabled":                                            false,
		"creator_subscriptions_tweet_preview_api_enabled":                         true,
		"responsive_web_graphql_timeline_navigation_enabled":                      true,
		"responsive_web_graphql_skip_user_profile_image_extensions_enabled":       false,
		"tweetypie_unmention_optimization_enabled":                                true,
		"responsive_web_edit_tweet_api_enabled":                                   true,
		"graphql_is_translatable_rweb_tweet_is_translatable_enabled":              true,
		"view_counts_everywhere_api_enabled":                                      true,
		"longform_notetweets_consumption_enabled":                                 true,
		"responsive_web_twitter_article_tweet_consumption_enabled":                false,
		"tweet_awards_web_tipping_enabled":                                        false,
		"freedom_of_speech_not_reach_fetch_enabled":                               true,
		"standardized_nudges_misinfo":                                             true,
		"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled": true,
		"longform_notetweets_rich_text_read_enabled":                              true,
		"longform_notetweets_inline_media_enabled":                                true,
		"responsive_web_media_download_video_enabled":                             false,
		"responsive_web_enhance_cards_enabled":                                    false,
	})
	features = string(_features)
	load_settings()
}

func load_settings() {
	settings := data.GetData("settings")
	csrf_token, _ = settings.SetDefault("twitter_csrf_token", "").(string)
	auth_token, _ = settings.SetDefault("twitter_auth_token", "").(string)
	data.SetData("settings", settings)
}

func GetTwitter(tid string) gjson.Result {
	Url, _ := url.Parse("https://x.com/i/api/graphql/NmCeCgkVlsRGS1cAwqtgmw/TweetDetail")
	params := url.Values{}

	variables, _ := json.Marshal(map[string]any{
		"focalTweetId":                           tid,
		"with_rux_injections":                    false,
		"includePromotedContent":                 true,
		"withCommunity":                          true,
		"withQuickPromoteEligibilityTweetFields": true,
		"withBirdwatchNotes":                     true,
		"withVoice":                              true,
		"withV2Timeline":                         true,
	})

	params.Set("variables", string(variables))
	params.Set("features", features)
	Url.RawQuery = params.Encode()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", Url.String(), nil)
	req.Header.Add("referer", fmt.Sprintf("https://x.com/i/status/%s", tid))
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
	req.Header.Add("content-type", "application/json; charset=utf-8")
	req.Header.Add("authorization", "Bearer AAAAAAAAAAAAAAAAAAAAANRILgAAAAAAnNwIzUejRCOuH5E6I8xnZz4puTs%3D1Zv7ttfk8LF81IUq16cHjhLTvJu4FA33AGWWjCpTnA")
	req.Header.Add("x-csrf-token", csrf_token)
	req.Header.Add("cookie", fmt.Sprintf("auth_token=%s; ct0=%s", auth_token, csrf_token))
	req.Header.Add("X-Twitter-Client-Language", "zh-cn")
	req.Header.Add("X-Twitter-Active-User", "yes")

	r, err := client.Do(req)
	if err != nil {
		return gjson.Parse(`{"xerror":"连接超时"}`)
	}
	body, _ := io.ReadAll(r.Body)
	res := gjson.Parse(string(body))
	if errors := res.Get("errors"); errors.Exists() && len(errors.Array()) > 0 {
		if res.Get("errors.0.code").Int() == 144 {
			return gjson.Parse(`{"xerror":"推文不存在"}`)
		}
		s := strings.Replace(res.Get("errors.0.message").String(), "you", "", 1)
		return gjson.Parse(`{"xerror":"` + s + `"}`)
	}
	entries := res.Get("data.threaded_conversation_with_injections_v2.instructions.0.entries")
	var tweet_entrie []gjson.Result
	entries.ForEach(func(_, i gjson.Result) bool {
		if i.Get("entryId").String() == "Tweet-"+tid || i.Get("entryId").String() == "tweet-"+tid {
			tweet_entrie = append(tweet_entrie, i)
		}
		return true
	})
	if len(tweet_entrie) == 0 {
		return gjson.Parse(`{"xerror":"解析失败"}`)
	}
	tweet_result := tweet_entrie[0].Get("content.itemContent.tweet_results.result")
	if tweet_result.Get("tweet").Exists() {
		return tweet_result.Get("tweet")
	}
	return tweet_result
}

func ParseMsg(res gjson.Result) string {
	tweet := res.Get("legacy")
	user := res.Get("core.user_results.result.legacy")

	// tid = tweet['id_str']
	full_text := tweet.Get("full_text").String()
	if tweet.Get("entities.urls").Exists() {
		tweet.Get("entities.urls").ForEach(func(_, value gjson.Result) bool {
			full_text = strings.ReplaceAll(full_text, value.Get("url").String(), value.Get("expanded_url").String())
			return true
		})
	}
	re := regexp.MustCompile(`\s*https:\/\/t\.co\/\w+$`)
	full_text = re.ReplaceAllString(full_text, "")
	// full_text = re.sub(r'#([^ \n#]+)', r'<a href="https://x.com/hashtag/\1">#\1</a>', full_text)
	// full_text = re.sub(r'([^@]*[^/@]+)@([0-9a-zA-Z_]*)', r'\1<a href="https://x.com/\2">@\2</a>', full_text)

	nickname := user.Get("name").String()
	// username = user['screen_name']
	// utc_time = time.strptime(tweet['created_at'], r'%a %b %d %H:%M:%S %z %Y')
	// local_time = time.localtime(
	//   time.mktime(utc_time) + utc_time.tm_gmtoff - time.timezone
	// )
	// create_time = time.strftime('%Y年%m月%d日 %H:%M:%S', local_time)
	msg := nickname
	if full_text != "" {
		msg = msg + ":\n" + full_text
	}
	return msg
}

type MediaType int

const (
	Photo = 0
	Video = 1
)

type Media struct {
	Type     MediaType
	Url      string
	Thumb    string
	Variants Variants
}
type Variants []Variant
type Variant struct {
	Bitrate int64
	Url     string
}

func (v Variants) Len() int           { return len(v) }
func (v Variants) Less(i, j int) bool { return v[i].Bitrate > v[j].Bitrate }
func (v Variants) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

func ParseMedias(res gjson.Result) []Media {
	tweet := res.Get("legacy")
	result := make([]Media, 0)
	if !tweet.Get("extended_entities").Exists() {
		return result
	}
	medias := tweet.Get("extended_entities.media")
	medias.ForEach(func(_, media gjson.Result) bool {
		switch media.Get("type").String() {
		case "photo":
			result = append(result, Media{
				Type:  Photo,
				Url:   media.Get("media_url_https").String() + ":orig",
				Thumb: media.Get("media_url_https").String() + "::small",
			})
		case "video":
			variants := make(Variants, 0)
			media.Get("video_info.variants").ForEach(func(_, v gjson.Result) bool {
				if v.Get("content_type").String() == "video/mp4" {
					variants = append(variants, Variant{
						Bitrate: v.Get("bitrate").Int(),
						Url:     v.Get("url").String(),
					})
				}
				return true
			})
			sort.Sort(variants)
			result = append(result, Media{
				Type:     Video,
				Variants: variants,
			})
		}
		return true
	})
	return result
}
