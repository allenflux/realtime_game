package domain

import (
	"github.com/shopspring/decimal"
)

// ParseAmountToDB 把用户传入金额转换成数据库金额单位。
func ParseAmountToDB(amount string) (int64, error) {
	dec, err := decimal.NewFromString(amount)
	if err != nil {
		return 0, err
	}
	return dec.Truncate(4).Mul(AmountScaleDec).IntPart(), nil
}

// ParseUserMultipleToBetField 把用户传入倍数转换为 bet 表存储格式。
// 旧系统里 bet 的自动 / 手动兑现倍数都按：真实倍数 * 100 * 10000 存储。
func ParseUserMultipleToBetField(multiple string) (int64, error) {
	dec, err := decimal.NewFromString(multiple)
	if err != nil {
		return 0, err
	}
	return dec.Mul(MultipleScaleDec).Mul(MultipleTailDec).Ceil().IntPart(), nil
}

// CurrentMultipleToBetField 把当前局倍数(*100)转换成 bet 字段使用的格式。
func CurrentMultipleToBetField(currentMultiple int64) int64 {
	return currentMultiple * MultipleTail
}

// BetFieldToHumanMultiple 把 bet 字段中的倍数换算成字符串。
func BetFieldToHumanMultiple(value int64) string {
	return decimal.NewFromInt(value).Div(MultipleScaleDec).Div(MultipleTailDec).String()
}

// DBAmountToString 把数据库金额转成可用于外部接口的字符串。
func DBAmountToString(value int64) string {
	return decimal.NewFromInt(value).Div(AmountScaleDec).Truncate(4).String()
}
