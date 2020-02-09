package positions

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

const ID = "positions"

type Service struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewService(db *sql.DB, router *http.ServeMux, logger *logrus.Logger) (*Service, error) {
	p := &Service{
		db:     db,
		logger: logger,
	}
	if err := p.registerHttpRoutes(router); err != nil {
		return nil, err
	}
	return p, nil
}

// Return sum of positions by domain
func (s *Service) Summary(domain string) (int64, error) {
	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctxt, "SELECT SUM(position) as summary FROM positions WHERE domain = ?", domain)
	if row == nil {
		return 0, nil
	}

	var summary sql.NullInt64
	err := row.Scan(&summary)
	if err != nil {
		return 0, err
	}
	if !summary.Valid {
		return 0, err
	}

	return summary.Int64, nil
}

// Return list of positions by domain
func (s *Service) Positions(domain string, limit, offset int, orderBy string) ([]*Position, error) {
	if orderBy == "" {
		orderBy = "volume"
	}

	query := fmt.Sprintf(`SELECT keyword, position, url, volume, results, cpc, updated
		FROM positions
		WHERE domain = ?
		ORDER BY %s ASC
		LIMIT ?
		OFFSET ?`,
		orderBy,
	)

	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctxt, query, domain, limit, offset)
	if err != nil {
		if err == sql.ErrNoRows {
			return make([]*Position, 0, 0), nil
		}
		return nil, err
	}
	defer rows.Close()

	result := make([]*Position, 0, limit)
	for rows.Next() {
		pos := &Position{}
		err = rows.Scan(
			&pos.Keyword,
			&pos.Position,
			&pos.Url,
			&pos.Volume,
			&pos.Results,
			&pos.Cpc,
			&pos.Updated,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, pos)
	}

	return result, nil
}

func checkOrder(orderBy string) error {
	// TODO: generate fields list
	for _, field := range []string{"", "volume", "results", "updated", "cpc", "url", "position", "keyword"} {
		if field == orderBy {
			return nil
		}
	}

	return fmt.Errorf("Unknown field '%s' in order", orderBy)
}
