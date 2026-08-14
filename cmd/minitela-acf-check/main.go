package main

import (
	"fmt"
	"os"

	"github.com/twossh/minitela/internal/acf"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println(
			"Uso: MiniTelaACFCheck arquivo.acf",
		)
		os.Exit(2)
	}

	path := os.Args[1]

	data, err :=
		os.ReadFile(path)

	if err != nil {
		fail(err)
	}

	stored, err :=
		acf.StoredChecksum(data)

	if err != nil {
		fail(err)
	}

	calculated :=
		acf.ComputeChecksum(data)

	fmt.Println(
		"=== MiniTela ACF Check ===",
	)

	fmt.Println()

	fmt.Printf(
		"Arquivo     : %s\n",
		path,
	)

	fmt.Printf(
		"Tamanho     : %d bytes\n",
		len(data),
	)

	fmt.Printf(
		"Armazenado  : 0x%08X\n",
		stored,
	)

	fmt.Printf(
		"Calculado   : 0x%08X\n",
		calculated,
	)

	if err :=
		acf.ValidateChecksum(
			data,
		); err != nil {

		fmt.Printf(
			"Validação   : ERRO - %v\n",
			err,
		)

		os.Exit(1)
	}

	fmt.Println(
		"Validação   : OK",
	)
}

func fail(err error) {
	fmt.Fprintf(
		os.Stderr,
		"Erro: %v\n",
		err,
	)

	os.Exit(1)
}
