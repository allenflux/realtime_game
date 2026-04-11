package common

import (
	"crypto/md5"
	"encoding/base32"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spaolacci/murmur3"
)

const (
	MaxTick       = 1024
	Epoch   int64 = 1711900800000 // 2024-04-01 00:00:00
)

var (
	idMtx       sync.Mutex
	idTick      int64
	idTimestamp int64
	svcNode     int64
)

func GetUUID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func Hash64(data []byte) uint64 {
	return murmur3.Sum64(data)
}

func init() {
	tmp := GetUUID()
	hash := Hash64([]byte(tmp))
	svcNode = int64(hash % 8192)
}

/*
GenerateId  可使用到 2058 年
|---------------64位----------------------|
|-1-|----40---------|----13-----|----10---|
|填充|--时间戳差值-----|-自定填充---|--自增数--|
*/
func GenerateId() int64 {
	idMtx.Lock()
	defer idMtx.Unlock()

RETRY:
	now := time.Now().UnixMilli()
	if idTimestamp == now {
		idTick++
		if idTick > MaxTick {
			time.Sleep(time.Duration(2) * time.Millisecond)
			goto RETRY
		}
	} else {
		idTick = 0
	}
	idTimestamp = now
	return (now-Epoch)<<23 | svcNode<<10 | idTick
}

// bizCode 枚举值为三位数，000-999
type bizCode int16

const (
	SOPDeposit bizCode = 000 // 自营支付充值
	TPDeposit  bizCode = 001 // 三方支付充值

	Promotion bizCode = 010 // 活动金

	AgentCommission bizCode = 020 // 代理返佣
	AgentDividend   bizCode = 021 // 代理占成

	VipBonus     bizCode = 030 // vip晋级奖励
	MemberRebate bizCode = 031 // 会员返水

	AdminDeposit bizCode = 100 // 管理员上分

	NormalWithdraw bizCode = 500 // 常规提现

	AdminWithdraw bizCode = 600 // 管理员下分

	NormalTransfer bizCode = 900 // 常规转账
)

// GenerateOrderId 生成订单号，cmpId 为公司ID，bizCode 为业务类型，uniqueKey 为唯一标识，返回的id为三十四位字符串
func GenerateOrderId(cmpId int64, bizCode bizCode, uniqueKey string) string {
	// 将bizCode转换成两位三十二进制数
	if bizCode < 0 || bizCode > 999 {
		panic("bizCode must be 0-999")
	}
	bizCodeStr := fmt.Sprintf("%02s", strconv.FormatInt(int64(bizCode), 32))

	// 将cmpId转换成六位三十二进制数
	if cmpId < 0 || cmpId > 1073741823 {
		panic("cmpId must be 0-1073741823")
	}
	cmpIdStr := fmt.Sprintf("%06s", strconv.FormatInt(cmpId, 32))

	// 将uniqueKey转换成md5，并编码为二十六位三十二进制数
	hash := md5.Sum([]byte(uniqueKey))
	encoded := base32.StdEncoding.EncodeToString(hash[:])

	return fmt.Sprintf("%s%s%s", strings.ToUpper(bizCodeStr), strings.ToUpper(cmpIdStr), encoded[:26])
}

// GenerateOrderIdWithPrefix 生成订单号，prefix为按公司分配的前缀，最大长度5位，bizCode 为业务类型，uniqueKey 为唯一标识，返回的id为三十四位字符串
func GenerateOrderIdWithPrefix(prefix string, cmpId int64, bizCode bizCode, uniqueKey string) string {
	prefixLen := len(prefix)
	if prefixLen == 0 {
		return GenerateOrderId(cmpId, bizCode, uniqueKey)
	}

	if prefixLen > 5 {
		prefix = prefix[:5]
		prefixLen = 5
	}
	// 将bizCode转换成两位三十二进制数
	if bizCode < 0 || bizCode > 999 {
		panic("bizCode must be 0-999")
	}
	bizCodeStr := fmt.Sprintf("%03s", strconv.FormatInt(int64(bizCode), 32))

	remainLen := 5 - prefixLen
	if remainLen > 0 && remainLen < 5 {
		// 根据len长度生成随机数
		const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		randStr := make([]byte, remainLen)
		for i := 0; i < remainLen; i++ {
			randStr[i] = letters[rand.Intn(len(letters))]
		}
		prefix = prefix + string(randStr[:])
	}

	// 将uniqueKey转换成md5，并编码为二十六位三十二进制数
	hash := md5.Sum([]byte(uniqueKey))
	encoded := base32.StdEncoding.EncodeToString(hash[:])

	return fmt.Sprintf("%s%s%s", prefix, strings.ToUpper(bizCodeStr), encoded[:26])
}
