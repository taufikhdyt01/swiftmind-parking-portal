package payment

import "github.com/google/uuid"

// Scenario values exercised from the UI (test-only inputs).
const (
	ScenarioSuccess = "success"
	ScenarioFailed  = "failed"
)

// Payment outcome statuses.
const (
	StatusPaid   = "paid"
	StatusFailed = "failed"
)

// ChargeResult mirrors the mocked provider's response:
//
//	PaymentService.charge(invoice_id, amount, scenario) -> { status, transaction_id }
type ChargeResult struct {
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id"`
}

// Charge is the mocked payment provider. The scenario is a test-only argument
// used to simulate the outcome: "success" -> paid, anything else -> failed.
func Charge(invoiceID string, amount int64, scenario string) ChargeResult {
	txn := "txn_" + uuid.NewString()
	if scenario == ScenarioSuccess {
		return ChargeResult{Status: StatusPaid, TransactionID: txn}
	}
	return ChargeResult{Status: StatusFailed, TransactionID: txn}
}
