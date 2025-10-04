package service

import (
	payment "ecommerce/models/payments"
	"ecommerce/repository"

	"github.com/jinzhu/gorm"
)

type PaymentService struct {
	db         *gorm.DB
	repository repository.EcommerceRepository
}

func NewPaymentService(db *gorm.DB, repo repository.EcommerceRepository) *PaymentService {
	return &PaymentService{
		db:         db,
		repository: repo,
	}
}

func (s *PaymentService) CreatePayment(newPayment *payment.Payment) error {
	uow := repository.NewUnitOfWork(s.db, false)
	defer uow.RollBack()

	if err := s.repository.Add(uow, newPayment); err != nil {
		return err
	}

	uow.Commit()
	return nil
}

func (s *PaymentService) GetAllPayments(allPayments *[]payment.Payment) error {
	uow := repository.NewUnitOfWork(s.db, true)
	defer uow.RollBack()

	// Just get all payments; UserID is already a column
	err := s.repository.GetAll(uow, allPayments,
		repository.NotDeleted(),
	)
	if err != nil {
		return err
	}

	uow.Commit()
	return nil
}
