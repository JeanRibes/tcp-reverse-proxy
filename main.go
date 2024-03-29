package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"net"
	"net/http"
	"strconv"
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
ports : 1998 -> port d'écoute du reverse proxy pour la destination serveur
1999: port d'écoute pour la source client
*/
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
	println("listening")
	client, _ := net.Listen("tcp", ":1999")
	serveur, _ := net.Listen("tcp", ":1998") //serveur
	for {
		serveurConn, _ := serveur.Accept()
		connectedServers.Inc()
		fmt.Println("server connected " + serveurConn.RemoteAddr().String())

		clientConn, _ := client.Accept()
		connectedClients.Inc()
		fmt.Print("client connected ... ")
		_, werr := io.Writer(serveurConn).Write([]byte(clientConn.RemoteAddr().String() + "\x02")) //signale de gérer le client
		if werr == nil {
			//démarrage de la fonction de proxy
			fmt.Println("starting proxy for client " + clientConn.RemoteAddr().String())
			go func() {
				_, err := io.Copy(serveurConn, clientConn)
				if err != nil {
					fmt.Println(err)
					println("fermeture des connexions")
					serveurConn.Close()
					clientConn.Close()
				}
			}()
			go func() {
				_, err := io.Copy(clientConn, serveurConn)
				if err != nil {
					fmt.Println(err)
					println("fermeture des connexions")
					serveurConn.Close()
					clientConn.Close()
				}
			}()
		} else {
			fmt.Println(werr)
		}
		println("fin boucle")
	}

}
