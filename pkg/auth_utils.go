package pkg

import (
	"errors"
	"fmt"
	"strings"
)

func GetAccessToken(header string) (string, error) {

	if header == "" {
		return "", errors.New("Harap login terlebih dahulu sebelum mengakses fitur ini")
	}
	parts := strings.SplitN(header, " ", 2)

	if len(parts) != 2 || parts[0] != "Bearer" {
		fmt.Println(parts)
		return "", errors.New("Harap login terlebih dahulu sebelum mengakses fitur ini")
	}

	return parts[1], nil

}
