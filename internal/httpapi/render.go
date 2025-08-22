package httpapi

import (
	"net/http"
	"text/template"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmpls := []string{"./web/layout.html", "./web/" + tmpl + ".html"}
	t, err := template.ParseFiles(tmpls...)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error parsing templates", err)
		return
	}

	err = t.ExecuteTemplate(w, "layout", data)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "error executing templates", err)
		return
	}
}
