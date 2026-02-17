package order

import "github.com/jackc/pgx/v5/pgxpool"

type OrderRepositoryInterface interface {
	todo()
}

type OrderRepositoryImpl struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(p *pgxpool.Pool) *OrderRepositoryImpl {
	return &OrderRepositoryImpl{
		pool: p,
	}
}

func (o *OrderRepositoryImpl) todo() {

}
