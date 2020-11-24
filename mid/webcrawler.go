package main
import (
	"fmt"
	"net/http"
	"github.com/jackdanger/collectlinks"
	"net/url"
)

var visited = make(map[string]bool)

func main(){
	//url := "http://example.com/"
	url := "https://www.google.com/"
	
	queue := make(chan string)
	go func(){
		queue <- url
	}()
	for uri := range queue{
		download(uri,queue)
	}
}

func download(url string,queue chan string){
	visited[url] = true
	client := &http.Client{}
	req,_ := http.NewRequest("GET",url,nil)
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := client.Do(req)
	if err != nil{
		fmt.Println("http get error",err)
		return
	}
    //stop
	if len(visited)>2{
		resp.Body.Close()
	}

	links := collectlinks.All(resp.Body)
	for _,link := range links{
		absolute := urlJoin(link,url)
		if url != " "{
			if !visited[absolute]{
				fmt.Println("parse url : ",absolute)
				go func() {
					queue <- absolute
				}()
			}
			
		}
	}
}

func urlJoin(herf,base string) string{
	uri, err := url.Parse(herf)
	if err != nil{
		return " "
	} 
	baseUrl, err := url.Parse(base)
	if err != nil{
		return " "
	} 
	return baseUrl.ResolveReference(uri).String()
}