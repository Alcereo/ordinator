package filters

import (
	"balancer/balancer"
	"log"
	"net/http"
	"os"
	templ "text/template"
)

type LogFilterHandler struct {
	next     *balancer.RequestHandler
	template *templ.Template
	Name     string
}

func (filter *LogFilterHandler) SetNext(nextHandler balancer.RequestHandler) {
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
	err := filter.template.Execute(os.Stdout, data)
	if err != nil {
		log.Printf("Log filter error: %v. Template error: %v", filter.Name, err)
	}
	if filter.next != nil {
		(*filter.next).Handle(writer, request)
	} else {
		log.Printf("Log filter error: %v. Next handler is empty", filter.Name,)
	}

}

// Factory

func CreateLogFilter(name string, template string, next balancer.RequestHandler) *LogFilterHandler {
	parse, err := templ.New(name).Parse(template)
	if err != nil {
		log.Printf("Log filter templ error: %v. Skip filter", err)
		return nil
	}
	return &LogFilterHandler{
		next: &next,
		Name: name,
		template: parse,
	}
}