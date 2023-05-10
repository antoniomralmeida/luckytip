package megasena

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/antoniomralmeida/luckytip/lib"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	combinations "github.com/mxschmitt/golang-combinations"
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
	Histogram      []int       `json:"histogram"`
}

type ConcursoMega struct {
	DataApuracao                 string   `json:"dataApuracao"`
	DezenasSorteadasOrdemSorteio []string `json:"dezenasSorteadasOrdemSorteio"`
	Numero                       int      `json:"numero"`
}

const (
	URL_MEGASENA_CAIXA_API = "https://servicebus2.caixa.gov.br/portaldeloterias/api/megasena"
	URL_MEGASENA_CAIXA     = "https://loterias.caixa.gov.br/paginas/mega-sena.aspx"

	MEGAJSON = "./data/megasena.json"
)

func (ms *MegaSena) BestN(n int) []int {
	return ms.Histogram[0:n]
}

type Bets struct {
	Value  float64 `json:"Value"`
	Bets   [][]int `json:"Bets"`
	Change float64 `json:"Change"`
}

func (ms *MegaSena) Aposta(valor float64) (bets Bets, js string) {
	var grauliberdade = 1
	const MAX_TENTATIVAS = 100
	bets.Value = valor
	bets.Change = valor
	if valor < ms.Setup.Modalidades[0].Valor {
		return
	}

	for i := len(ms.Setup.Modalidades) - 1; i >= 0; {
		if bets.Change >= ms.Setup.Modalidades[i].Valor {
			bets.Bets = append(bets.Bets, make([]int, ms.Setup.Modalidades[i].Numeros))
			bets.Change -= ms.Setup.Modalidades[i].Valor
		} else {
			i--
		}
	}
	fmt.Println("Gerando apostas...")
	for i, _ := range bets.Bets {
		fmt.Printf("%v%%\r", 100.0*i/len(bets.Bets))
		size := len(bets.Bets[i])
		best := ms.BestN(size + grauliberdade)
		result := combinations.Combinations(best, size)

		max := 0
	global:
		for {
			max++
			if max > MAX_TENTATIVAS {
				grauliberdade++
				best = ms.BestN(size + grauliberdade)
				result = combinations.Combinations(best, size)
				max = 0
			}
			rndgame := rand.Intn(len(result))
			bet := result[rndgame]
			for j := 0; j < i; j++ {
				if lib.Contains(bets.Bets[j], bet) {
					continue global
				}
			}
			sort.Slice(bet, func(i, j int) bool {
				return bet[i] < bet[j]
			})
			bets.Bets[i] = bet
			break
		}
	}
	fmt.Println("OK   ")
	result, _ := json.Marshal(bets)
	js = string(result)
	return
}

func LoadSetup() (setup MegaSetup, err error) {
	var (
		resp *http.Response
	)
	setup = MegaSetup{NumerosPossiveis: 60}
	fmt.Println("Lendo Setup...")
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if resp, err = http.Get(URL_MEGASENA_CAIXA); err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Println(resp)
		err = errors.New(URL_MEGASENA_CAIXA + " " + resp.Status)
		return
	}
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
	fmt.Println("OK   ")
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
	if len(ms.Setup.Modalidades) == 0 {
		if ms.Setup, err = LoadSetup(); err != nil {
			return
		}
	}

	if cm, err = LerConcurso(0); err != nil {
		return
	}
	UltimaMega = cm.Numero
	fmt.Println("Lendo jogos...")
	for c := ms.UltimoConcurso + 1; c <= UltimaMega; c++ {
		fmt.Printf("%v%%\r", 100.0*c/(UltimaMega-ms.UltimoConcurso))

		if cm, err = LerConcurso(c); err != nil {
			return
		}
		for _, x := range cm.DezenasSorteadasOrdemSorteio {
			n, _ := strconv.Atoi(x)
			ms.Estatistica[n] = ms.Estatistica[n] + c
		}
	}
	fmt.Println("OK   ")
	ms.UltimoConcurso = UltimaMega
	//Histogram
	keys := make([]int, 0, ms.Setup.NumerosPossiveis)

	for key := range ms.Estatistica {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return ms.Estatistica[keys[i]] > ms.Estatistica[keys[j]]
	})
	ms.Histogram = keys

	if body, err = json.MarshalIndent(ms, " ", ""); err != nil {
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

func (ms *MegaSena) CreateBarChart() {
	// create a new bar instance
	bar := charts.NewBar()

	// Set global options
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "MegaSena",
		Subtitle: "Tabela de frequência ponderada por tempo",
	}))

	// Put data into instance
	bar.SetXAxis(ms.generateAxisItems()).
		AddSeries("Número", ms.generateBarItems())
	f, _ := os.Create("views/bar.html")
	_ = bar.Render(f)
}

func (ms *MegaSena) generateAxisItems() []string {
	items := make([]string, 0)
	for i := 0; i < ms.Setup.NumerosPossiveis; i++ {
		n := ms.Histogram[i]
		items = append(items, strconv.Itoa(n))
	}
	return items
}

func (ms *MegaSena) generateBarItems() []opts.BarData {
	items := make([]opts.BarData, 0)
	for i := 0; i < ms.Setup.NumerosPossiveis; i++ {
		n := ms.Histogram[i]
		items = append(items, opts.BarData{Value: ms.Estatistica[n]})
	}
	return items
}
