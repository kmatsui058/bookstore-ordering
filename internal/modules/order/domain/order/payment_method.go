package order

// PaymentMethod は、顧客が代金を支払う方法（支払方法。LikeC4 ビュー bkst_ordr_ordm_pmt）を表す列挙型。
// 値は契約リポジトリのモデルの列挙値と同じ文字列にする。
type PaymentMethod string

const (
	// PaymentMethodCreditCard は、クレジットカード。
	PaymentMethodCreditCard PaymentMethod = "CREDIT_CARD"
	// PaymentMethodBankTransfer は、銀行振込。
	PaymentMethodBankTransfer PaymentMethod = "BANK_TRANSFER"
	// PaymentMethodCashOnDelivery は、代金引換（配送業者が届けるときに代金を受け取る）。
	PaymentMethodCashOnDelivery PaymentMethod = "CASH_ON_DELIVERY"
)

// AllPaymentMethods は、支払方法のすべての値を返す。値を足したらここにも足す。
func AllPaymentMethods() []PaymentMethod {
	return []PaymentMethod{PaymentMethodCreditCard, PaymentMethodBankTransfer, PaymentMethodCashOnDelivery}
}

// String は、支払方法の値を返す。
func (p PaymentMethod) String() string {
	return string(p)
}

// isValid は、定義された支払方法の値かどうかを返す。
func (p PaymentMethod) isValid() bool {
	switch p {
	case PaymentMethodCreditCard, PaymentMethodBankTransfer, PaymentMethodCashOnDelivery:
		return true
	}
	return false
}
