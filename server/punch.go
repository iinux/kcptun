package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func punchHole() string {
	laddr := &net.UDPAddr{Port: 8021}

	ur := func(address string, receive bool) string {
		raddr, err := net.ResolveUDPAddr("udp", address)
		checkError(err)

		conn, err := net.DialUDP("udp", laddr, raddr)
		checkError(err)

		_, err = conn.Write([]byte("hi"))
		checkError(err)

		var data string
		if receive {
			dataBytes := make([]byte, 32)
			_, err = conn.Read(dataBytes)
			checkError(err)

			data = string(dataBytes)
		}

		err = conn.Close()
		checkError(err)

		return data
	}

	d1 := ur("hw.iinux.cn:8021", true)
	fmt.Println(d1)
	d1s := strings.Split(d1, " ")

	d2 := ur("rack.iinux.cn:8021", true)
	fmt.Println(d2)
	d2s := strings.Split(d2, " ")

	if d1s[1] == d2s[1] {
		fmt.Println("check success", d1s[1], d2s[1])
	} else {
		fmt.Println("check fail")
		os.Exit(1)
	}

	self := strings.Trim(d1s[1], "\x00")

	client := &http.Client{}
	req, err := http.NewRequest("GET",
		"https://m.iinux.cn/api/switch/ph/"+self, nil)
	checkError(err)

	req.Host = "m.iinux.cn"
	req.Header.Add("Content-Type", `application/json`)

	var partner string

	for partner == "" {
		resp, err := client.Do(req)
		respData, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(respData))

		respObj := switchResp{}
		err = json.Unmarshal(respData, &respObj)
		checkError(err)

		for _, datum := range respObj.Data {
			if datum == self {
				continue
			} else {
				partner = datum
			}
		}

		if partner != "" {
			break
		} else {
			time.Sleep(5 * time.Second)
		}
	}

	fmt.Println("partner is ", partner)
	ur(partner, false)

	return partner
}

type switchResp struct {
	Code int      `json:"code"`
	Data []string `json:"data"`
}
