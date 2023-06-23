package todo

import (
	"html/template"
	"net/http"
)

var tmpl *template.Template

type List struct {
	Object string
	Finish bool
}

type PageInfo struct {
	Title string 
	Todos []List
}

func lis(w http.ResponseWriter, r *http.Request) {
	data:= PageInfo{
		Title : "Todo List",
		Todos : []List{
			{Object: "write script", Finish: true},
			{Object: "shoot video", Finish: true},
			{Object: "edit the video", Finish: false},
		},
	}
	tmpl.Execute(w, data)
}