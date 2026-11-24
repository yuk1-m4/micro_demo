// Package fx は外貨建て取引を1件ずつ基準通貨（円）に換算する関数を提供する。
//
// 一般的な会計ソフトと同じ方式をとる。
//   - 取引ごとに換算し、その場で1回だけ丸めて端数を確定させる。
//   - 外貨の金額とレートは取引の付加情報として持つ。
//   - 決済時・期末評価時の差額は為替差損益として出す。
//
// 金額は最小単位の整数で持ち、float は使わない。
// レートは10進数の文字列から有理数（math/big.Rat）として読み込み、誤差なく扱う。
package fx

import (
	"errors"
	"fmt"
	"math/big"
	"time"
)

// Currency は通貨を表す。MinorUnits は最小単位の桁数（JPY は 0、USD は 2）。
type Currency struct {
	Code       string
	MinorUnits int
}

var (
	JPY = Currency{Code: "JPY", MinorUnits: 0}
	USD = Currency{Code: "USD", MinorUnits: 2}
	EUR = Currency{Code: "EUR", MinorUnits: 2}
)

// Money は通貨と組にした金額。Amount は最小単位の整数（USD なら セント）。
type Money struct {
	Amount   int64
	Currency Currency
}

func (m Money) String() string {
	return fmt.Sprintf("%d %s(minor)", m.Amount, m.Currency.Code)
}

// RateKind はレートの種類。
type RateKind string

const (
	TTM RateKind = "TTM" // 仲値
	TTS RateKind = "TTS" // 対顧客電信売相場
	TTB RateKind = "TTB" // 対顧客電信買相場
)

// Rate は為替レート。1 From = Value To を表す（例: 1 USD = 150.25 JPY）。
type Rate struct {
	From  Currency
	To    Currency
	Value *big.Rat
	Kind  RateKind
	Date  time.Time
}

// NewRate は10進数の文字列からレートを作る。float を経由しないので誤差が入らない。
func NewRate(from, to Currency, value string, kind RateKind, date time.Time) (Rate, error) {
	v, ok := new(big.Rat).SetString(value)
	if !ok {
		return Rate{}, fmt.Errorf("fx: invalid rate %q", value)
	}
	if v.Sign() <= 0 {
		return Rate{}, fmt.Errorf("fx: rate must be positive: %q", value)
	}
	return Rate{From: from, To: to, Value: v, Kind: kind, Date: date}, nil
}

// RoundingMode は丸めの方法。呼び出し側が必ず明示する。
type RoundingMode int

const (
	RoundHalfUp   RoundingMode = iota + 1 // 四捨五入（0から遠い方へ）
	RoundHalfEven                         // 銀行丸め
	RoundDown                             // 切り捨て（0に近い方へ）
)

var (
	ErrCurrencyMismatch = errors.New("fx: currency mismatch")
	ErrUnknownRounding  = errors.New("fx: unknown rounding mode")
	ErrOverflow         = errors.New("fx: amount overflows int64")
)

// Convert は外貨の金額をレートで換算し、最後に1回だけ丸める。
//
//	換算後(最小単位) = 元の最小単位 / 10^元の桁数 × レート × 10^先の桁数
func Convert(m Money, r Rate, mode RoundingMode) (Money, error) {
	if m.Currency != r.From {
		return Money{}, fmt.Errorf("%w: money is %s, rate is from %s", ErrCurrencyMismatch, m.Currency.Code, r.From.Code)
	}
	x := new(big.Rat).SetInt64(m.Amount)
	x.Mul(x, r.Value)
	x.Mul(x, pow10Rat(r.To.MinorUnits-m.Currency.MinorUnits))

	n, err := round(x, mode)
	if err != nil {
		return Money{}, err
	}
	if !n.IsInt64() {
		return Money{}, ErrOverflow
	}
	return Money{Amount: n.Int64(), Currency: r.To}, nil
}

// Transaction は1件の外貨建て取引。円換算額で記帳し、外貨の金額とレートは付加情報として持つ。
type Transaction struct {
	Foreign Money // 外貨の金額（付加情報）
	Rate    Rate  // 取引時のレート（付加情報）
	Booked  Money // 記帳した円換算額（確定値）
}

// Book は取引を取引時レートで換算して記帳する。端数はここで確定する。
func Book(foreign Money, rate Rate, mode RoundingMode) (Transaction, error) {
	booked, err := Convert(foreign, rate, mode)
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{Foreign: foreign, Rate: rate, Booked: booked}, nil
}

// Side は外貨建ての資産か負債か。差損益の符号が逆になる。
type Side int

const (
	Asset     Side = iota + 1 // 外貨建て債権・預金など（円安で益）
	Liability                 // 外貨建て債務など（円安で損）
)

// Settle は取引を決済時レートで換算し直し、為替差損益を返す。
// 正の値が差益、負の値が差損。
func Settle(tx Transaction, settleRate Rate, side Side, mode RoundingMode) (settled Money, gainLoss Money, err error) {
	settled, err = Convert(tx.Foreign, settleRate, mode)
	if err != nil {
		return Money{}, Money{}, err
	}
	gainLoss, err = diff(settled, tx.Booked, side)
	if err != nil {
		return Money{}, Money{}, err
	}
	return settled, gainLoss, nil
}

// Revalue は期末に残っている外貨残高を期末レートで換算し直し、帳簿上の円換算額との差を為替差損益として返す。
// 正の値が差益、負の値が差損。
func Revalue(balance Money, bookedBase Money, closingRate Rate, side Side, mode RoundingMode) (revalued Money, gainLoss Money, err error) {
	revalued, err = Convert(balance, closingRate, mode)
	if err != nil {
		return Money{}, Money{}, err
	}
	gainLoss, err = diff(revalued, bookedBase, side)
	if err != nil {
		return Money{}, Money{}, err
	}
	return revalued, gainLoss, nil
}

func diff(now, booked Money, side Side) (Money, error) {
	if now.Currency != booked.Currency {
		return Money{}, fmt.Errorf("%w: %s vs %s", ErrCurrencyMismatch, now.Currency.Code, booked.Currency.Code)
	}
	d := now.Amount - booked.Amount
	if side == Liability {
		d = -d
	}
	return Money{Amount: d, Currency: now.Currency}, nil
}

func pow10Rat(exp int) *big.Rat {
	p := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(abs(exp))), nil)
	if exp >= 0 {
		return new(big.Rat).SetInt(p)
	}
	return new(big.Rat).SetFrac(big.NewInt(1), p)
}

// round は有理数を整数に丸める。
func round(x *big.Rat, mode RoundingMode) (*big.Int, error) {
	num := new(big.Int).Set(x.Num())
	den := x.Denom() // 常に正
	neg := num.Sign() < 0
	num.Abs(num)

	q, r := new(big.Int).QuoRem(num, den, new(big.Int))
	if r.Sign() != 0 {
		// 端数と 0.5 の比較: 2r と den を比べる
		cmp := new(big.Int).Lsh(r, 1).Cmp(den)
		switch mode {
		case RoundDown:
			// そのまま
		case RoundHalfUp:
			if cmp >= 0 {
				q.Add(q, big.NewInt(1))
			}
		case RoundHalfEven:
			if cmp > 0 || (cmp == 0 && q.Bit(0) == 1) {
				q.Add(q, big.NewInt(1))
			}
		default:
			return nil, ErrUnknownRounding
		}
	} else if mode != RoundDown && mode != RoundHalfUp && mode != RoundHalfEven {
		return nil, ErrUnknownRounding
	}
	if neg {
		q.Neg(q)
	}
	return q, nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
