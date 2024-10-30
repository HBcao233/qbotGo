package twitter

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"regexp"

	"github.com/Logiase/MiraiGo-Template/client"
	"github.com/Logiase/MiraiGo-Template/global"
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

	fm := message.NewForwardMessage().AddNode(&message.ForwardNode{
		GroupId:    group_id,
		SenderId:   event.Raw.SelfID,
		SenderName: bot.Client.Nickname,
		Message:    []message.IMessageElement{message.NewText(msg)},
	}).AddNode(&message.ForwardNode{
		GroupId:    group_id,
		SenderId:   event.Raw.SelfID,
		SenderName: bot.Client.Nickname,
		Message:    []message.IMessageElement{message.NewText("x.com/i/status/" + tid)},
	})
	var i message.IMessageElement
	for _, m := range medias {
		switch m.Type {
		case Photo:
			fpath, _ := bot.CQDownloadFile(m.Url, gjson.Result{}, 1)["data"].(map[string]any)["file"].(string)
			f := global.ReadFile(fpath)
			if f == nil {
				log.Error("读取图片失败")
				return
			}
			token := make([]byte, 8)
			rand.Read(token)
			f = append(f, token...)

			source := message.Source{
				SourceType: message.SourceGroup,
				PrimaryID:  group_id,
			}
			var err error
			i, err = bot.Client.UploadImage(source, bytes.NewReader(f))
			if err != nil {
				log.Error("图片上传失败")
				return
			}
		case Video:
			url := m.Variants[0].Url
			i = message.NewText("[视频: " + url + "]")
		}

		fm.AddNode(&message.ForwardNode{
			GroupId:    group_id,
			SenderId:   event.Raw.SelfID,
			SenderName: bot.Client.Nickname,
			Message:    []message.IMessageElement{i},
		})
	}
	fe := bot.Client.NewForwardMessageBuilder(group_id).Main(fm)
	m := bot.Client.SendGroupForwardMessage(group_id, fe)
	if m == nil || m.Id == -1 {
		log.Warn("合并转发消息发送失败，可能被风控")
	}
}

func init() {
	Init()
	client.AddHandler(twitter)
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
