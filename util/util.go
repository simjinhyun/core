package util

func EncodeToBase62(num uint64) string {
	if num == 0 {
		return "0"
	}

	const s = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var buf [15]byte // log62(2^64) ≈ 11.8, 여유 있게 15
	i := len(buf)

	for num > 0 {
		i--
		buf[i] = s[num%62]
		num /= 62
	}

	return string(buf[i:])
}
