package positions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const PageSize = 10

func (s *Service) RequirementsHandler(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func(logger *logrus.Logger, r *http.Request) {
			if err := r.Body.Close(); err != nil {
				s.logger.Errorln(err)
			}
		}(s.logger, request)

		domain := request.URL.Query().Get("domain")
		if domain == "" {
			s.logger.Warnln("Empty domain in request")
			http.Error(writer, "Domain field MUST be specified", http.StatusBadRequest)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func (s *Service) HandleSummary(writer http.ResponseWriter, request *http.Request) {
	domain := request.URL.Query().Get("domain")

	summary, err := s.Summary(domain)
	if err != nil {
		s.handleServerError(writer, err)
		return
	}

	resultMap := map[string]interface{}{"domain": domain, "positions_count": summary}
	encodedResult, err := json.Marshal(resultMap)

	if err != nil {
		s.handleServerError(writer, err)
		return
	}

	s.responseSuccess(writer, encodedResult)
}

func (s *Service) HandlePositions(writer http.ResponseWriter, request *http.Request) {
	var (
		page = 1
		err  error
	)
	domain := request.URL.Query().Get("domain")

	pageStr := request.URL.Query().Get("page")
	if pageStr != "" {
		if page, err = strconv.Atoi(pageStr); err != nil {
			s.logger.Errorln(err)
			http.Error(writer, "Incorrect page param value", http.StatusBadRequest)
			return
		}
	}

	order := request.URL.Query().Get("orderBy")
	if err = checkOrder(order); err != nil {
		s.logger.Errorln(err)
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	positions, err := s.Positions(domain, PageSize, PageSize*(page-1), order)
	if err != nil {
		s.handleServerError(writer, err)
		return
	}

	resultMap := map[string]interface{}{"domain": domain, "positions": positions}
	encodedResult, err := json.Marshal(resultMap)
	if err != nil {
		s.handleServerError(writer, err)
		return
	}

	s.responseSuccess(writer, encodedResult)
}

func (s *Service) registerHttpRoutes(mux *http.ServeMux) error {
	h := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   "SEM",
			Subsystem:   ID,
			Name:        "request_duration",
			Help:        fmt.Sprintf("Request duration for %s service", ID),
			ConstLabels: nil,
			Buckets:     prometheus.LinearBuckets(0, .05, 20),
		},
		nil,
	)

	mux.Handle("/api/summary", promhttp.InstrumentHandlerDuration(h, s.RequirementsHandler(s.HandleSummary)))
	mux.Handle("/api/positions", promhttp.InstrumentHandlerDuration(h, s.RequirementsHandler(s.HandlePositions)))
	return nil
}

func (s *Service) responseSuccess(w http.ResponseWriter, data []byte) {
	w.WriteHeader(200)
	if _, err := w.Write(data); err != nil {
		s.logger.Errorln(err)
	}
}

func (s *Service) handleServerError(w http.ResponseWriter, err error) {
	s.logger.Errorln(err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
