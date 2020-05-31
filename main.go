package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
	threads   = flag.Int("threads", 4, "-threads=6") // Указатель кол-во потоков
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
	Name  string
	Value string
}

func (r *regular) Create(name, value string) {
	_, err := regexp.Compile(name)
	if err != nil {
		log.Panic(err)
	}
	_, err = regexp.Compile(value)
	if err != nil {
		log.Panic(err)
	}
	r.Name = name
	r.Value = value
}

// -------------------------------------------------------------------------------
// Реестр целей
// -------------------------------------------------------------------------------

type target struct {
	Url  string
	Cur  string
	data []byte
}

func (t *target) FindSubmatch(reg string) string {
	re := regexp.MustCompile(reg)
	result := re.FindStringSubmatch(string(t.data))
	if len(result) < 2 {
		log.Print(string(t.data))
		panic(fmt.Sprintf("func:target.FindSubmatch target %s not found on site %s!", reg, t.Url))
	}
	return string(result[1])
}

func (t *target) GetBody() {
	resp, err := httpClient.Get(t.Url)
	check("func:target.GetBody http:Get ", err)
	defer resp.Body.Close()
	t.data, err = ioutil.ReadAll(resp.Body)
	re := regexp.MustCompile(`<meta.*charset\s*=\s*"?[Uu][Tt][Ff]-8"`)
	if charset := re.FindStringSubmatch(string(t.data)); len(charset) == 0 {
		panic("func:target.GetBody encoding not UTF-8 on page " + t.Url)
	}
	check("func:target.GetBody ioutil.ReadAll ", err)
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

func (s *source) CreateTarget(Url, Cur string) {
	s.Data = append(s.Data, &target{Url, Cur, nil})
}

func (s *source) Lenght() int {
	return len(s.Data)
}

func (s *source) CreateRegexp(url, name, value string) {
	var reg regular
	reg.Create(name, value)
	s.Regulars[url] = &reg
}

func (s *source) GetRegexp(t *target) (string, string) {
	re, _ := regexp.Compile(`http.://([^/]+)/`)
	url := re.FindStringSubmatch(t.Url)
	if len(url) > 1 {
		if reg, ok := s.Regulars[url[1]]; ok {
			return reg.Name, reg.Value
		}
	}
	return "", ""
}

func (s *source) DeleteRegexp(url string) {
	delete(s.Regulars, url)
}

// -------------------------------------------------------------------------------
// Функция для потоков
// -------------------------------------------------------------------------------

func WorkerHandle(number int, e chan *target) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recover! %s", r)
			go WorkerHandle(number, e) // Что-то сломалось, но это не важно, ведь лог уже записан.
		}
	}()
	for elem := range e {
		log.Printf("Start worker %d work with %s...", number, elem.Url)
		elem.GetBody()
		Rname, Rvalue := database.GetRegexp(elem)
		if Rname == "" || Rvalue == "" {
			log.Panicf("Regexp for %s not found!\n", elem.Url)
		}
		name := elem.FindSubmatch(Rname)
		value := elem.FindSubmatch(Rvalue)
		if value != elem.Cur {
			fmt.Printf("%s:%s\t| %s\t%s\n", elem.Cur, value, name, elem.Url)
		} else {
			fmt.Printf("%s:%s\t| %s\n", elem.Cur, value, name)
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

func check(msg string, err error) {
	if err != nil {
		log.Panic("PANIC: " + msg + err.Error())
	}
}

func SaveData(s *source) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf(r.(string))
		}
	}()
	file, err := os.Create(*filename)
	check("func:SaveData os.Create", err)
	defer file.Close()
	yamlData, err := yaml.Marshal(s)
	check("func:SaveData yaml.Marshal", err)
	file.Write(yamlData)
	log.Printf("Save data to file: " + file.Name())
}

func LoadData(t *source) {
	file, err := os.Open(*filename)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()
	yamlData, err := ioutil.ReadAll(file)
	check("func:LoadData ioutil.ReadAll ", err)
	err = yaml.Unmarshal(yamlData, t)
	if err != nil {
		log.Fatal(err.Error())
	}
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
	LoadData(&database)
	defer func() {
		SaveData(&database)
		log.Printf("Done save database.")
	}()
	if *addRegexp {
		if len(flag.Args()) != 3 {
			log.Panic(`Usage: parser -addRegexp "resource name" "RegexpForName" "RegexpForValue"`)
		} else {
			database.CreateRegexp(flag.Arg(0), flag.Arg(1), flag.Arg(2))
		}
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
