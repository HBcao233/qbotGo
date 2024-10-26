package twitter

import (
	"encoding/json"
	"reflect"
	"regexp"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/global/coolq"
	"github.com/Mrs4s/MiraiGo/message"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/HBcao233/qbotGo/util/data"
)

var pattern = regexp.MustCompile(`^(?:\[CQ:.*\] *)?(?:/?tid) ?(?:https?://)?(?:[a-z]*?(?:twitter|x)\.com/[a-zA-Z0-9_]+/status/)?(\d{13,20})(?:[^0-9a-z\n].*)?$`)

func twitter(bot *coolq.CQBot, event *coolq.Event) {
	if event.Raw.PostType != "message" && event.Raw.PostType != "message_sent" {
		return
	}
	if event.Raw.DetailType != "group" {
		return
	}
	text, _ := event.Raw.Others["raw_message"].(string)
	match := pattern.FindStringSubmatch(text)
	if len(match) == 0 {
		return
	}
	tid := match[1]
	if tid == "" {
		return
	}
	if csrf_token == "" || auth_token == "" {
		log.Warnf("csrf_token 或 auth_token 未设置")
		return
	}
	group_id, _ := event.Raw.Others["group_id"].(int64)
	settings := data.GetData("settings")
	allow_groups := settings.SetDefault("allow_groups", make([]float64, 0))
	data.SetData("settings", settings)
	if !containsInt(allow_groups, group_id) {
		return
	}
	log.Infof("event: %s", event.JSONString())
	log.Infof("tid: %s", tid)

	res := GetTwitter(tid)
	if res.Get("xerror").Exists() {
		bot.SendGroupMessage(group_id, &message.SendingMessage{
			Elements: []message.IMessageElement{message.NewText(res.Get("xerror").String())},
		})
		return
	}

	msg := ParseMsg(res)
	medias := ParseMedias(res)
	if len(medias) == 0 {
		bot.SendGroupMessage(group_id, &message.SendingMessage{
			Elements: []message.IMessageElement{message.NewText(msg)},
		})
		return
	}

	result := make([]string, 0)
	for _, url := range medias {
		fpath, _ := bot.CQDownloadFile(url, gjson.Result{}, 1)["data"].(map[string]any)["file"].(string)
		result = append(result, fpath)
	}

	m := []interface{}{
		map[string]interface{}{
			"type": "node",
			"data": map[string]interface{}{
				"name":    bot.Client.Nickname,
				"uin":     event.Raw.SelfID,
				"content": msg,
			},
		},
		map[string]interface{}{
			"type": "node",
			"data": map[string]interface{}{
				"name":    bot.Client.Nickname,
				"uin":     event.Raw.SelfID,
				"content": "https://x.com/i/status/" + tid,
			},
		},
	}
	for _, fpath := range result {
		m = append(m, map[string]interface{}{
			"type": "node",
			"data": map[string]interface{}{
				"name":    bot.Client.Nickname,
				"uin":     event.Raw.SelfID,
				"content": "[CQ:image,file=file:///" + fpath + "]",
			},
		})
	}
	m1, _ := json.Marshal(m)
	bot.CQSendGroupForwardMessage(group_id, gjson.Parse(string(m1)))
}

func init() {
	Init()
	bot.AddHandler(twitter)
}

func containsInt(slice interface{}, element int64) bool {
	if reflect.TypeOf(slice).Kind() == reflect.Slice {
		s := reflect.ValueOf(slice)
		for i := 0; i < s.Len(); i++ {
			e := s.Index(i)
			v, _ := e.Interface().(float64)
			if int64(v) == element {
				return true
			}
		}
	}
	return false
}
