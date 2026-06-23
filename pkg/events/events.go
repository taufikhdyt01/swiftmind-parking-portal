// Package events defines the async event contracts shared between services so
// producers and consumers agree on routing keys and payload shapes.
package events

import "time"

// Topic exchange and routing keys for the RabbitMQ event bus.
const (
	Exchange = "swiftmind.events"

	RoutingViolationCreated = "violation.created"
	RoutingPaymentCompleted = "payment.completed"
)

// ViolationCreated is published when an officer issues a violation. The payment
// service consumes it to create an invoice.
type ViolationCreated struct {
	ViolationID   string    `json:"violation_id"`
	Plate         string    `json:"plate"`
	ViolationType string    `json:"violation_type"`
	OwnerEmail    string    `json:"owner_email"`
	IssuedByEmail string    `json:"issued_by_email"`
	FinalAmount   int64     `json:"final_amount"`
	CreatedAt     time.Time `json:"created_at"`
}

// PaymentCompleted is published when an invoice is paid. The violation service
// marks the violation paid; the notification service notifies the member and the
// officer who issued the violation.
type PaymentCompleted struct {
	InvoiceID     string    `json:"invoice_id"`
	ViolationID   string    `json:"violation_id"`
	Plate         string    `json:"plate"`
	OwnerEmail    string    `json:"owner_email"`
	IssuedByEmail string    `json:"issued_by_email"`
	Amount        int64     `json:"amount"`
	TransactionID string    `json:"transaction_id"`
	PaidAt        time.Time `json:"paid_at"`
}
