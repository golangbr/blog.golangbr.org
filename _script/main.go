package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	//"github.com/google/go-github/github"
)

type XMLFeed struct {
	Entradas []Entrada `xml:"entry"`
}

type Entrada struct {
	Id             string `xml:"id"`
	Conteudo       string `xml:"content"`
	Titulo         string `xml:"title"`
	DataPublicacao string `xml:"published"`
	Categorias             []Categoria `xml:"category"`
}

type Categoria struct {
	Termo string `xml:"term,attr"`
}

func (x XMLFeed) String() string {
	return fmt.Sprintf("%s", x.Entradas)
}

func (e Entrada) String() string {
	return fmt.Sprintf("%s", e.Conteudo)
}

func main() {
	//1- [OK]	olha arquivo texto no repositório do blog com lista de blogs
	//2- [OK]	iniciar uma goroutine para olhar se o blog possui feeds
	//3- [OK]	caso possua, parseia entradas 
	//3.1[OK]	e olha se a última possui a tag golang (apenas a última para simplificar)
	//3.2[TODO]	criar estrutura pra controlar a data e hora da ultima execução. 
	//		Parsear apenas entradas que tenham data de publicação superior a esta marca.
	//4- [OK] 	caso possua a tag transforma em markdown com o post e informações 
	//4.1[TODO]	e "commita" no diretório de posts
	//5- [TODO] 	committ dispara o deploy automatizado que atualiza o blog
	
	listaBlogs := listaBlogs()

	var w sync.WaitGroup
	w.Add(len(listaBlogs))

	for i := 0; i < len(listaBlogs); i++ {
		go gravaUltimaEntrada(listaBlogs[i], &w)
	}

	w.Wait()
	fmt.Println("pronto");
}



func gravaUltimaEntrada(urlBlog string, w *sync.WaitGroup) {
	if len(urlBlog) > 0 {
		fmt.Println("Processando o blog: " + urlBlog);
		resp, err := http.Get(urlBlog)
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

			ultimaEntrada := x.Entradas[0]

			ehGolang := false;
			for i := 0; i < len(ultimaEntrada.Categorias); i++ {
				if strings.Contains(strings.ToLower(ultimaEntrada.Categorias[i].Termo), "golang") {
					ehGolang = true;
				}
			}
		
			if ehGolang {
				nomeArquivo := ultimaEntrada.DataPublicacao[0:10] + "-" + url.QueryEscape(ultimaEntrada.Titulo) + ".md"
				postMarkdown, err := os.Create("../_posts/" + nomeArquivo)
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
	}
	w.Done()
}

func listaBlogs() []string {
	resp, err := http.Get("https://raw.githubusercontent.com/maiconio/blog.golangbr.org/master/_BLOGS")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer resp.Body.Close()
		bytes, _ := ioutil.ReadAll(resp.Body)
		blogs := strings.Split(string(bytes), "\n")
		return blogs
	}

	return nil
}
