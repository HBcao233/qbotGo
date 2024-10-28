package html

import (
	"html/template"
	"net/http"

	"github.com/Logiase/MiraiGo-Template/client"
	c "github.com/Mrs4s/MiraiGo/client"
)

var tpl *template.Template

func init() {
	go main()
}

func main() {
	// 定义一个 HTTP 请求处理函数
	home()
	api()

	http.ListenAndServe(":8080", nil)
}

func home() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tpl = template.Must(template.ParseFiles("static/index.html"))
		groupList := make([]*c.GroupInfo, 0)
		if client.GetBot() != nil {
			groupList = append(groupList, client.GetClient().GroupList...)
		}
		tpl.Execute(w, map[string]any{
			"GroupList": groupList,
		})
	})
}
