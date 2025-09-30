package web

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
)

type Parser struct {
	Params map[string]string
	Form   url.Values
}

func NewParser(r *http.Request) *Parser {
	r.ParseForm()
	return &Parser{
		Params: mux.Vars(r),
		Form:   r.Form,
	}
}

func (p *Parser) GetParameter(paramName string) string {
	paramString := p.Params[paramName]
	return paramString
}

func (p *Parser) ParseLimitAndOffset() (limit, offset int) {
	limitparam := p.Form.Get("limit")
	offsetparam := p.Form.Get("offset")
	var err error
	if len(limitparam) > 0 {
		limit, err = strconv.Atoi(limitparam)
		if err != nil {
			return
		}
	}
	if len(offsetparam) > 0 {
		offset, err = strconv.Atoi(offsetparam)
		if err != nil {
			return
		}
	}
	return
}
