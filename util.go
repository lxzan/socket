package socket

type Form map[string]string

func packLength(msg []byte) int {
	var i int
	n := len(msg)
	for i = n - 1; i >= 0; i-- {
		if msg[i] != 0 {
			break
		}
	}
	return i + 1
}


