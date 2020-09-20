package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	cli "github.com/urfave/cli/v2"
)

var (
	baseApi = "http://demo.notr.tech/api-service/v1"
)

func main() {
	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:  "login",
				Usage: "login to cframe, access token will store in token.conf",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "username",
						Usage: "login username",
						Value: "cframe",
					},
					&cli.StringFlag{
						Name:  "password",
						Usage: "login password",
						Value: "123123qwe",
					},
				},
				Action: func(c *cli.Context) error {
					username := c.String("username")
					password := c.String("password")
					Login(username, password)
					return nil
				},
			},
			{
				Name:  "edge",
				Usage: "edge control(add,del,list,topology)",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add edage node",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Usage:    "edge name",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "hostaddr",
								Usage:    "edge host address(public_ip:port)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "cidr",
								Usage:    "vpc cidr",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							name := c.String("name")
							hostAddr := c.String("hostaddr")
							cidr := c.String("cidr")
							AddEdge(name, hostAddr, cidr)
							return nil
						},
					},
					{
						Name:  "del",
						Usage: "delete edage node",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "name",
								Usage:    "edge name",
								Required: true,
							},
						},
						Action: func(c *cli.Context) error {
							edgeName := c.String("name")
							DelEdge(edgeName)
							return nil
						},
					},
					{
						Name:  "list",
						Usage: "add edage node",
						Action: func(c *cli.Context) error {
							GetEdges()
							return nil
						},
					},
					{
						Name:  "topology",
						Usage: "get topology info",
						Action: func(c *cli.Context) error {
							GetTopology()
							return nil
						},
					},
				},
			},
		},
		Name:  "cfctrl",
		Usage: "cfctrl is cframe api server cli.",
	}
	app.Run(os.Args)
}

type replyBody struct {
	Code    int
	Message string
	Data    interface{}
}

type loginReply struct {
	Token    string
	Username string
}

type loginForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(username, password string) {
	url := fmt.Sprintf("%s/user/signin", baseApi)

	f := &loginForm{
		Username: username,
		Password: password,
	}
	cnt, _ := req(url, "", f)
	fmt.Println("login reply: ", string(cnt))
	saveToken(string(cnt))
}

func GetEdges() {
	token := readToken()
	if len(token) <= 0 {
		fmt.Println("require login first")
		return
	}

	cli := http.Client{
		Timeout: time.Second * 10,
	}

	url := fmt.Sprintf("%s/edge/list", baseApi)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Access-Token", token)

	resp, err := cli.Do(req)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		return
	}
	defer resp.Body.Close()

	cnt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		return
	}
	fmt.Println(string(cnt))
}

func GetTopology() {
	token := readToken()
	if len(token) <= 0 {
		fmt.Println("require login first")
		return
	}

	cli := http.Client{
		Timeout: time.Second * 10,
	}

	url := fmt.Sprintf("%s/edge/topology", baseApi)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Access-Token", token)

	resp, err := cli.Do(req)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		return
	}
	defer resp.Body.Close()

	cnt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		return
	}
	fmt.Println(string(cnt))
}

type addEdgeForm struct {
	Name     string `json:"name"`
	HostAddr string `json:"hostaddr"`
	Cidr     string `json:"cidr"`
}

func AddEdge(name, hostaddr, cidr string) {
	token := readToken()
	if len(token) <= 0 {
		fmt.Println("require login first")
		return
	}

	url := fmt.Sprintf("%s/edge/add", baseApi)
	f := addEdgeForm{
		Name:     name,
		HostAddr: hostaddr,
		Cidr:     cidr,
	}
	cnt, _ := req(url, token, f)
	fmt.Println(string(cnt))
}

type delEdgeForm struct {
	Name string `json:"name"`
}

func DelEdge(name string) {
	token := readToken()
	if len(token) <= 0 {
		fmt.Println("require login first")
		return
	}

	url := fmt.Sprintf("%s/edge/del", baseApi)
	f := delEdgeForm{
		Name: name,
	}
	cnt, _ := req(url, token, f)
	fmt.Println(string(cnt))
}

func saveToken(token string) error {
	fp, err := os.Create("token.conf")
	if err != nil {
		return err
	}
	defer fp.Close()
	fp.Write([]byte(token))
	fmt.Println("save loginReply to token.conf")
	return nil
}

func readToken() string {
	cnt, err := ioutil.ReadFile("token.conf")
	if err != nil {
		fmt.Println("load token from token.conf fail, please login")
		os.Exit(0)
	}

	body := replyBody{}
	json.Unmarshal(cnt, &body)
	bd, _ := json.Marshal(body.Data)
	tokenData := loginReply{}
	json.Unmarshal(bd, &tokenData)
	return tokenData.Token
}

func req(url, token string, f interface{}) (string, error) {
	client := http.Client{
		Timeout: time.Second * 10,
	}

	body, _ := json.Marshal(f)

	br := bytes.NewReader(body)
	request, _ := http.NewRequest("POST", url, br)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Access-Token", token)

	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	cnt, err := ioutil.ReadAll(resp.Body)
	return string(cnt), err
}
