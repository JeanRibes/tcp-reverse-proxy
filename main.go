package main

import (
	"bufio"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var (
	connectedServers = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tcp_reverse_proxy_connected_servers",
		Help: "Current number of servers that connected to proxy",
	})
)
var (
	connectedClients = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tcp_reverse_proxy_connected_clients",
		Help: "Current number of clients that connected to proxy",
	})
)
var (
	scrapedWebsitesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "security_crawler_scraped_websites_total",
		Help: "Total number of websites scraped by server",
	})
)
var scrapedWebsites = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "security_crawler_scraped_websites",
		Help: "Number of scrapes by website",
	}, []string{"website", "ratio"},
)
var ratios = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "security_crawler_scraped_http_https_ratio",
	Help: "http to https ratio of scraped websites",
})

var ratio = 0

/*
ports : 1998 -> port d'écoute du reverse proxy pour la destination host
1999: port d'écoute pour la source client
*/
var maplock sync.Mutex

type host struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func newHost(conn net.Conn, reader *bufio.Reader, writer *bufio.Writer) host {
	return host{
		conn:   conn,
		reader: reader,
		writer: writer,
	}
}

var routingMap = make(map[string]*host)

func main() {
	ratios.Set(0.0)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/report", func(writer http.ResponseWriter, request *http.Request) {
		println("report strating")
		scrapedWebsitesTotal.Inc()
		key, ok := request.URL.Query()["ratio"]
		lien, lok := request.URL.Query()["website"]
		if ok {
			fmt.Println("key:" + key[0])
			ra, _ := strconv.Atoi(key[0])
			if ratio == 0 {
				ratio = ra
			} else {
				ratio = (ratio + ra) / 2
			}
			ratios.Set(float64(ratio))
			if lok {
				println(lien[0])
				scrapedWebsites.WithLabelValues(lien[0], key[0]).Inc()
			}
		} else {
			println("non ok " + strconv.Itoa(len(key)))
		}

		fmt.Fprint(writer, "merci")
	})
	go http.ListenAndServe(":2112", nil)

	listener, _ := net.Listen("tcp", ":2323")
	println("listening")
	for {
		conn, _ := listener.Accept()
		go dispatch(conn)
	}
}
func dispatch(conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	identifier, _ := reader.ReadString('\n')
	if len(identifier) > 2 {
		channel := strings.TrimSuffix(identifier[1:], "\n")
		fmt.Println("channel " + channel)

		if identifier[0] == 's' {
			println("server")
			go handleServer(channel, conn, reader, writer)
		} else {
			if identifier[0] == 'c' {
				println("client")
				go handleClient(channel, conn, reader, writer)
			} else {
				println("incorrect host type, closing")
				writer.WriteString("message incorrect: le 1er message envoyé doit être de la forme 'sname' (host) ou 'cname' (client)\n")
				conn.Close()
			}
		}
	} else {
		writer.WriteString("il faut un identifiant de canal\n")
		writer.Flush()
	}
}
func handleServer(channel string, conn net.Conn, reader *bufio.Reader, writer *bufio.Writer) {

	maplock.Lock()
	routingMap[channel] = &host{
		conn:   conn,
		reader: reader,
		writer: writer,
	}
	maplock.Unlock()
	writer.WriteString("registred as server in channel " + channel + "\n")
	writer.Flush()
}

func handleClient(channel string, conn net.Conn, reader *bufio.Reader, writer *bufio.Writer) {
	maplock.Lock()
	server := routingMap[channel]
	if server != nil {
		routingMap[channel] = nil
		maplock.Unlock()
		client := newHost(conn, reader, writer)
		writer.WriteString("registred as client in channel " + channel + "\n")
		writer.Flush()
		println("starting copy")
		go copy(client, *server)
		go copy(*server, client)
		println("started")

	} else {
		maplock.Unlock()
		writer.WriteString("no servers found in channel " + channel + "\n")
		writer.Flush()
		println("client tried to join empty channel, closing")
		conn.Close()
	}
}
func copy(sender host, receiver host) {
	up := true
	for up {
		str, rerr := sender.reader.ReadString('\n')
		for _, char := range str {
			fmt.Printf("%q", char)
			print(" ")
		}
		if len(str) > 3 {
			println(str[0:3])
			if str[0:3] == "EOF" {
				receiver.writer.WriteString("EOF\n")
				goodbye(sender, receiver)
				break
			}
		}
		if rerr != nil {
			up = false
			println(rerr.Error())
		} else {
			_, werr := receiver.writer.WriteString(str)
			if werr != nil {
				up = false
				println(werr.Error())
			}
			ferr := receiver.writer.Flush()
			if ferr != nil {
				up = false
				println(ferr.Error())
			}
		}
	}
	write(sender.writer, "closing because of an error on one side")
	write(receiver.writer, "closing because of an error on one side")
	sender.conn.Close()
	receiver.conn.Close()
	println("copy stopped")
}

func goodbye(sender host, receiver host) {
	write(sender.writer, "okay, goodbye")
	write(receiver.writer, "other side disconnected, goodbye")
	sender.conn.Close()
	receiver.conn.Close()
}

func write(writer *bufio.Writer, str string) (error, error) {
	_, err1 := writer.WriteString(str)
	err2 := writer.Flush()
	return err1, err2
}
