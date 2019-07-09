// Copyright (C) 2019 Storj Labs, Inc.
// See LICENSE for copying information.

package localpayments

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/skyrings/skyring-common/tools/uuid"
	"github.com/zeebo/errs"
	monkit "gopkg.in/spacemonkeygo/monkit.v2"

	"storj.io/storj/satellite/payments"
)

var (
	// creationDate is a Storj creation date.
	creationDate = time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)

	mon = monkit.Package()
)

// StorjCustomer is a predefined customer
// which is linked with every user by default
var storjCustomer = payments.Customer{
	ID:        []byte("0"),
	Name:      "Storj",
	Email:     "storj@mail.test",
	CreatedAt: creationDate,
}

// defaultPaymentMethod represents one and only payment method for local payments,
// which attached to all customers by default
var defaultPaymentMethod = payments.PaymentMethod{
	ID:         []byte("0"),
	CustomerID: []byte("0"),
	Card: payments.Card{
		Country:  "us",
		Brand:    "visa",
		Name:     "Storj Labs",
		ExpMonth: 12,
		ExpYear:  2024,
		LastFour: "3567",
	},
	CreatedAt: creationDate,
}

// internalPaymentsErr is a wrapper for local payments service errors
var internalPaymentsErr = errs.Class("internal payments error")

// service is internal payments.Service implementation
type service struct{}

// NewService create new instance of local payments service
func NewService() payments.Service {
	return &service{}
}

// CreateCustomer creates new payments.Customer with random id to satisfy unique db constraint
func (*service) CreateCustomer(ctx context.Context, params payments.CreateCustomerParams) (_ *payments.Customer, err error) {
	defer mon.Task()(&ctx)(&err)

	var b [8]byte

	_, err = rand.Read(b[:])
	if err != nil {
		return nil, internalPaymentsErr.New("error creating customer")
	}

	return &payments.Customer{
		ID: b[:],
	}, nil
}

// GetCustomer always returns default storjCustomer
func (*service) GetCustomer(ctx context.Context, userID uuid.UUID) (_ *payments.Customer, err error) {
	defer mon.Task()(&ctx)(&err)
	return &storjCustomer, nil
}

// GetCustomerDefaultPaymentMethod always returns defaultPaymentMethod
func (*service) GetCustomerDefaultPaymentMethod(ctx context.Context, customerID []byte) (_ *payments.PaymentMethod, err error) {
	defer mon.Task()(&ctx)(&err)
	return &defaultPaymentMethod, nil
}

// GetCustomerPaymentsMethods always returns payments.Customer list with defaultPaymentMethod
func (*service) GetCustomerPaymentsMethods(ctx context.Context, customerID []byte) (_ []payments.PaymentMethod, err error) {
	defer mon.Task()(&ctx)(&err)
	return []payments.PaymentMethod{defaultPaymentMethod}, nil
}

// GetPaymentMethod always returns defaultPaymentMethod or error
func (*service) GetPaymentMethod(ctx context.Context, id []byte) (_ *payments.PaymentMethod, err error) {
	defer mon.Task()(&ctx)(&err)
	return &defaultPaymentMethod, nil
}

// CreateProjectInvoice creates invoice from provided params
func (*service) CreateProjectInvoice(ctx context.Context, params payments.CreateProjectInvoiceParams) (_ *payments.Invoice, err error) {
	defer mon.Task()(&ctx)(&err)
	return nil, internalPaymentsErr.New("invoice creation is not allowed with local payments")
}

// GetInvoice retrieves invoice information from project invoice stamp by invoice id
// and returns invoice
func (*service) GetInvoice(ctx context.Context, id []byte) (_ *payments.Invoice, err error) {
	defer mon.Task()(&ctx)(&err)
	return nil, internalPaymentsErr.New("invoice creation is not allowed with local payments")
}

// GetProjectInvoices returns nil invoices slice and nil err
func (*service) GetProjectInvoices(ctx context.Context, projectID uuid.UUID) (_ []payments.Invoice, err error) {
	defer mon.Task()(&ctx)(&err)
	return nil, nil
}

// GetProjectInvoiceByStartDate returns an error as no invoice creation is allowed with local payments
func (*service) GetProjectInvoiceByStartDate(ctx context.Context, projectID uuid.UUID, startDate time.Time) (_ *payments.Invoice, err error) {
	defer mon.Task()(&ctx)(&err)
	return nil, internalPaymentsErr.New("invoice creation is not allowed with local payments")
}
