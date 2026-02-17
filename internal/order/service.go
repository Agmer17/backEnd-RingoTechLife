package order

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type TransactionService struct {
	muTransactionsData sync.Mutex
	transactionsData   map[uuid.UUID]*time.Timer
	transactionRepo    OrderRepositoryInterface
}

func NewTransactionService(orderRepo *OrderRepositoryImpl) *TransactionService {
	return &TransactionService{
		transactionsData: make(map[uuid.UUID]*time.Timer, 0),
		transactionRepo:  orderRepo,
	}
}
