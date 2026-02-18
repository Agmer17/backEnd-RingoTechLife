package payment

type PayementService struct {
	paymentRepo PaymentRepositoryInterface
}

func NewPaymentService(repo *PaymentRepositoryImpl) *PayementService {
	return &PayementService{
		paymentRepo: repo,
	}
}
