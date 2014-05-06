package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type XMLFeed struct {
	Entradas []Entrada `xml:"entry"`
}

type Entrada struct {
	Id             string `xml:"id"`
	Conteudo       string `xml:"content"`
	Titulo         string `xml:"title"`
	DataPublicacao string `xml:"published"`
}

func (x XMLFeed) String() string {
	return fmt.Sprintf("%s", x.Entradas)
}

func (e Entrada) String() string {
	return fmt.Sprintf("%s", e.Conteudo)
}

func main() {
        //1- [TODO] olha arquivo texto no repositório do blog com lista de blogs
        //2- [TODO] iniciar uma goroutine para olhar se o blog possui feeds
        //3- [OK]   caso possua, parseia entradas e olha se a última possui a tag golang (apenas a última para simplificar)
        //4- [OK]   caso possua a tag transforma em markdown com o post e informações e "commita" no diretório de posts
        //5- [TODO] commit dispara o deploy automatizado que atualiza o blog

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

		ultimaEntrada := x.Entradas[len(x.Entradas)-1]

		nomeArquivo := ultimaEntrada.DataPublicacao[0:10] + "-" + url.QueryEscape(ultimaEntrada.Titulo) + ".md"
		postMarkdown, err := os.Create("../_posts/"+nomeArquivo)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		} else {

			io.WriteString(postMarkdown, "---\n")
			io.WriteString(postMarkdown, "layout: default\n")
			io.WriteString(postMarkdown, "title: "+ultimaEntrada.Titulo+"\n")
			io.WriteString(postMarkdown, "---\n")
			io.WriteString(postMarkdown, ultimaEntrada.Conteudo)
			postMarkdown.Close()
		}

	}

}
