package domain

import (
	"crypto/sha256"
	"fmt"
	"math"
	mrand "math/rand"
	"strconv"
	"time"
)

// CalcCurrentMultiple 根据飞行开始时间和增长指数计算当前倍数。
// 返回值单位为 *100。
func CalcCurrentMultiple(incNum float64, flyingStartAtMs, nowMs int64) int64 {
	if nowMs <= flyingStartAtMs {
		return MultipleScale
	}
	elapsedSec := float64(nowMs-flyingStartAtMs) / 1000.0
	if elapsedSec < 0 {
		elapsedSec = 0
	}
	value := math.Pow(incNum, elapsedSec) * float64(MultipleScale)
	if value < float64(MultipleScale) {
		value = float64(MultipleScale)
	}
	return int64(math.Floor(value))
}

// CalcCrashDurationMs 反推从起飞到爆炸的时长。
func CalcCrashDurationMs(incNum float64, crashMultiple int64) int64 {
	if incNum <= 1 {
		return 0
	}
	if crashMultiple <= MultipleScale {
		return 0
	}
	target := float64(crashMultiple) / float64(MultipleScale)
	seconds := math.Log(target) / math.Log(incNum)
	if seconds < 0 {
		return 0
	}
	return int64(seconds * 1000)
}

// NextVersion 返回新的版本号。
func NextVersion(v int64) int64 { return v + 1 }

// BuildTermHash 生成局 hash。
func BuildTermHash(termID int64, ctime int64) string {
	raw := strconv.FormatInt(termID, 10) + "-" + strconv.FormatInt(ctime, 10)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
}

// RandCrashMultiple 生成随机爆点，单位 *100。
// 这里沿用旧系统的随机公式，但不再依赖 multiplier-rpc。
func RandCrashMultiple(divisor, ctrlCoef, maxCashoutMultiple int64) int64 {
	if divisor <= 0 {
		return MultipleScale
	}
	if maxCashoutMultiple <= 0 {
		maxCashoutMultiple = 1000
	}
	x := mrand.New(mrand.NewSource(time.Now().UnixNano())).Float64()
	randDiv := mrand.Int63n(divisor) + 1
	if randDiv == divisor {
		return MultipleScale
	}
	k := float64(ctrlCoef)
	if k <= 0 {
		k = 1
	}
	m := (1 - x/k) / (1 - x)
	if m > float64(maxCashoutMultiple) {
		m = float64(maxCashoutMultiple)
	}
	if m < 1 {
		m = 1
	}
	return int64(m * float64(MultipleScale))
}
