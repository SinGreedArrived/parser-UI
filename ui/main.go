package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/net/proxy"
	"gopkg.in/yaml.v2"
)

var (
	database   source
	httpClient = &http.Client{}
	proxyFlag  = flag.String("proxy", "", "-proxy=\"127.0.0.1:9050\"")
	filename   = flag.String("file", filepath.Dir(os.Args[0])+"/save.yaml", "-file=test.yaml") // Флаг для выбора файла с целями
	//fileLog  = flag.String("log", "Stderr", "-log=parser.log")     // Флаг для логов
	threads   = flag.Int("threads", 1, "-threads=6") // Указатель кол-во потоков
	addRegexp = flag.Bool("addRegexp", false, `-addRegexp "SiteName" "RegexpForName" "RegexpForValue"`)
	addTarget = flag.Bool("addTarget", false, `-addTarget "Url" "CurrentValue"`)
	update    = flag.Bool("update", false, "-update") // Обновлять ли отслеживаемуе значения?
	maximus   = flag.Bool("max", false, "-max")       // Использовать нитей столько, сколько целей в файле json
	wg        sync.WaitGroup                          // Контроль
)

// -------------------------------------------------------------------------------
// Хранилище регулярных выражений
// -------------------------------------------------------------------------------
type regular struct {
	Exp  string
	Mask string
}

func (r *regular) Create(reg, mask string) error {
	_, err := regexp.Compile(reg)
	if err != nil {
		return errors.New(fmt.Sprintf("func:Create regula.Create\n\t%s", err))
	}
	r.Exp = reg
	r.Mask = mask
	return nil
}

// -------------------------------------------------------------------------------
// Реестр целей
// -------------------------------------------------------------------------------

type target struct {
	Url  string
	Cur  string
	data []byte
}

func (t *target) GetData() ([]byte, error) {
	if len(t.data) == 0 {
		_, err := t.GetBody()
		if err != nil {
			return nil, err
		}
	}
	return t.data, nil
}

/*
func (t *target) FindSubmatch(reg string) (string, error) {
	re := regexp.MustCompile(reg)
	result := re.FindStringSubmatch(string(t.data))
	if len(result) < 2 {
		log.Print(string(t.data))
		return "", errors.New(fmt.Sprintf("func:target.FindSubmatch target %s not found on site %s!", reg, t.Url))
	}
	return string(result[1]), nil
}
*/

func (t *target) GetBody() ([]byte, error) {
	resp, err := httpClient.Get(strings.TrimSpace(t.Url))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("func:target.GetBody http.Get\n\t%s", err))
	}
	defer resp.Body.Close()
	t.data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("func:target.GetBody ioutil.ReadAll\n\t%s", err))
	}
	re := regexp.MustCompile(`\n`)
	t.data = re.ReplaceAll(t.data, []byte(""))
	re = regexp.MustCompile(`<meta.*charset\s*=\s*"?[Uu][Tt][Ff]-8"`)
	if charset := re.FindStringSubmatch(string(t.data)); len(charset) == 0 {
		return nil, errors.New(fmt.Sprintf("func:target.GetBody encoding not UTF-8 on page %s", t.Url))
	}
	return t.GetData()
}

func (t *target) UpdateCur(s string) {
	t.Cur = s
}

// -------------------------------------------------------------------------------
// Общие хранилище
// -------------------------------------------------------------------------------

type source struct {
	Data     []*target
	Regulars map[string]*regular
}

func (s *source) Init() {
	s.Regulars = make(map[string]*regular)
}

func (s *source) CreateTarget(Url, Cur string) *target {
	for _, trg := range s.Data {
		if trg.Url == Url {
			return s.GetTarget(Url)
		}
	}
	s.Data = append(s.Data, &target{Url, Cur, nil})
	return s.GetTarget(Url)
}

func (s *source) Lenght() int {
	return len(s.Data)
}

func (s *source) CreateRegexp(url, exp, mask string) error {
	reg := new(regular)
	err := reg.Create(exp, mask)
	if err != nil {
		return err
	}
	re, _ := regexp.Compile(`http.://([^/]+)/`)
	Url := re.FindStringSubmatch(url)
	if len(Url) > 1 {
		s.Regulars[Url[1]] = reg
	} else {
		return errors.New("Not correct url address")
	}
	return nil
}

func (s *source) GetRegexp(Url string) *regular {
	re, _ := regexp.Compile(`http.://([^/]+)/`)
	url := re.FindStringSubmatch(Url)
	if len(url) > 1 {
		if reg, ok := s.Regulars[url[1]]; ok {
			return reg
		}
	}
	return nil
}

func (s *source) DeleteRegexp(url string) {
	delete(s.Regulars, url)
}

func (s *source) GetTarget(url string) *target {
	for ind, _ := range s.Data {
		if s.Data[ind].Url == url {
			return s.Data[ind]
		}
	}
	return nil
}

// -------------------------------------------------------------------------------
// Функция для потоков
// -------------------------------------------------------------------------------

func WorkerHandle(number int, e chan *target) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Worker %d crash...", number)
			go WorkerHandle(number, e) // Что-то сломалось, но это не важно, ведь лог уже записан.
		}
	}()
	for elem := range e {
		log.Printf("Start worker %d work with %s...", number, elem.Url)
		_, err := elem.GetBody()
		if err != nil {
			log.Panic(err)
		}
		reg := database.GetRegexp(elem.Url)
		if reg == nil {
			log.Panic(fmt.Sprintf("Regexp for %s not found!\n", elem.Url))
		}
		regex := regexp.MustCompile(reg.Exp)
		res := regex.ReplaceAllString(string(elem.data), reg.Mask)
		regex = regexp.MustCompile(`\d+`)
		value := regex.FindString(res)
		if elem.Cur != value {
			fmt.Printf("%s:%s\t%s\n", elem.Cur, res, elem.Url)
		} else {
			fmt.Printf("%s:%s\n", elem.Cur, res)
		}
		if *update {
			elem.UpdateCur(value)
		}
		log.Printf("Worker %d done work with %s!", number, elem.Url)
	}
	log.Printf("Worker %d done!", number)
	wg.Done()
}

// -------------------------------------------------------------------------------
// Main code
// -------------------------------------------------------------------------------

func SaveData(s *source) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf(r.(string))
		}
	}()
	file, err := os.Create(*filename)
	if err != nil {
		return errors.New(fmt.Sprintf("func:SaveData os.Create\n\t%s", err))
	}
	defer file.Close()
	yamlData, err := yaml.Marshal(s)
	if err != nil {
		return errors.New(fmt.Sprintf("func:SaveData yaml.Marshal\n\t%s", err))
	}
	file.Write(yamlData)
	log.Printf("Save data to file: " + file.Name())
	return nil
}

func LoadData(t *source) error {
	file, err := os.OpenFile(*filename, os.O_CREATE|os.O_RDONLY, 0755)
	if err != nil {
		return errors.New(fmt.Sprintf("Can't open or create filename: %s", *filename))
	}
	defer file.Close()
	yamlData, err := ioutil.ReadAll(file)
	if err != nil {
		return errors.New(fmt.Sprintf("func:LoadData ioutil.ReadAll:\n\t%s", err))
	}
	err = yaml.Unmarshal(yamlData, t)
	if err != nil {
		return errors.New(fmt.Sprintf("func:LoadData yaml.Unmarshal\n\t%s", err))
	}
	return nil
}

func ProxyInit(addr string) (err error) {
	dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
	log.Printf("Accept proxy connect")
	httpTransport := &http.Transport{}
	if err != nil {
		return err
	}
	httpTransport.Dial = dialer.Dial
	httpClient.Transport = httpTransport
	return nil
}

func main() {
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stderr)
	if *proxyFlag != "" {
		err := ProxyInit(*proxyFlag)
		if err != nil {
			log.Printf("func:main ProxyInit(%s) %s", *proxyFlag, err)
		}
	}
	database.Init()
	log.Printf("Load database from %s", *filename)
	err := LoadData(&database)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := SaveData(&database)
		if err != nil {
			log.Panic(err)
		}
		log.Printf("Done save database.")
	}()
	if *addRegexp {
		Run()
		/*		if len(flag.Args()) != 3 {
					log.Panic(`Usage: parser -addRegexp "resource name" "RegexpForName" "RegexpForValue"`)
				} else {
					database.CreateRegexp(flag.Arg(0), flag.Arg(1), flag.Arg(2))
				}
		*/
	}
	if *addTarget {
		if len(flag.Args()) != 2 {
			log.Panic(`Usage: parser -addTarget "Url" "CurrentValue"`)
		} else {
			database.CreateTarget(flag.Arg(0), flag.Arg(1))
		}
	}
	if !*addTarget && !*addRegexp {
		if *maximus {
			*threads = database.Lenght()
		}
		channelData := make(chan *target)
		log.Printf("Start %d workers for parsing...", *threads)
		for i := 1; i <= *threads; i++ {
			go WorkerHandle(i, channelData)
			wg.Add(1)
		}
		for i, _ := range database.Data {
			log.Printf("Putting %s in queue.", database.Data[i].Url)
			channelData <- database.Data[i]
		}
		close(channelData)
		wg.Wait()
	}
}
