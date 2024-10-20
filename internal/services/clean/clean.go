package clean

func CleanHahsedUrl(hashedUrl []byte) string {
	var res string
	for i := 0; i < len(hashedUrl); i++ {
		if string(hashedUrl[i]) != "/" {
			res += string(hashedUrl[i])
		}
	}

	return res
}
