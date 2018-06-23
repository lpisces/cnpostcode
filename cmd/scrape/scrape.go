package scrape

import (
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

const (
	MaxPostcode = 999999
	BaseUrl     = "http://v.juhe.cn/postcode/"
	Retry       = 5
)

func Run(c *cli.Context) (err error) {

	spiderNumber := c.Int("number")
	if spiderNumber <= 0 {
		spiderNumber = 1
	}

	key := c.String("key")
	if len(key) == 0 {
		return fmt.Errorf("invalid key")
	}

	cache := c.String("cache")
	if c.Bool("debug") {
		log.Println(cache)
	}
	if _, err := os.Stat(cache); os.IsNotExist(err) {
		if e := os.Mkdir(cache, os.ModeDir); e != nil {
			return e
		}
	}

	output := c.String("output")
	if c.Bool("debug") {
		log.Println(output)
	}
	if _, err := os.Stat(output); os.IsNotExist(err) {
		if e := os.Mkdir(output, os.ModeDir); e != nil {
			return e
		}
	}
	dataFile := output + "/" + "postcode.dat"

	endpoint := "query"

	var Failed []string
	var wg sync.WaitGroup
	limitChan := make(chan int, spiderNumber)
	lines := make(chan string, MaxPostcode)
	mu := sync.Mutex{}

	for i := 0; i < MaxPostcode; i++ {
		wg.Add(1)
		limitChan <- i
		if c.Bool("debug") {
			log.Printf("%d pushed", i)
		}

		go func() {
			defer func() {
				wg.Done()
				i := <-limitChan
				log.Printf("%d poped", i)
			}()

			retry := Retry
			code := fmt.Sprintf("%06d", i)
			cachePath := cache + "/" + code
			var body []byte
			if _, err := os.Stat(cachePath); os.IsNotExist(err) {
			R:
				//log.Println(i)
				u, _ := url.Parse(fmt.Sprintf("%s%s", BaseUrl, endpoint))
				param := url.Values{}

				//配置请求参数,方法内部已处理urlencode问题,中文参数可以直接传参
				//param.Set("postcode","") //邮编，如：215001
				param.Set("key", key)       //应用APPKEY(应用详细页查询)
				param.Set("page", "1")      //页数，默认1
				param.Set("pagesize", "50") //每页返回，默认:20,最大不超过50
				param.Set("dtype", "json")  //返回数据的格式,xml或json，默认json
				param.Set("postcode", code)
				u.RawQuery = param.Encode()

				client := &http.Client{}
				req, _ := http.NewRequest("GET", u.String(), nil)

				resp, err := client.Do(req)
				if err != nil {
					retry--
					if retry > 0 {
						goto R
					}
					Failed = append(Failed, code)
					if c.Bool("debug") {
						log.Println(err)
					}
					return
				}

				// status code check
				if resp.StatusCode != http.StatusOK {
					retry--
					if retry > 0 {
						goto R
					}
					Failed = append(Failed, code)
					if c.Bool("debug") {
						log.Printf("http code: %d", resp.StatusCode)
					}
					return
				}

				body, _ = ioutil.ReadAll(resp.Body)

				// error code check
				errorCode := gjson.GetBytes(body, "error_code")
				if errorCode.Int() != 0 {
					retry--
					if retry > 0 {
						goto R
					}
					Failed = append(Failed, code)
					if c.Bool("debug") {
						log.Printf("api error code: %d", errorCode.Int())
					}
					return
				}

				// write cache
				if _, err := os.Stat(cachePath); os.IsNotExist(err) {
					ioutil.WriteFile(cachePath, body, 0644)
				}
			} else {
				body, _ = ioutil.ReadFile(cachePath)
			}

			cnt := gjson.GetBytes(body, "result.list.#")
			if cnt.Int() == 0 {
				return
			}
			province := gjson.GetBytes(body, "result.list.#.Province")
			city := gjson.GetBytes(body, "result.list.#.City")
			district := gjson.GetBytes(body, "result.list.#.District")

			var pStr, cStr, dStr string
			for _, v := range province.Array() {
				if v.String() != "" {
					pStr = v.String()
					break
				}
			}

			for _, v := range city.Array() {
				if v.String() != "" {
					cStr = v.String()
					break
				}
			}

			for _, v := range district.Array() {
				if v.String() != "" {
					dStr = v.String()
					break
				}
			}

			line := fmt.Sprintf("%s,%s,%s,%s\n", code, pStr, cStr, dStr)
			mu.Lock()
			lines <- line
			mu.Unlock()

			if c.Bool("debug") {
				log.Print(line)
			}
		}()
	}
	wg.Wait()
	close(lines)

	//ioutil.WriteFile(dataFile, []byte(line), 0644)
	f, err := os.OpenFile(dataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	for line := range lines{
		if _, err = f.WriteString(line); err != nil {
			log.Println(err)
		}
	}

	if c.Bool("debug") {
		log.Printf("get %d errors, postcode is %v", len(Failed), Failed)
	}

	return
}
