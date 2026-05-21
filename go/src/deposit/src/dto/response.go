package dto

import (
	deposit "interview_mock_deposits_go/deposit/src"
)

type FormattedDeposit struct {
	deposit.Deposit
}

func FormatResponse(d deposit.Deposit) FormattedDeposit {
	return FormattedDeposit{Deposit: d}
}
