package pkg

import "github.com/jinzhu/copier"

func MustCopyValue(to interface{}, from interface{}) {
	err := copier.Copy(to, from)
	if err != nil {
		panic(err)
	}
}

func StringInSlice(str string, list []string) bool {
	for _, val := range list {
		if str == val {
			return true
		}
	}
	return false
}
