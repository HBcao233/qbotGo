package html

import (
	"net/http"

	"github.com/Logiase/MiraiGo-Template/client"
	log "github.com/sirupsen/logrus"
)

func api() {
	group()
}

func group() {
	getMessages()
	getImageUrl()
}

func getMessages() {
	http.HandleFunc("/api/group/getMessages", func(w http.ResponseWriter, r *http.Request) {
		code, err := paramInt64(r, "code")
		offset, err1 := paramInt64(r, "offset")
		count, err2 := paramInt64(r, "count")
		if err != nil {
			fail(w, nil, "")
			return
		}
		if err1 != nil {
			offset = 0
		}
		if err2 != nil {
			count = 10
		}
		log.Infof("code: %d, offset: %d, count: %d", code, offset, count)

		messages := make([]any, 0)
		var lastMsgSeq int64
		cli := client.GetClient()
		if g, err := cli.GetGroupInfo(code); err == nil {
			lastMsgSeq = g.LastMsgSeq
			if history, err := cli.GetGroupMessages(code, lastMsgSeq-offset-count, lastMsgSeq+1-offset); err == nil {
				for _, v := range history {
					messages = append(messages, map[string]any{
						"id":         v.Id,
						"group_code": code,
						"group_name": v.GroupName,
						"time":       v.Time,
						"elements":   v.Elements,
						"sender": map[string]any{
							"uin":            v.Sender.Uin,
							"nickname":       v.Sender.Nickname,
							"cardname":       v.Sender.CardName,
							"anonymous_info": v.Sender.AnonymousInfo,
							"is_friend":      v.Sender.IsFriend,
						},
					})
				}
			}
		}
		success(w, map[string]any{
			"messages":     messages,
			"last_msg_seq": lastMsgSeq,
			"offset":       offset,
			"count":        count,
			"actual_count": len(messages),
		}, "")
	})
}

func getImageUrl() {
	http.HandleFunc("/api/group/getImageUrl", func(w http.ResponseWriter, r *http.Request) {
		code, err := paramInt64(r, "code")
		fileId, err1 := paramInt64(r, "fileId")
		md5 := paramString(r, "md5")
		if err != nil || err1 != nil {
			fail(w, nil, "")
			return
		}

		cli := client.GetClient()
		url, err := cli.GetGroupImageDownloadUrl(fileId, code, []byte(md5))
		if err != nil {
			fail(w, nil, "获取失败")
			return
		}
		success(w, map[string]any{
			"url": url,
		}, "")
	})
}
