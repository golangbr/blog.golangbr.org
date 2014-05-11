package main

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/xml"
	"fmt"
	"github.com/google/go-github/github"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type XMLFeed struct {
	Entradas []Entrada `xml:"entry"`
}

type Entrada struct {
	Id             string      `xml:"id"`
	Conteudo       string      `xml:"content"`
	Titulo         string      `xml:"title"`
	DataPublicacao string      `xml:"published"`
	Categorias     []Categoria `xml:"category"`
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
	//3.1[10%]	parsear entradas com a lib github.com/SlyMarbo/rss
	//3.2[OK]	e olha se a última possui a tag golang (apenas a última para simplificar)
	//3.3[OK]	criar estrutura pra controlar a data e hora da ultima execução.
	//		Parsear apenas entradas que tenham data de publicação superior a esta marca.
	//4- [OK] 	caso possua a tag transforma em markdown com o post e informações
	//4.1[OK]	e "commita" no diretório de posts
	//5- [OK] 	committ dispara o deploy automatizado que atualiza o blog

	listaBlogs := listaBlogs()
	ultimaLeitura := lerUltimaLeitura()

	var w sync.WaitGroup
	w.Add(len(listaBlogs))

	for i := 0; i < len(listaBlogs); i++ {
		go gravaUltimasEntradas(listaBlogs[i], &w, ultimaLeitura)
	}

	w.Wait()
	escreverUltimaLeitura()
	fmt.Println("pronto")
}

//trocar p/ gravar as ultimas 5 entradas.
func gravaUltimasEntradas(urlBlog string, w *sync.WaitGroup, ultimaLeitura string) {
	if len(urlBlog) > 0 {
		fmt.Println("Processando o blog: " + urlBlog)
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

			for i := 0; i < len(x.Entradas); i++ {
				d := x.Entradas[i].DataPublicacao

				var f string
				switch {
				case strings.HasSuffix(strings.ToUpper(d), "Z"):
					f = "2006-01-02T15:04:05Z"
				default:
					f = "2006-01-02T15:04:05-07:00"
				}
				t, _ := time.Parse(f, d)
				d = t.UTC().Format(time.RFC3339Nano)
				d = d[0:4] + d[5:7] + d[8:10] + d[11:13] + d[14:16]

				fmt.Println(d)
				fmt.Println(ultimaLeitura)

				if d > ultimaLeitura {
					entrada := x.Entradas[i]

					ehGolang := false
					for i := 0; i < len(entrada.Categorias); i++ {
						if strings.Contains(strings.ToLower(entrada.Categorias[i].Termo), "golang") {
							ehGolang = true
						}
					}

					if ehGolang {
						nomeArquivo := entrada.DataPublicacao[0:10] + "-" + url.QueryEscape(entrada.Titulo) + ".md"
						conteudo := "---\n"
						conteudo = conteudo + "layout: default\n"
						conteudo = conteudo + "title: " + entrada.Titulo + "\n"
						conteudo = conteudo + "---\n"
						conteudo = conteudo + entrada.Conteudo

						comitaArquivo(nomeArquivo, conteudo)
					}
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

func lerUltimaLeitura() string {
	dat, _ := ioutil.ReadFile("ultimaLeitura")
	return string(dat)
}

func escreverUltimaLeitura() {
	f, err := os.Create("ultimaLeitura")
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		n := time.Now().UTC().Format(time.RFC3339Nano)
		n = n[0:4] + n[5:7] + n[8:10] + n[11:13] + n[14:16]
		io.WriteString(f, n)
		f.Close()
	}
}

func comitaArquivo(nomeArquivo, conteudo string) {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: ""},
	}

	if t != nil {
		client := github.NewClient(t.Client())

		arquivoPost, _, _, _ := client.Repositories.GetContents("maiconio", "blog.golangbr.org", "_posts/"+nomeArquivo, &github.RepositoryContentGetOptions{})

		if arquivoPost == nil {
			opt := &github.CommitsListOptions{}
			commits, _, err := client.Repositories.ListCommits("maiconio", "blog.golangbr.org", opt)
			if err == nil {
				menssagem := "Adicionando post: " + nomeArquivo
				bConteudo := []byte(conteudo)

				repositoryContentsOptions := &github.RepositoryContentFileOptions{
					Message: &menssagem,
					Content: bConteudo,
					SHA:     commits[0].SHA, //
					Committer: &github.CommitAuthor{
						Name: github.String("maiconio"), Email: github.String("maiconscosta@gmail.com")},
				}

				_, _, err = client.Repositories.CreateFile("maiconio", "blog.golangbr.org", "_posts/"+nomeArquivo, repositoryContentsOptions)
				if err != nil {
					fmt.Printf("Erro ao efetuar o commit do arquivo: %v", err)
				} else {
					fmt.Println(nomeArquivo + " gravado.")
				}
			} else {
				fmt.Printf("Erro ao obter a lista de commits: %v", err)
			}
		} else {
			fmt.Printf("Arquivo " + nomeArquivo + " já existe")
		}
	} else {
		fmt.Println("Não foi possível inicializar o client o github")
	}
}
