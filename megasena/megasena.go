package megasena

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type MegaModalidade struct {
	Numeros int     `json:"numeros"`
	Valor   float64 `json:"valor"`
}

type MegaPremio struct {
	Acertos int    `json:"acertos"`
	Premio  string `json:"premio"`
}

type MegaSetup struct {
	Modalidades      []MegaModalidade `json:"modalidades"`
	Premios          []MegaPremio     `json:"premios"`
	NumerosPossiveis int              `json:"numerospossiveis"`
}

type MegaSena struct {
	UltimoConcurso int         `json:"ultimoconcurso"`
	Estatistica    map[int]int `json:"estatistica"`
	Setup          MegaSetup   `json:"setup"`
}

type ConcursoMega struct {
	DataApuracao                 string   `json:"dataApuracao"`
	DezenasSorteadasOrdemSorteio []string `json:"dezenasSorteadasOrdemSorteio"`
	Numero                       int      `json:"numero"`
}

const (
	URL_MEGASENA_CAIXA_API = "https://servicebus2.caixa.gov.br/portaldeloterias/api/megasena"
	URL_MEGASENA_CAIXA     = "https://loterias.caixa.gov.br/Paginas/Mega-Sena.aspx"

	MEGAJSON = "megasena.json"
)

func (ms *MegaSena) SortKeysDesc() []int {

	keys := make([]int, 0, ms.Setup.NumerosPossiveis)

	for key := range ms.Estatistica {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return ms.Estatistica[keys[i]] > ms.Estatistica[keys[j]]
	})
	return keys
}

func (ms *MegaSena) Aposta(valor float64) (jogos [][]int, rest float64) {
	rest = valor
	if valor < ms.Setup.Modalidades[0].Valor {
		return
	}

	for i := len(ms.Setup.Modalidades) - 1; i >= 0; {
		if rest > ms.Setup.Modalidades[i].Valor {
			jogos = append(jogos, make([]int, ms.Setup.Modalidades[i].Numeros))
			rest -= ms.Setup.Modalidades[i].Valor
		} else {
			i--
		}
	}

	hist := ms.SortKeysDesc()
	fmt.Println(hist)

	return
}

func LoadSetup() (setup MegaSetup, err error) {
	var (
		resp *http.Response
	)
	setup = MegaSetup{NumerosPossiveis: 60}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if resp, err = http.Get(URL_MEGASENA_CAIXA); err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		if index == 0 {
			tablehtml.Find("th").Each(func(indexth int, rowhtml *goquery.Selection) {
				if indexth == 1 {
					rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
						if indexth == 2 {
							fmt.Println(tablecell.Text())
						}
					})
				}
			})

			tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
				if indextr == 1 {
					n := 6
					rowhtml.Find("th").Each(func(indexth int, tablecell *goquery.Selection) {
						setup.Premios = append(setup.Premios, MegaPremio{Acertos: n, Premio: tablecell.Text()})
						n--
					})
				}
				var modalidade int
				var valor float64
				rowhtml.Find("td").Each(func(indextd int, tablecell *goquery.Selection) {
					if indextd == 0 {
						modalidade, _ = strconv.Atoi(tablecell.Text())
					}
					if indextd == 1 {
						txt := strings.ReplaceAll(strings.ReplaceAll(tablecell.Text(), ".", ""), ",", ".")
						valor, _ = strconv.ParseFloat(txt, 64)
						setup.Modalidades = append(setup.Modalidades, MegaModalidade{Numeros: modalidade, Valor: valor})
					}
				})
			})
		}
	})
	return

}

func CreateFactory() (ms *MegaSena, err error) {
	var (
		body       []byte
		cm         *ConcursoMega
		fjson      *os.File
		UltimaMega int
	)
	ms = new(MegaSena)
	ms.Estatistica = make(map[int]int)
	data, _ := ioutil.ReadFile(MEGAJSON)
	if len(data) > 0 {
		if err = json.Unmarshal(data, &ms); err != nil {
			return
		}
	}
	if ms.Setup, err = LoadSetup(); err != nil {
		return
	}

	if cm, err = LerConcurso(0); err != nil {
		return
	}
	UltimaMega = cm.Numero
	for c := ms.UltimoConcurso + 1; c <= UltimaMega; c++ {
		if cm, err = LerConcurso(c); err != nil {
			return
		}
		for _, x := range cm.DezenasSorteadasOrdemSorteio {
			n, _ := strconv.Atoi(x)
			ms.Estatistica[n] = ms.Estatistica[n] + 1
		}
	}
	ms.UltimoConcurso = UltimaMega
	if body, err = json.Marshal(ms); err != nil {
		return
	}
	if fjson, err = os.Create(MEGAJSON); err != nil {
		return
	}
	fjson.Write(body)
	fjson.Close()
	return
}

func LerConcurso(numero int) (cm *ConcursoMega, err error) {
	var (
		resp *http.Response
		body []byte
	)
	cm = new(ConcursoMega)
	endpoint := URL_MEGASENA_CAIXA_API
	if numero > 0 {
		endpoint = endpoint + "/" + strconv.Itoa(numero)
	}
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if resp, err = http.Get(endpoint); err != nil {
		return
	}
	defer resp.Body.Close()
	if body, err = io.ReadAll(resp.Body); err != nil {
		return
	}

	if err = json.Unmarshal(body, cm); err != nil {
		return
	}
	return
}
