package fx

import (
	"errors"
	"math"
	"testing"
	"time"
)

var day = time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)

func mustRate(t *testing.T, from, to Currency, v string) Rate {
	t.Helper()
	r, err := NewRate(from, to, v, TTM, day)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestConvert(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
		from   Currency
		to     Currency
		rate   string
		mode   RoundingMode
		want   int64
	}{
		{"USD→JPY 割り切れる", 10000, USD, JPY, "150", RoundHalfUp, 15000},     // $100.00 × 150
		{"USD→JPY 四捨五入で切り上げ", 1001, USD, JPY, "150.5", RoundHalfUp, 1507}, // $10.01 × 150.5 = 1506.505
		{"USD→JPY 切り捨て", 1001, USD, JPY, "150.5", RoundDown, 1506},
		{"ちょうど0.5 四捨五入", 1, USD, JPY, "50", RoundHalfUp, 1}, // $0.01 × 50 = 0.5
		{"ちょうど0.5 銀行丸め（偶数へ）", 1, USD, JPY, "50", RoundHalfEven, 0},
		{"ちょうど1.5 銀行丸め（偶数へ）", 3, USD, JPY, "50", RoundHalfEven, 2},
		{"負の金額（返金）四捨五入", -1, USD, JPY, "50", RoundHalfUp, -1},
		{"JPY→USD 桁が増える", 1000, JPY, USD, "0.006655", RoundHalfUp, 666},      // ¥1000 → $6.655 → 666セント
		{"float だと誤差が出る値", 115, USD, JPY, "100", RoundDown, 115},             // float64 では 1.15*100 = 114.999... で切り捨てると 114 になる
		{"長い小数のレート", 12345, EUR, JPY, "163.123456789", RoundHalfEven, 20138}, // 123.45 × 163.123456789 = 20137.5907...
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(Money{tt.amount, tt.from}, mustRate(t, tt.from, tt.to, tt.rate), tt.mode)
			if err != nil {
				t.Fatal(err)
			}
			if got.Amount != tt.want || got.Currency != tt.to {
				t.Fatalf("got %v, want %d %s", got, tt.want, tt.to.Code)
			}
		})
	}
}

func TestConvertErrors(t *testing.T) {
	r := mustRate(t, USD, JPY, "150")
	if _, err := Convert(Money{100, EUR}, r, RoundHalfUp); !errors.Is(err, ErrCurrencyMismatch) {
		t.Errorf("通貨不一致: err = %v", err)
	}
	if _, err := Convert(Money{100, USD}, r, 0); !errors.Is(err, ErrUnknownRounding) {
		t.Errorf("丸め未指定: err = %v", err)
	}
	if _, err := NewRate(USD, JPY, "abc", TTM, day); err == nil {
		t.Error("不正なレート文字列を受け付けた")
	}
	if _, err := NewRate(USD, JPY, "-1", TTM, day); err == nil {
		t.Error("負のレートを受け付けた")
	}
	if _, err := Convert(Money{1 << 62, USD}, mustRate(t, USD, JPY, "1000"), RoundHalfUp); !errors.Is(err, ErrOverflow) {
		t.Errorf("オーバーフロー: err = %v", err)
	}
}

// 売掛金 $1,000.00 を 1USD=150 で計上し、1USD=155.5 で入金された。
func TestSettleAsset(t *testing.T) {
	tx, err := Book(Money{100000, USD}, mustRate(t, USD, JPY, "150"), RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	if tx.Booked != (Money{150000, JPY}) {
		t.Fatalf("booked = %v", tx.Booked)
	}
	settled, gl, err := Settle(tx, mustRate(t, USD, JPY, "155.5"), Asset, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	if settled != (Money{155500, JPY}) || gl != (Money{5500, JPY}) {
		t.Fatalf("settled = %v, gainLoss = %v, want 155500 / +5500", settled, gl)
	}
}

// 買掛金 $1,000.00 を 1USD=150 で計上し、1USD=155.5 で支払った（円安で差損）。
func TestSettleLiability(t *testing.T) {
	tx, _ := Book(Money{100000, USD}, mustRate(t, USD, JPY, "150"), RoundHalfUp)
	_, gl, err := Settle(tx, mustRate(t, USD, JPY, "155.5"), Liability, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	if gl != (Money{-5500, JPY}) {
		t.Fatalf("gainLoss = %v, want -5500", gl)
	}
}

// 取引時の丸めで確定した端数が、決済時の差損益に含まれることを確認する。
// $10.01 × 150.25 = 1504.0025 → 1504 で記帳、$10.01 × 150.25 で決済 → 差は0。
// $10.01 × 150.30 = 1504.503 → 1505 で決済 → 差益1。
func TestSettleRoundingIsAbsorbedIntoGainLoss(t *testing.T) {
	tx, _ := Book(Money{1001, USD}, mustRate(t, USD, JPY, "150.25"), RoundHalfUp)
	_, gl, _ := Settle(tx, mustRate(t, USD, JPY, "150.25"), Asset, RoundHalfUp)
	if gl.Amount != 0 {
		t.Fatalf("同じレートでの決済で差損益 = %v", gl)
	}
	_, gl, _ = Settle(tx, mustRate(t, USD, JPY, "150.30"), Asset, RoundHalfUp)
	if gl.Amount != 1 {
		t.Fatalf("gainLoss = %v, want 1", gl)
	}
}

// 期末に残っている外貨預金 $2,500.00（帳簿上 ¥370,000）を期末レート 1USD=145.2 で評価し直す。
func TestRevalue(t *testing.T) {
	revalued, gl, err := Revalue(Money{250000, USD}, Money{370000, JPY}, mustRate(t, USD, JPY, "145.2"), Asset, RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	if revalued != (Money{363000, JPY}) || gl != (Money{-7000, JPY}) {
		t.Fatalf("revalued = %v, gainLoss = %v, want 363000 / -7000", revalued, gl)
	}
}

func TestRevalueCurrencyMismatch(t *testing.T) {
	_, _, err := Revalue(Money{100, USD}, Money{100, EUR}, mustRate(t, USD, JPY, "150"), Asset, RoundHalfUp)
	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("err = %v", err)
	}
}

// 換算はどちらも成功するが、差損益が int64 に収まらない場合はエラーにする（Codex レビューの指摘）。
func TestSettleGainLossOverflow(t *testing.T) {
	tx, err := Book(Money{math.MinInt64, USD}, mustRate(t, USD, JPY, "100"), RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := Settle(tx, mustRate(t, USD, JPY, "0.000000000000000001"), Asset, RoundHalfUp); !errors.Is(err, ErrOverflow) {
		t.Fatalf("資産の差損益: err = %v", err)
	}
}

// 差が math.MinInt64 のとき、負債として符号を反転するとあふれる。
func TestSettleLiabilityNegationOverflow(t *testing.T) {
	tx := Transaction{Foreign: Money{math.MinInt64, JPY}, Booked: Money{0, JPY}}
	if _, _, err := Settle(tx, mustRate(t, JPY, JPY, "1"), Liability, RoundHalfUp); !errors.Is(err, ErrOverflow) {
		t.Fatalf("負債の符号反転: err = %v", err)
	}
}

// 小数3桁の通貨（クウェート・ディナール）も、最小単位（1/1000 KWD）の整数で同じように扱える。
func TestConvertThreeDecimalCurrency(t *testing.T) {
	KWD := Currency{Code: "KWD", MinorUnits: 3}
	// 12.345 KWD × 487.6543 = 6020.0923335 円 → 6020 円
	got, err := Convert(Money{12345, KWD}, mustRate(t, KWD, JPY, "487.6543"), RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	if got != (Money{6020, JPY}) {
		t.Fatalf("KWD→JPY: got %v", got)
	}
	// 10,000 円 × 0.002051 = 20.51 KWD → 20510（1/1000 KWD）
	got, err = Convert(Money{10000, JPY}, mustRate(t, JPY, KWD, "0.002051"), RoundHalfUp)
	if err != nil {
		t.Fatal(err)
	}
	if got != (Money{20510, KWD}) {
		t.Fatalf("JPY→KWD: got %v", got)
	}
}
