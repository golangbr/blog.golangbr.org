package main

import (
	"code.google.com/p/goauth2/oauth"
	//	"encoding/xml"
	"fmt"
	"github.com/google/go-github/github"
	rss "github.com/maiconio/go-pkg-rss"
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
	//3.1[OK]	parsear entradas com a lib go-pkg-rss - suporte a RSS e ATOM!!!
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

		feed := rss.New(5, true, nil, nil)
		feed.Fetch(urlBlog, nil)

		if len(feed.Channels) > 0 {
			c := feed.Channels[0]
			for i := 0; i < len(c.Items); i++ {
				dataPublicacao, _ := c.Items[i].ParsedPubDate()
				d := dataPublicacao.UTC().Format(time.RFC3339Nano)
				d = d[0:4] + d[5:7] + d[8:10] + d[11:13] + d[14:16]

				if d > ultimaLeitura {
					entrada := c.Items[i]
					ehGolang := false
					for j := 0; j < len(entrada.Categories); j++ {
						if strings.Contains(strings.ToLower(entrada.Categories[j].Text), "golang") {
							ehGolang = true
						}
					}

					if ehGolang {
						nomeArquivo := dataPublicacao.UTC().Format(time.RFC3339Nano)[0:10] + "-" + url.QueryEscape(entrada.Title) + ".md"
						conteudo := "---\n"
						conteudo = conteudo + "layout: default\n"
						conteudo = conteudo + "title: " + entrada.Title + "\n"
						conteudo = conteudo + "---\n"
						conteudo = conteudo + entrada.Content.Text

						comitaArquivo(nomeArquivo, conteudo)
					}
				}
			}
		}
	}
	w.Done()
}

func listaBlogs() []string {
	resp, err := http.Get("https://github.com/maiconio/blog.golangbr.org/blob/master/_BLOGS")
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
		Token: &oauth.Token{AccessToken: "CHAVE SUPER SECRETA AQUI"},
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
