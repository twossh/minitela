package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/twossh/minitela/internal/r15m"
)

func main() {
	page := flag.Int(
		"page",
		5,
		"página do R15M entre 1 e 7",
	)

	flag.Parse()

	if *page < 1 || *page > 7 {
		fmt.Fprintln(
			os.Stderr,
			"Erro: página deve estar entre 1 e 7",
		)

		os.Exit(2)
	}

	fmt.Println(
		"=== MiniTela - Page Test ===",
	)

	fmt.Printf(
		"Página desejada: %d\n",
		*page,
	)

	fmt.Println()
	fmt.Println(
		"Conectando ao R15M...",
	)

	conn, err := r15m.Connect()

	if err != nil {
		fail(err)
	}

	defer conn.Close()

	fmt.Printf(
		"Dispositivo : %s\n",
		conn.Device.Path,
	)

	fmt.Println(
		"Handshake   : OK",
	)

	fmt.Printf(
		"Selecionando página %d...\n",
		*page,
	)

	actual, err :=
		conn.WriteRegisterVerified(
			r15m.RegisterCurrentPage,
			uint32(*page),
		)

	if err != nil {
		fail(err)
	}

	fmt.Printf(
		"Página      : %d\n",
		actual,
	)

	fmt.Println(
		"Status      : OK",
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
