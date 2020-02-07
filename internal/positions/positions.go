package positions

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type Positions struct {
	db        *sql.DB
	logger    *log.Logger
	summary   *sql.Stmt
	positions *sql.Stmt
}

func NewPositions(db *sql.DB, logger *log.Logger) *Positions {
	p := &Positions{
		db:        db,
		logger:    logger,
		summary:   nil,
		positions: nil,
	}
	return p
}

func (p *Positions) Prepare() error {
	var err error
	p.summary, err = p.db.Prepare("SELECT position FROM positions WHERE domain = :domain")
	if err != nil {
		p.logger.Println(err)
		return err
	}

	p.positions, err = p.db.Prepare("SELECT keyword, position, url, volume, results, cpc, updated FROM positions WHERE domain = :domain LIMIT :l OFFSET :o")
	if err != nil {
		p.logger.Println(err)
		return err
	}

	return nil
}

func (p *Positions) Summary(domain string) (uint64, error) {
	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rows, err := p.summary.QueryContext(ctxt, domain)
	if err != nil {
		p.logger.Println(err)
		return 0, err
	}
	defer rows.Close()
	var position, summary uint64
	for rows.Next() {
		err = rows.Scan(&position)
		if err != nil {
			p.logger.Println(err)
			return 0, err
		}
		summary += position
	}

	return summary, nil
}

func (p *Positions) Positions(domain string, limit, offset int) ([]*Position, error) {
	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	rows, err := p.positions.QueryContext(ctxt, domain, limit, offset)
	if err != nil {
		p.logger.Println(err)
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
			p.logger.Println(err)
			return nil, err
		}
		result = append(result, pos)
	}

	return result, nil
}
