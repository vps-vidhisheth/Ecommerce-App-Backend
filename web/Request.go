package web

import (
	"ecommerce/components/log"
	"ecommerce/errors"
	"encoding/json"
	"strings"

	"io"
	"net/http"
)

const MaxBodySize = 10 << 20

func UnmarshalJSON(request *http.Request, out interface{}) error {
	log := log.GetLogger()
	defer request.Body.Close()
	if request.Body == nil {

		log.Print("err request body is nil")
		return errors.NewHTTPError(errors.ErrorCodeEmptyRequestBody, http.StatusBadRequest)
	}

	request.Body = http.MaxBytesReader(nil, request.Body, MaxBodySize) //limits the size of reader

	body, err := io.ReadAll(request.Body)
	if err != nil {

		if strings.Contains(err.Error(), "http: request body too large") {
			log.Print("err: request body exceeded max size")
			return errors.NewHTTPError("Request body too large. Maximum allowed is 1MB.", http.StatusRequestEntityTooLarge)
		}

		log.Print("err io.readall")
		return errors.NewHTTPError(err.Error(), http.StatusBadRequest)
	}
	if len(body) == 0 {
		log.Print("err len(body)==0 ")
		return errors.NewHTTPError(errors.ErrorCodeEmptyRequestBody, http.StatusBadRequest)
	}
	if err := json.Unmarshal(body, out); err != nil {
		log.Print("rr json.Unmarshal")
		log.Print(body)
		log.Print(out)
		return errors.NewHTTPError(err.Error(), http.StatusBadRequest)
	}
	return nil
}
