package main

// lol, "crypto"
const initialKey = byte(0xab)

func encrypt(in []byte) []byte {
	out := make([]byte, len(in))
	key := initialKey
	for i := 0; i < len(in); i++ {
		out[i] = in[i] ^ key
		key = out[i]
	}
	return out
}

func decrypt(in []byte) []byte {
	out := make([]byte, len(in))
	key := initialKey
	for i := 0; i < len(in); i++ {
		out[i] = in[i] ^ key
		key = in[i]
	}
	return out
}
