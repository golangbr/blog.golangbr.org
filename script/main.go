package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type XMLFeed struct {
	Entradas []Entrada `xml:"entry"`
}

type Entrada struct {
	Id string `xml:"id"`
	Conteudo string `xml:"content"`
}

func (x XMLFeed) String() string {
	return fmt.Sprintf("%s", x.Entradas)
}

func (e Entrada) String() string {
	return fmt.Sprintf("%s", e.Conteudo)
}

func main() {
	//response, err := http.Get("http://maiconio.blogspot.com/feeds/posts/default")
	resp, err := http.Get("http://127.0.0.1/teste/default")

	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		conteudoXML, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		var x XMLFeed
		xml.Unmarshal(conteudoXML, &x)

                ultimaEntrada := x.Entradas[len(x.Entradas)-1:]

		fmt.Printf("\t%s\n", ultimaEntrada)

                //TODO: salva arquivo markdown
	}

}
