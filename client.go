package sks3200m8g0y1xf

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	ErrLoginFailed = errors.New("login failed")
	ErrNotLogin    = errors.New("not login")
)

type Client struct {
	url      string
	username string
	password string
	response string
}

func NewClient(url string) *Client {
	return &Client{url: url}
}

func (c *Client) Login(username, password string) error {
	// /login.cgi
	c.username = username
	c.password = password

	// usernameとpasswordの連結文字列からMD5 Hashを計算
	hash := md5.New()
	hash.Write([]byte(username))
	hash.Write([]byte(password))
	response := fmt.Sprintf("%x", hash.Sum(nil))

	formData := url.Values{
		"username": {username},
		"password": {password},
		"language": {"EN"},
		"Response": {response},
	}

	// POSTリクエストを送信
	if _, err := c.post("login.cgi", formData); err != nil {
		return ErrLoginFailed
	}
	c.response = response
	return nil
}

func (c *Client) Logout() error {
	// /logout.cgi
	if _, err := c.get("logout.cgi"); err != nil {
		return err
	}
	c.response = ""
	return nil
}

type PortStatic struct {
	PortNumber int    `json:"port_number"`
	State      string `json:"state"`
	LinkStatus string `json:"link_status"`
	TxGoodPkt  uint64 `json:"tx_good_pkt"`
	TxBadPkt   uint64 `json:"tx_bad_pkt"`
	RxGoodPkt  uint64 `json:"rx_good_pkt"`
	RxBadPkt   uint64 `json:"rx_bad_pkt"`
}

func (c *Client) GetMonitoringPortStatics() ([]PortStatic, error) {
	ports := []PortStatic{}
	// /port.cgi?page=stats
	body, err := c.get("port.cgi?page=stats")
	if err != nil {
		return nil, err
	}

	// goqueryでHTMLをパース
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	// テーブルの行を取得
	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		// テーブルの列を取得
		port := PortStatic{}
		s.Find("td").Each(func(j int, s *goquery.Selection) {
			switch j {
			case 0:
				// Port Number "Port N"をパース
				s := strings.Split(s.Text(), " ")
				if len(s) != 2 {
					return
				}
				p, err := toInt(s[1])
				if err != nil {
					return
				}
				port.PortNumber = p
			case 1:
				// State
				port.State = strings.TrimSpace(s.Text())
			case 2:
				// Link Status
				port.LinkStatus = strings.TrimSpace(s.Text())
			case 3:
				// Tx Good Pkt
				p, err := toInt64(s.Text())
				if err != nil {
					return
				}
				port.TxGoodPkt = uint64(p)
			case 4:
				// Tx Bad Pkt
				p, err := toInt64(s.Text())
				if err != nil {
					return
				}
				port.TxBadPkt = uint64(p)
			case 5:
				// Rx Good Pkt
				p, err := toInt64(s.Text())
				if err != nil {
					return
				}
				port.RxGoodPkt = uint64(p)
			case 6:
				// Rx Bad Pkt
				p, err := toInt64(s.Text())
				if err != nil {
					return
				}
				port.RxBadPkt = uint64(p)
			}
		})
		if port.State != "" && port.LinkStatus != "" {
			ports = append(ports, port)
		}
	})
	return ports, nil
}

func toInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

func toInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func (c *Client) post(p string, formData url.Values) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", c.url+path.Join("/", p), strings.NewReader(formData.Encode()))
	if err != nil {
		return "", err
	}
	// responseがない場合はログインしていないのでCookieを追加しない
	if c.response != "" {
		cookie := &http.Cookie{
			Name:  c.username,
			Value: c.response,
		}
		req.AddCookie(cookie)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	strBody := string(body)
	if strings.Contains(strBody, "/login.cgi") {
		return "", ErrNotLogin
	}

	return strBody, nil
}

func (c *Client) get(p string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", c.url+path.Join("/", p), nil)
	if err != nil {
		return "", err
	}
	cookie := &http.Cookie{
		Name:  c.username,
		Value: c.response,
	}
	req.AddCookie(cookie)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	strBody := string(body)

	if strings.Contains(strBody, "/login.cgi") {
		return "", ErrNotLogin
	}

	return strBody, nil
}
