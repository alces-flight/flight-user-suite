package main

import (
	"fmt"
	"html/template"
	"io"

	"github.com/labstack/echo/v5"
)

// TemplateRenderer is a custom renderer that supports jinja2-inspired template
// inheritance and custom layouts.
//
//   - Templates are split into "layouts", "partials" and "pages".
//   - On startup, all layout and partial templates should be parsed. They must
//     not `define` conflicting sections.
//   - Layouts define a "content" section that pages plug.
//   - All pages define a "content" that is plugged into the layout.  These
//     necessarily conflict, so only a single page can be added  to a template set.
//   - On each request, the base set of templates is cloned and the single page
//     being rendered is added to the set.
//   - If the data argument contains a "layout" entry that layout is used to
//     render the template. The "application" layout is used by default.
type TemplateRenderer struct {
	// The base set of templates that can be parsed on startup.
	Template *template.Template
}

// Render implements the echo.Renderer interface.
func (tr *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	clone := template.Must(tr.Template.Clone())
	pageTemplate := fmt.Sprintf("%s/pages/%s.gohtml", getDirectory("views"), name)
	withPage, err := clone.ParseFiles(pageTemplate)
	if err != nil {
		return err
	}
	if data == nil {
		data = make(map[string]any)
	}
	switch data := data.(type) {
	case map[string]any:
		layout := "application"
		if l, ok := data["layout"]; ok {
			switch l := l.(type) {
			case string:
				layout = l
			default:
				c.Logger().Warn("invalid layout type", "type", fmt.Sprintf("%T", l))
			}
		}
		return withPage.ExecuteTemplate(w, fmt.Sprintf("layouts.%s", layout), AddCommonData(c, data))
	default:
		return fmt.Errorf("unsupported data format got %T want map[string]any", data)
	}
}
