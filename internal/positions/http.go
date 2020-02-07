package positions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (p *Positions) RequireDomainHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		domain := request.URL.Query().Get("domain")
		if domain == "" {
			p.logger.Println("Empty domain in request")
			// TODO: Send info about error in headers
			http.Error(writer, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			request.Body.Close()
			return
		}

		request = request.WithContext(
			context.WithValue(request.Context(), "domain", domain),
		)

		next.ServeHTTP(writer, request)
	})
}

func (p *Positions) HandleSummary(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	domain := request.URL.Query().Get("domain")

	summary, err := p.Summary(domain)
	if err != nil {
		// TODO: Send info about error in headers
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	resultMap := map[string]interface{}{"domain": domain, "summary": summary}
	encodedResult, err := json.Marshal(resultMap)
	if err != nil {
		p.logger.Println(err)
		// TODO: Send info about error in headers
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "%s", encodedResult)
}

func (p *Positions) HandlePositions(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	domain := request.URL.Query().Get("domain")

	positions, err := p.Positions(domain, 10, 0)
	if err != nil {
		p.logger.Println(err)
		// TODO: Send info about error in headers
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	resultMap := map[string]interface{}{"domain": domain, "positions": positions}
	encodedResult, err := json.Marshal(resultMap)
	if err != nil {
		p.logger.Println(err)
		// TODO: Send info about error in headers
		http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(200)
	fmt.Fprintf(writer, "%s", encodedResult)
}
