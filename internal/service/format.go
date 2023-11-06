package service

func FormatAddress(addr string) string {
	if len(addr) <= 10 {
		return addr
	}
	start := addr[:6]
	end := addr[len(addr)-4:]
	return start + "..." + end
}
