package torrnado

import (
	"fmt"
	"math/rand"
	"time"

	"golang.org/x/text/encoding/charmap"
)

func Operror(op string, err error) error {
	return fmt.Errorf("%s: %s", op, err)
}

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func ConjureFluckyVerse(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func ISaidNow() string {
	customFormat := "02-01-2006 15:04:05"
	return time.Now().Format(customFormat)
	// return time.Now().Format(time.RFC3339)
}

func ConjureTopicPlan(start, count int) TopicPlan {

	plan := make(TopicPlan, count)
	for i := range count {
		plan[start+i] = ISaidNow()
	}
	return plan
}

func AccidentalPeriodSec(min, max int) time.Duration {
	randomSeconds := rand.Intn(max-min) + min
	return time.Duration(randomSeconds) * time.Second
}

func PlotDeal[T, M any](s []T, f func(T) M) []M {
	result := make([]M, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

func DecodeWin1251(s string) string {
	dec := charmap.Windows1251.NewDecoder()
	out, _ := dec.Bytes([]byte(s))
	return string(out)
}

func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
