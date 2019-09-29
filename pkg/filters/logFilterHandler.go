package filters

import (
	"bytes"
	"github.com/Alcereo/ordinator/pkg/common"
	log "github.com/sirupsen/logrus"
	"net/http"
	templ "text/template"
)

type LogFilterHandler struct {
	next     *common.RequestHandler
	template *templ.Template
	Name     string
}

func (filter *LogFilterHandler) SetNext(nextHandler common.RequestHandler) {
	filter.next = &nextHandler
}

func (filter *LogFilterHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	data := struct {
		Request *http.Request
		Filter  *LogFilterHandler
	}{
		request,
		filter,
	}
	var tpl bytes.Buffer
	err := filter.template.Execute(&tpl, data)
	if err != nil {
		log.Warnf("Log filter error: %v. Template error: %v", filter.Name, err)
	}

	log.Info(tpl.String())
	if filter.next != nil {
		(*filter.next).Handle(writer, request)
	} else {
		log.Debugf("Log filter error: %v. Next handler is empty", filter.Name)
	}

}

// Factory

func CreateLogFilter(name string, template string, next common.RequestHandler) *LogFilterHandler {
	parse, err := templ.New(name).Parse(template)
	if err != nil {
		log.Warnf("Log filter templ error: %v. Skip filter", err)
		return nil
	}
	return &LogFilterHandler{
		next:     &next,
		Name:     name,
		template: parse,
	}
}
