package pkg

import (
	"errors"
	"fmt"
	"strings"
)

func GetAccessToken(header string) (string, error) {

	fmt.Println("\n\n\n\n\n header token nya : ", header)

	if header == "" {
		fmt.Println("\n\n\n\n\n kosong headernya : ", header)
		return "", errors.New("Harap login terlebih dahulu sebelum mengakses fitur ini")
	}
	parts := strings.SplitN(header, " ", 2)
	fmt.Println("\n\n\n\n\n parts token nya : ", parts)

	if len(parts) != 2 || parts[0] != "Bearer" {
		fmt.Println("\n\nMasuk ke error karean len bukan 2 dan part[0] bukan bearer ", parts)
		return "", errors.New("Harap login terlebih dahulu sebelum mengakses fitur ini")
	}

	return parts[1], nil

}
