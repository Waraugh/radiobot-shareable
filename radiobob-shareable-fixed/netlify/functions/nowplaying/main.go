package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const streamURL = "https://regiocast.streamabc.net/regc-radiobob2000rock2507507-mp3-192-9881528"

type result struct {
	Title     string `json:"title"`
	Raw       string `json:"raw,omitempty"`
	Station   string `json:"station"`
	UpdatedAt string `json:"updatedAt"`
}

func readTitle() (string, string, error) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", "regiocast.streamabc.net:443", &tls.Config{ServerName: "regiocast.streamabc.net"})
	if err != nil {
		return "", "", err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))

	request := "GET /regc-radiobob2000rock2507507-mp3-192-9881528 HTTP/1.1\r\n" +
		"Host: regiocast.streamabc.net\r\n" +
		"User-Agent: RADIO-BOB-Web-Player/1.0\r\n" +
		"Accept: */*\r\n" +
		"Icy-MetaData: 1\r\n" +
		"Connection: close\r\n\r\n"
	if _, err := io.WriteString(conn, request); err != nil {
		return "", "", err
	}

	br := bufio.NewReaderSize(conn, 32768)
	status, err := br.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	if !strings.Contains(status, " 200 ") {
		return "", "", fmt.Errorf("stream returned %s", strings.TrimSpace(status))
	}

	metaint := 0
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return "", "", err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if i := strings.IndexByte(line, ':'); i > 0 {
			name := strings.ToLower(strings.TrimSpace(line[:i]))
			value := strings.TrimSpace(line[i+1:])
			if name == "icy-metaint" {
				metaint, _ = strconv.Atoi(value)
			}
		}
	}
	if metaint <= 0 {
		return "", "", fmt.Errorf("stream did not provide icy-metaint")
	}

	audio := make([]byte, metaint)
	for i := 0; i < 12; i++ {
		if _, err := io.ReadFull(br, audio); err != nil {
			return "", "", err
		}
		size, err := br.ReadByte()
		if err != nil {
			return "", "", err
		}
		metaLen := int(size) * 16
		if metaLen == 0 {
			continue
		}
		meta := make([]byte, metaLen)
		if _, err := io.ReadFull(br, meta); err != nil {
			return "", "", err
		}
		raw := strings.TrimRight(string(meta), "\x00")
		const key = "StreamTitle='"
		begin := strings.Index(raw, key)
		if begin < 0 {
			continue
		}
		begin += len(key)
		finish := strings.Index(raw[begin:], "';")
		if finish < 0 {
			continue
		}
		title := strings.TrimSpace(raw[begin : begin+finish])
		if title != "" {
			return title, raw, nil
		}
	}
	return "", "", fmt.Errorf("no StreamTitle received")
}

func handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	title, raw, err := readTitle()
	status := 200
	payload := map[string]interface{}{}
	if err != nil {
		status = 502
		payload["error"] = err.Error()
	} else {
		payload = map[string]interface{}{
			"title":     title,
			"raw":       raw,
			"station":   "RADIO BOB! 2000er Rock",
			"updatedAt": time.Now().UTC().Format(time.RFC3339),
		}
	}
	body, _ := json.Marshal(payload)
	return &events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Content-Type":                "application/json; charset=utf-8",
			"Cache-Control":               "no-store, max-age=0",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func main() { lambda.Start(handler) }
