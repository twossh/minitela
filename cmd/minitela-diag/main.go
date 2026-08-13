package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/twossh/minitela/internal/r15m"
)

func main() {
	reg := flag.Int(
		"read-string",
		-1,
		"registrador de string para leitura",
	)

	maxLen := flag.Int(
		"max-len",
		64,
		"tamanho máximo da string",
	)

	flag.Parse()

	if *reg < 0 ||
		*reg > 65535 ||
		*maxLen < 1 ||
		*maxLen > 1024 {
		fmt.Fprintln(
			os.Stderr,
			"parâmetros inválidos",
		)
		os.Exit(2)
	}

	conn, err := r15m.Connect()
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"conexão: %v\n",
			err,
		)
		os.Exit(1)
	}
	defer conn.Close()

	value, err := conn.ReadStringRegister(
		uint16(*reg),
		uint16(*maxLen),
	)
	if err != nil {
		fmt.Fprintf(
			os.Stderr,
			"leitura: %v\n",
			err,
		)
		os.Exit(1)
	}

	fmt.Printf(
		"reg[%d/0x%04X] = %q\n",
		*reg,
		*reg,
		string(value),
	)
}
